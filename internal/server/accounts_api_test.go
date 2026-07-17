package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Guliveer/twitch-miner-go/internal/store"
)

// fakeStore is an in-memory Store used in tests.
type fakeStore struct {
	rows    map[string]store.AccountRow
	changes chan struct{}
}

func newFakeStore() *fakeStore {
	return &fakeStore{
		rows:    make(map[string]store.AccountRow),
		changes: make(chan struct{}, 1),
	}
}

func (f *fakeStore) ListAccounts() ([]store.AccountRow, error) {
	out := make([]store.AccountRow, 0, len(f.rows))
	for _, r := range f.rows {
		out = append(out, r)
	}
	return out, nil
}

func (f *fakeStore) GetAccount(username string) (*store.AccountRow, error) {
	r, ok := f.rows[username]
	if !ok {
		return nil, nil
	}
	return &r, nil
}

func (f *fakeStore) UpsertAccount(row store.AccountRow) error {
	f.rows[row.Username] = row
	return nil
}

func (f *fakeStore) DeleteAccount(username string) error {
	delete(f.rows, username)
	return nil
}
func (f *fakeStore) Ping() error { return nil }

func (f *fakeStore) TouchLastStartedAt(string) error { return nil }
func (f *fakeStore) Changes() <-chan struct{}         { return f.changes }
func (f *fakeStore) Close() error                    { return nil }

// newTestAccountsServer returns a server with the given store wired in and a
// registered ServeMux (mirrors how server.go wires up routes).
func newTestAccountsServer(t *testing.T, st store.Store) *AnalyticsServer {
	t.Helper()
	s := NewAnalyticsServer(":0", newTestLogger(t), nil, "")
	s.SetAccountStore(st)
	return s
}

