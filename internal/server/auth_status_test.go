package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func decodeObject(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decoding response body: %v", err)
	}
	return body
}

func TestAuthStatusesEmptyWithoutFunc(t *testing.T) {
	s := NewAnalyticsServer(":0", newTestLogger(t), nil, "")

	rec := httptest.NewRecorder()
	s.handleAuthStatuses(rec, httptest.NewRequest(http.MethodGet, "/api/auth-status", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if body := decodeObject(t, rec); len(body) != 0 {
		t.Fatalf("expected an empty map without a status list func, got %v", body)
	}
}

func TestAuthStatusesEmptyWhenFuncReturnsNil(t *testing.T) {
	s := NewAnalyticsServer(":0", newTestLogger(t), nil, "")
	s.SetAuthStatusListFunc(func() map[string]any { return nil })

	rec := httptest.NewRecorder()
	s.handleAuthStatuses(rec, httptest.NewRequest(http.MethodGet, "/api/auth-status", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if body := decodeObject(t, rec); len(body) != 0 {
		t.Fatalf("expected an empty map for a nil result, got %v", body)
	}
}

func TestAuthStatusesReturnsPendingFlows(t *testing.T) {
	s := NewAnalyticsServer(":0", newTestLogger(t), nil, "")
	s.SetAuthStatusListFunc(func() map[string]any {
		return map[string]any{
			"alice": map[string]string{
				"status":           "pending",
				"user_code":        "ABCD-1234",
				"verification_uri": "https://www.twitch.tv/activate",
			},
		}
	})

	rec := httptest.NewRecorder()
	s.handleAuthStatuses(rec, httptest.NewRequest(http.MethodGet, "/api/auth-status", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	body := decodeObject(t, rec)
	alice, ok := body["alice"].(map[string]any)
	if !ok {
		t.Fatalf("expected alice in the response, got %v", body)
	}
	if alice["user_code"] != "ABCD-1234" {
		t.Fatalf("expected the user code to be passed through, got %v", alice["user_code"])
	}
}

// The batched route must not be swallowed by the /api/auth-status/{username}
// pattern, which requires a path segment.
func TestAuthStatusesRouteIsRegistered(t *testing.T) {
	s := NewAnalyticsServer(":0", newTestLogger(t), nil, "")
	s.SetAuthStatusListFunc(func() map[string]any {
		return map[string]any{"bob": map[string]string{"status": "pending"}}
	})

	rec := httptest.NewRecorder()
	s.srv.Handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/auth-status", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 from the batched route, got %d", rec.Code)
	}

	body := decodeObject(t, rec)
	if _, ok := body["bob"]; !ok {
		t.Fatalf("expected bob in the batched response, got %v", body)
	}
}
