package store

// NoopStore is a Store that always returns empty results.
// It is used when DB_ENABLED=false so the rest of the code can rely on the
// Store interface without nil checks.
type NoopStore struct{}

func (NoopStore) Ping() error                       { return nil }
func (NoopStore) ListAccounts() ([]AccountRow, error)    { return nil, nil }
func (NoopStore) GetAccount(string) (*AccountRow, error)  { return nil, nil }
func (NoopStore) UpsertAccount(AccountRow) error          { return nil }
func (NoopStore) DeleteAccount(string) error              { return nil }
func (NoopStore) TouchLastStartedAt(string) error         { return nil }
func (NoopStore) Changes() <-chan struct{}                 { return nil }
func (NoopStore) Close() error                            { return nil }