func accountsRequest(t *testing.T, s *AnalyticsServer, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var req *http.Request
	if body != nil {
		b, _ := json.Marshal(body)
		req = httptest.NewRequest(method, path, bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	rr := httptest.NewRecorder()
	s.srv.Handler.ServeHTTP(rr, req)
	return rr
}

// ── GET /api/accounts ─────────────────────────────────────────────────────────

func TestListAccounts_Empty(t *testing.T) {
	s := newTestAccountsServer(t, newFakeStore())
	rr := accountsRequest(t, s, "GET", "/api/accounts", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var out []any
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if len(out) != 0 {
		t.Fatalf("expected empty list, got %d entries", len(out))
	}
}

func TestListAccounts_WithEntries(t *testing.T) {
	st := newFakeStore()
	st.rows["alice"] = store.AccountRow{Username: "alice", Enabled: true, UpdatedAt: time.Now()}
	st.rows["bob"] = store.AccountRow{Username: "bob", Enabled: false, UpdatedAt: time.Now()}

	s := newTestAccountsServer(t, st)
	rr := accountsRequest(t, s, "GET", "/api/accounts", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var out []map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if len(out) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(out))
	}
}

func TestListAccounts_NoStore(t *testing.T) {
	s := newTestAccountsServer(t, nil)
	rr := accountsRequest(t, s, "GET", "/api/accounts", nil)
	if rr.Code != http.StatusNotImplemented {
		t.Fatalf("expected 501, got %d", rr.Code)
	}
}

// ── GET /api/accounts/{username} ─────────────────────────────────────────────

func TestGetAccount_Found(t *testing.T) {
	st := newFakeStore()
	cfg := minimalAccountCfg("alice")
	blob, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	st.rows["alice"] = store.AccountRow{
		Username:   "alice",
		ConfigJSON: string(blob),
		Enabled:    true,
		UpdatedAt:  time.Now(),
	}

	s := newTestAccountsServer(t, st)
	rr := accountsRequest(t, s, "GET", "/api/accounts/alice", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var out map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out["username"] != "alice" {
		t.Errorf("expected username alice, got %v", out["username"])
	}
	if out["config"] == nil {
		t.Error("expected config field in response")
	}
}

func TestGetAccount_NotFound(t *testing.T) {
	s := newTestAccountsServer(t, newFakeStore())
	rr := accountsRequest(t, s, "GET", "/api/accounts/nobody", nil)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestGetAccount_NoStore(t *testing.T) {
	s := newTestAccountsServer(t, nil)
	rr := accountsRequest(t, s, "GET", "/api/accounts/alice", nil)
	if rr.Code != http.StatusNotImplemented {
		t.Fatalf("expected 501, got %d", rr.Code)
	}
}

// ── POST /api/accounts ────────────────────────────────────────────────────────

func TestCreateAccount_Valid(t *testing.T) {
	st := newFakeStore()
	s := newTestAccountsServer(t, st)

	rr := accountsRequest(t, s, "POST", "/api/accounts", minimalAccountCfg("newuser"))
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}
	if _, ok := st.rows["newuser"]; !ok {
		t.Error("account not persisted in store")
	}
}

func TestCreateAccount_MissingUsername(t *testing.T) {
	s := newTestAccountsServer(t, newFakeStore())
	rr := accountsRequest(t, s, "POST", "/api/accounts", map[string]any{})
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestCreateAccount_InvalidJSON(t *testing.T) {
	s := newTestAccountsServer(t, newFakeStore())
	req := httptest.NewRequest("POST", "/api/accounts", bytes.NewBufferString("{invalid"))
	rr := httptest.NewRecorder()
	s.srv.Handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestCreateAccount_ValidationFail(t *testing.T) {
	s := newTestAccountsServer(t, newFakeStore())
	// username but no streamers/followers/watchers — fails config.Validate
	rr := accountsRequest(t, s, "POST", "/api/accounts", map[string]any{
		"username": "bad",
	})
	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestCreateAccount_NoStore(t *testing.T) {
	s := newTestAccountsServer(t, nil)
	rr := accountsRequest(t, s, "POST", "/api/accounts", minimalAccountCfg("x"))
	if rr.Code != http.StatusNotImplemented {
		t.Fatalf("expected 501, got %d", rr.Code)
	}
}

// ── PUT /api/accounts/{username} ─────────────────────────────────────────────

func TestUpdateAccount_Valid(t *testing.T) {
	st := newFakeStore()
	st.rows["alice"] = store.AccountRow{Username: "alice", ConfigJSON: "{}", Enabled: true}

	s := newTestAccountsServer(t, st)
	updated := minimalAccountCfg("alice")
	rr := accountsRequest(t, s, "PUT", "/api/accounts/alice", updated)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestUpdateAccount_InvalidJSON(t *testing.T) {
	s := newTestAccountsServer(t, newFakeStore())
	req := httptest.NewRequest("PUT", "/api/accounts/alice", bytes.NewBufferString("{bad"))
	rr := httptest.NewRecorder()
	s.srv.Handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestUpdateAccount_ValidationFail(t *testing.T) {
	s := newTestAccountsServer(t, newFakeStore())
	rr := accountsRequest(t, s, "PUT", "/api/accounts/alice", map[string]any{
		"username": "alice",
	})
	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestUpdateAccount_NoStore(t *testing.T) {
	s := newTestAccountsServer(t, nil)
	rr := accountsRequest(t, s, "PUT", "/api/accounts/alice", minimalAccountCfg("alice"))
	if rr.Code != http.StatusNotImplemented {
		t.Fatalf("expected 501, got %d", rr.Code)
	}
}

// ── DELETE /api/accounts/{username} ──────────────────────────────────────────

func TestDeleteAccount_Existing(t *testing.T) {
	st := newFakeStore()
	st.rows["alice"] = store.AccountRow{Username: "alice"}

	s := newTestAccountsServer(t, st)
	rr := accountsRequest(t, s, "DELETE", "/api/accounts/alice", nil)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}
	if _, ok := st.rows["alice"]; ok {
		t.Error("expected account to be removed from store")
	}
}

func TestDeleteAccount_NoStore(t *testing.T) {
	s := newTestAccountsServer(t, nil)
	rr := accountsRequest(t, s, "DELETE", "/api/accounts/alice", nil)
	if rr.Code != http.StatusNotImplemented {
		t.Fatalf("expected 501, got %d", rr.Code)
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

// minimalAccountCfg returns the smallest valid AccountConfig for the given username.
func minimalAccountCfg(username string) map[string]any {
	return map[string]any{
		"username":          username,
		"max_watch_streams": 1,
		"streamers": []any{
			map[string]any{"username": "some_streamer"},
		},
	}
}
