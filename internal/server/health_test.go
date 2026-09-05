package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func healthResponse(t *testing.T, s *AnalyticsServer) (int, map[string]string) {
	t.Helper()

	rec := httptest.NewRecorder()
	s.handleHealth(rec, httptest.NewRequest(http.MethodGet, "/health", nil))

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decoding health body: %v", err)
	}
	return rec.Code, body
}

func TestHealthDegradedWhenNoMinersRunning(t *testing.T) {
	s := NewAnalyticsServer(":0", newTestLogger(t), nil, "")
	s.SetMinerCountFunc(func() int { return 0 })

	code, body := healthResponse(t, s)

	if code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 while no miner runs, got %d", code)
	}
	if body["status"] != "degraded" {
		t.Fatalf("expected status degraded, got %q", body["status"])
	}
	if body["miners"] != "0" {
		t.Fatalf("expected miners=0, got %q", body["miners"])
	}
}

func TestHealthOKWhenMinersRunning(t *testing.T) {
	s := NewAnalyticsServer(":0", newTestLogger(t), nil, "")
	s.SetMinerCountFunc(func() int { return 2 })

	code, body := healthResponse(t, s)

	if code != http.StatusOK {
		t.Fatalf("expected 200 with miners running, got %d", code)
	}
	if body["status"] != "ok" {
		t.Fatalf("expected status ok, got %q", body["status"])
	}
	if body["miners"] != "2" {
		t.Fatalf("expected miners=2, got %q", body["miners"])
	}
}

// Without the setter wired up the endpoint must keep its previous semantics,
// so an embedder that never calls SetMinerCountFunc is not reported unhealthy.
func TestHealthOKWhenMinerCountUnknown(t *testing.T) {
	s := NewAnalyticsServer(":0", newTestLogger(t), nil, "")

	code, body := healthResponse(t, s)

	if code != http.StatusOK {
		t.Fatalf("expected 200 without a miner count func, got %d", code)
	}
	if _, ok := body["miners"]; ok {
		t.Fatalf("expected no miners key, got %q", body["miners"])
	}
}

func TestPprofNotRegisteredWithoutAuth(t *testing.T) {
	s := NewAnalyticsServer(":0", newTestLogger(t), nil, "")

	rec := httptest.NewRecorder()
	s.srv.Handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/debug/pprof/heap", nil))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected pprof to be absent without auth, got %d", rec.Code)
	}
}

func TestPprofRequiresAuthWhenRegistered(t *testing.T) {
	s := NewAnalyticsServer(":0", newTestLogger(t), nil, "secret-key")

	rec := httptest.NewRecorder()
	s.srv.Handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/debug/pprof/heap", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without the API key, got %d", rec.Code)
	}

	authed := httptest.NewRequest(http.MethodGet, "/debug/pprof/heap", nil)
	authed.Header.Set("X-API-Key", "secret-key")
	rec = httptest.NewRecorder()
	s.srv.Handler.ServeHTTP(rec, authed)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 with the API key, got %d", rec.Code)
	}
}
