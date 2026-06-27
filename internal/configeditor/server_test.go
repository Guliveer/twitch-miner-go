package configeditor

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func newTestServer(t *testing.T) (*Server, string) {
	t.Helper()
	dir := t.TempDir()
	return NewServer(dir), dir
}

func writeYAML(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name+".yaml"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func apiRequest(t *testing.T, srv *Server, method, path string, body any) *httptest.ResponseRecorder {
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
	srv.ServeHTTP(rr, req)
	return rr
}

// ── Schema ────────────────────────────────────────────────────────────────────

func TestGetSchema(t *testing.T) {
	srv, _ := newTestServer(t)
	rr := apiRequest(t, srv, "GET", "/api/schema", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var schema map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &schema); err != nil {
		t.Fatal(err)
	}
	if _, ok := schema["strategies"]; !ok {
		t.Error("schema missing 'strategies'")
	}
	if _, ok := schema["priorities"]; !ok {
		t.Error("schema missing 'priorities'")
	}
}

// ── List accounts ─────────────────────────────────────────────────────────────

func TestListAccountsEmpty(t *testing.T) {
	srv, _ := newTestServer(t)
	rr := apiRequest(t, srv, "GET", "/api/accounts", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	// Must be [] not null
	if rr.Body.String()[:2] != "[{" && rr.Body.String() != "[]\n" {
		body := rr.Body.String()
		if body == "null\n" || body == "null" {
			t.Error("empty accounts returned null instead of []")
		}
	}
	var accounts []accountMeta
	if err := json.Unmarshal(rr.Body.Bytes(), &accounts); err != nil {
		t.Fatal(err)
	}
	if accounts == nil {
		t.Error("accounts slice must not be nil (must serialize as [])")
	}
	if len(accounts) != 0 {
		t.Errorf("expected 0 accounts, got %d", len(accounts))
	}
}

func TestListAccountsSkipsExamples(t *testing.T) {
	srv, dir := newTestServer(t)
	writeYAML(t, dir, "real_user", "streamers:\n  - username: foo\n")
	if err := os.WriteFile(filepath.Join(dir, "example.yaml.example"), []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
	rr := apiRequest(t, srv, "GET", "/api/accounts", nil)
	var accounts []accountMeta
	json.Unmarshal(rr.Body.Bytes(), &accounts)
	if len(accounts) != 1 || accounts[0].Name != "real_user" {
		t.Errorf("expected [real_user], got %+v", accounts)
	}
}

// ── Create account ────────────────────────────────────────────────────────────

func TestCreateAccount(t *testing.T) {
	srv, dir := newTestServer(t)
	body := map[string]any{
		"name":   "myuser",
		"config": map[string]any{"streamers": []any{map[string]any{"username": "s1"}}},
	}
	rr := apiRequest(t, srv, "POST", "/api/accounts", body)
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}
	if _, err := os.Stat(filepath.Join(dir, "myuser.yaml")); err != nil {
		t.Error("YAML file not created")
	}
}

func TestCreateAccountInvalidName(t *testing.T) {
	srv, _ := newTestServer(t)
	body := map[string]any{"name": "has space", "config": map[string]any{}}
	rr := apiRequest(t, srv, "POST", "/api/accounts", body)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestCreateAccountConflict(t *testing.T) {
	srv, dir := newTestServer(t)
	writeYAML(t, dir, "existing", "streamers:\n  - username: foo\n")
	body := map[string]any{
		"name":   "existing",
		"config": map[string]any{"streamers": []any{map[string]any{"username": "foo"}}},
	}
	rr := apiRequest(t, srv, "POST", "/api/accounts", body)
	if rr.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rr.Code)
	}
}

func TestCreateAccountValidationFail(t *testing.T) {
	srv, _ := newTestServer(t)
	// no streamers, no followers, no watchers
	body := map[string]any{"name": "bad", "config": map[string]any{"max_watch_streams": float64(2)}}
	rr := apiRequest(t, srv, "POST", "/api/accounts", body)
	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d: %s", rr.Code, rr.Body.String())
	}
}

