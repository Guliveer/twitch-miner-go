package store

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/lib/pq"
)

const schema = `
CREATE TABLE IF NOT EXISTS accounts (
	username    TEXT PRIMARY KEY,
	config_json TEXT NOT NULL,
	enabled     BOOLEAN NOT NULL DEFAULT TRUE,
	updated_at  BIGINT NOT NULL
);

-- Notify on any change so listeners wake up immediately.
CREATE OR REPLACE FUNCTION accounts_notify() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN PERFORM pg_notify('accounts_changed', ''); RETURN NULL; END;
$$;

DROP TRIGGER IF EXISTS accounts_changed_trigger ON accounts;
CREATE TRIGGER accounts_changed_trigger
	AFTER INSERT OR UPDATE OR DELETE ON accounts
	FOR EACH STATEMENT EXECUTE FUNCTION accounts_notify();
`

// PostgresStore is a Store backed by a PostgreSQL database.
type PostgresStore struct {
	db      *sql.DB
	changes chan struct{}
	once    sync.Once
	done    chan struct{}
}

// OpenPostgres opens a connection to the given DSN and runs the schema migration.
func OpenPostgres(dsn string) (*PostgresStore, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("opening postgres: %w", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("connecting to postgres: %w", err)
	}

	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("running schema migration: %w", err)
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
	rows, err := s.db.Query(`SELECT username, config_json, enabled, updated_at FROM accounts`)
	if err != nil {
		return nil, fmt.Errorf("listing accounts: %w", err)
	}
	defer rows.Close()

	var accounts []AccountRow
	for rows.Next() {
		var r AccountRow
		var updatedAtUnix int64
		if err := rows.Scan(&r.Username, &r.ConfigJSON, &r.Enabled, &updatedAtUnix); err != nil {
			return nil, fmt.Errorf("scanning account row: %w", err)
		}
		r.UpdatedAt = time.Unix(updatedAtUnix, 0)
		accounts = append(accounts, r)
	}
	return accounts, rows.Err()
}

func (s *PostgresStore) UpsertAccount(row AccountRow) error {
	_, err := s.db.Exec(
		`INSERT INTO accounts (username, config_json, enabled, updated_at)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (username) DO UPDATE
		   SET config_json = EXCLUDED.config_json,
		       enabled     = EXCLUDED.enabled,
		       updated_at  = EXCLUDED.updated_at`,
		row.Username, row.ConfigJSON, row.Enabled, time.Now().Unix(),
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
