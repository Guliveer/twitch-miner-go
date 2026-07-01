package store

import (
	"database/sql"
	"embed"
	"fmt"
	"sync"
	"time"

	"github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var migrations embed.FS

// PostgresStore is a Store backed by a PostgreSQL database.
type PostgresStore struct {
	db      *sql.DB
	changes chan struct{}
	once    sync.Once
	done    chan struct{}
}

// OpenPostgres opens a connection to the given DSN and runs all pending migrations.
func OpenPostgres(dsn string) (*PostgresStore, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("opening postgres: %w", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("connecting to postgres: %w", err)
	}

	goose.SetBaseFS(migrations)
	goose.SetLogger(goose.NopLogger())
	if err := goose.SetDialect("postgres"); err != nil {
		db.Close()
		return nil, fmt.Errorf("setting goose dialect: %w", err)
	}
	if err := goose.Up(db, "migrations"); err != nil {
		db.Close()
		return nil, fmt.Errorf("running migrations: %w", err)
	}

	s := &PostgresStore{
		db:      db,
		changes: make(chan struct{}, 1),
		done:    make(chan struct{}),
	}
	go s.listen(dsn)
	return s, nil
}

// listen maintains a LISTEN connection via pq.Listener and forwards notifications
// to s.changes. Reconnects automatically on disconnect.
func (s *PostgresStore) listen(dsn string) {
	listener := pq.NewListener(dsn, 5*time.Second, time.Minute, nil)
	defer listener.Close()

	if err := listener.Listen("accounts_changed"); err != nil {
		// Non-fatal: fall back to polling only.
		return
	}

	for {
		select {
		case <-s.done:
			return
		case _, ok := <-listener.Notify:
			if !ok {
				return
			}
			s.notify()
		}
	}
}

func (s *PostgresStore) notify() {
	select {
	case s.changes <- struct{}{}:
	default:
	}
}

func (s *PostgresStore) ListAccounts() ([]AccountRow, error) {
	rows, err := s.db.Query(`SELECT username, config_json, enabled, updated_at, last_started_at FROM accounts`)
	if err != nil {
		return nil, fmt.Errorf("listing accounts: %w", err)
	}
	defer rows.Close()

	var accounts []AccountRow
	for rows.Next() {
		r, err := scanRow(rows.Scan)
		if err != nil {
			return nil, fmt.Errorf("scanning account row: %w", err)
		}
		accounts = append(accounts, r)
	}
	return accounts, rows.Err()
}

func (s *PostgresStore) GetAccount(username string) (*AccountRow, error) {
	row := s.db.QueryRow(
		`SELECT username, config_json, enabled, updated_at, last_started_at FROM accounts WHERE username = $1`,
		username,
	)
	r, err := scanRow(row.Scan)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting account %s: %w", username, err)
	}
	return &r, nil
}

func (s *PostgresStore) TouchLastStartedAt(username string) error {
	_, err := s.db.Exec(
		`UPDATE accounts SET last_started_at = $1 WHERE username = $2`,
		time.Now().UnixMilli(), username,
	)
	if err != nil {
		return fmt.Errorf("touching last_started_at for %s: %w", username, err)
	}
	return nil
}

// scanRow scans the 5-column account row using the provided Scan function.
// Timestamps are stored as Unix milliseconds.
func scanRow(scan func(...any) error) (AccountRow, error) {
	var r AccountRow
	var updatedAtMs int64
	var lastStartedAtMs sql.NullInt64
	if err := scan(&r.Username, &r.ConfigJSON, &r.Enabled, &updatedAtMs, &lastStartedAtMs); err != nil {
		return AccountRow{}, err
	}
	r.UpdatedAt = time.UnixMilli(updatedAtMs)
	if lastStartedAtMs.Valid {
		t := time.UnixMilli(lastStartedAtMs.Int64)
		r.LastStartedAt = &t
	}
	return r, nil
}

func (s *PostgresStore) UpsertAccount(row AccountRow) error {
	_, err := s.db.Exec(
		`INSERT INTO accounts (username, config_json, enabled, updated_at)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (username) DO UPDATE
		   SET config_json = EXCLUDED.config_json,
		       enabled     = EXCLUDED.enabled,
		       updated_at  = EXCLUDED.updated_at`,
		row.Username, row.ConfigJSON, row.Enabled, time.Now().UnixMilli(),
	)
	if err != nil {
		return fmt.Errorf("upserting account %s: %w", row.Username, err)
	}
	s.notify()
	return nil
}

func (s *PostgresStore) DeleteAccount(username string) error {
	_, err := s.db.Exec(`DELETE FROM accounts WHERE username = $1`, username)
	if err != nil {
		return fmt.Errorf("deleting account %s: %w", username, err)
	}
	s.notify()
	return nil
}

func (s *PostgresStore) Changes() <-chan struct{} {
	return s.changes
}

func (s *PostgresStore) Close() error {
	s.once.Do(func() { close(s.done) })
	return s.db.Close()
}