// ── Get account ───────────────────────────────────────────────────────────────

func TestGetAccount(t *testing.T) {
	srv, dir := newTestServer(t)
	writeYAML(t, dir, "alice", "streamers:\n  - username: streamer1\n")
	rr := apiRequest(t, srv, "GET", "/api/accounts/alice", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var cfg map[string]any
	json.Unmarshal(rr.Body.Bytes(), &cfg)
	streamers, _ := cfg["streamers"].([]any)
	if len(streamers) != 1 {
		t.Errorf("expected 1 streamer, got %d", len(streamers))
	}
}

func TestGetAccountNotFound(t *testing.T) {
	srv, _ := newTestServer(t)
	rr := apiRequest(t, srv, "GET", "/api/accounts/nobody", nil)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

// ── Secret stripping ──────────────────────────────────────────────────────────

func TestGetAccountStripsSecrets(t *testing.T) {
	srv, dir := newTestServer(t)
	writeYAML(t, dir, "secretuser", `
streamers:
  - username: s1
notifications:
  telegram:
    enabled: true
    token: "bot-secret-token"
    chat_id: "99999"
    events: []
  discord:
    enabled: false
    webhook_url: "https://discord.com/api/webhooks/secret"
    events: []
`)
	rr := apiRequest(t, srv, "GET", "/api/accounts/secretuser", nil)
	var cfg map[string]any
	json.Unmarshal(rr.Body.Bytes(), &cfg)

	notif, _ := cfg["notifications"].(map[string]any)
	tg, _ := notif["telegram"].(map[string]any)
	if _, hasToken := tg["token"]; hasToken {
		t.Error("telegram token must be stripped from GET response")
	}
	if _, hasChatID := tg["chat_id"]; hasChatID {
		t.Error("telegram chat_id must be stripped from GET response")
	}
	dc, _ := notif["discord"].(map[string]any)
	if _, hasWebhook := dc["webhook_url"]; hasWebhook {
		t.Error("discord webhook_url must be stripped from GET response")
	}
}

// ── Update account ────────────────────────────────────────────────────────────

func TestUpdateAccountPreservesSecrets(t *testing.T) {
	srv, dir := newTestServer(t)
	writeYAML(t, dir, "bob", `
streamers:
  - username: s1
notifications:
  telegram:
    enabled: true
    token: "original-token"
    chat_id: "42"
    events: []
`)
	// PUT without token/chat_id (as frontend would send after GET stripping)
	body := map[string]any{
		"config": map[string]any{
			"streamers": []any{map[string]any{"username": "s1"}},
			"notifications": map[string]any{
				"telegram": map[string]any{
					"enabled": true,
					"events":  []any{},
				},
			},
		},
	}
	rr := apiRequest(t, srv, "PUT", "/api/accounts/bob", body)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	// Re-load raw YAML and verify secrets are still there
	raw, _ := os.ReadFile(filepath.Join(dir, "bob.yaml"))
	content := string(raw)
	if !bytes.Contains(raw, []byte("original-token")) {
		t.Errorf("token not preserved in YAML:\n%s", content)
	}
	if !bytes.Contains(raw, []byte("42")) {
		t.Errorf("chat_id not preserved in YAML:\n%s", content)
	}
}

func TestUpdateAccountNotFound(t *testing.T) {
	srv, _ := newTestServer(t)
	body := map[string]any{"config": map[string]any{"streamers": []any{map[string]any{"username": "x"}}}}
	rr := apiRequest(t, srv, "PUT", "/api/accounts/ghost", body)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

// ── Delete account ────────────────────────────────────────────────────────────

func TestDeleteAccount(t *testing.T) {
	srv, dir := newTestServer(t)
	writeYAML(t, dir, "deleteme", "streamers:\n  - username: x\n")
	rr := apiRequest(t, srv, "DELETE", "/api/accounts/deleteme", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if _, err := os.Stat(filepath.Join(dir, "deleteme.yaml")); !os.IsNotExist(err) {
		t.Error("YAML file should have been deleted")
	}
}

func TestDeleteAccountNotFound(t *testing.T) {
	srv, _ := newTestServer(t)
	rr := apiRequest(t, srv, "DELETE", "/api/accounts/ghost", nil)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

// ── Validation ────────────────────────────────────────────────────────────────

func TestValidateConfig(t *testing.T) {
	cases := []struct {
		name    string
		cfg     map[string]any
		wantErr bool
	}{
		{
			name:    "valid with streamers",
			cfg:     map[string]any{"streamers": []any{map[string]any{"username": "foo"}}},
			wantErr: false,
		},
		{
			name:    "valid with followers",
			cfg:     map[string]any{"followers": map[string]any{"enabled": true}},
			wantErr: false,
		},
		{
			name:    "no sources",
			cfg:     map[string]any{"max_watch_streams": float64(2)},
			wantErr: true,
		},
		{
			name:    "empty streamer username",
			cfg:     map[string]any{"streamers": []any{map[string]any{"username": ""}}},
			wantErr: true,
		},
		{
			name: "category_watcher enabled but no categories",
			cfg: map[string]any{
				"category_watcher": map[string]any{"enabled": true, "categories": []any{}},
			},
			wantErr: true,
		},
		{
			name: "make_predictions without bet",
			cfg: map[string]any{
				"streamers":        []any{map[string]any{"username": "foo"}},
				"streamer_defaults": map[string]any{"make_predictions": true},
			},
			wantErr: true,
		},
		{
			name: "make_predictions with bet is valid",
			cfg: map[string]any{
				"streamers":        []any{map[string]any{"username": "foo"}},
				"streamer_defaults": map[string]any{"make_predictions": true, "bet": map[string]any{"strategy": "SMART"}},
			},
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			errs := validateConfig(tc.cfg)
			if tc.wantErr && len(errs) == 0 {
				t.Error("expected validation errors, got none")
			}
			if !tc.wantErr && len(errs) > 0 {
				t.Errorf("expected no errors, got: %v", errs)
			}
		})
	}
}

// ── Duration validation ───────────────────────────────────────────────────────

func TestIsValidDuration(t *testing.T) {
	valid := []string{"120s", "5m", "1h30m", "500ms", "1h", "2h30m15s"}
	invalid := []string{"", "5", "5x", "abc", "1.5", "-5s"}

	for _, s := range valid {
		if !isValidDuration(s) {
			t.Errorf("expected %q to be valid duration", s)
		}
	}
	for _, s := range invalid {
		if isValidDuration(s) {
			t.Errorf("expected %q to be invalid duration", s)
		}
	}
}

// ── cleanConfig ───────────────────────────────────────────────────────────────

func TestCleanConfig(t *testing.T) {
	cfg := map[string]any{
		"enabled":  true,
		"proxy":    "",
		"empty_map": map[string]any{},
		"streamers": []any{map[string]any{"username": "foo"}},
		"nil_field": nil,
	}
	cleaned := cleanConfig(cfg)
	if _, ok := cleaned["proxy"]; ok {
		t.Error("empty string proxy should be removed")
	}
	if _, ok := cleaned["empty_map"]; ok {
		t.Error("empty map should be removed")
	}
	if _, ok := cleaned["nil_field"]; ok {
		t.Error("nil field should be removed")
	}
	if _, ok := cleaned["enabled"]; !ok {
		t.Error("enabled=true should be kept")
	}
	streamers, _ := cleaned["streamers"].([]any)
	if len(streamers) != 1 {
		t.Error("streamers should be kept")
	}
}
