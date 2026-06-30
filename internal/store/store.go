// Package store defines the persistence interface for account configurations.
package store

import "time"

// AccountRow represents a single row in the accounts table.
type AccountRow struct {
	Username   string
	ConfigJSON string
	Enabled    bool
	UpdatedAt  time.Time
}

// Store is the persistence interface for account configurations.
type Store interface {
	// ListAccounts returns all account rows.
	ListAccounts() ([]AccountRow, error)

	// UpsertAccount inserts or updates an account row, setting updated_at to now.
	UpsertAccount(row AccountRow) error

	// DeleteAccount removes an account row by username.
	DeleteAccount(username string) error

	// Changes returns a channel that receives a signal whenever the accounts
	// table is modified. Implementations that do not support push notifications
	// may leave this channel nil — callers must handle that case.
	Changes() <-chan struct{}

	// Close releases resources held by the store.
	Close() error
}
