package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Guliveer/twitch-miner-go/internal/config"
	"github.com/Guliveer/twitch-miner-go/internal/store"
)

// SetAccountStore wires up the accounts REST API. When st is nil, the endpoints
// respond 501 Not Implemented (DB mode not enabled).
func (s *AnalyticsServer) SetAccountStore(st store.Store) {
	s.mu.Lock()
	s.accountStore = st
	s.mu.Unlock()
}

// handleListAccounts GET /api/accounts
func (s *AnalyticsServer) handleListAccounts(w http.ResponseWriter, r *http.Request) {
	st := s.getAccountStore()
	if st == nil {
		http.Error(w, "DB mode not enabled", http.StatusNotImplemented)
		return
	}
	rows, err := st.ListAccounts()
	if err != nil {
		writeInternalError(w, s.log, "GET /api/accounts", err)
		return
	}

	type accountSummary struct {
		Username      string     `json:"username"`
		Enabled       bool       `json:"enabled"`
		UpdatedAt     time.Time  `json:"updated_at"`
		LastStartedAt *time.Time `json:"last_started_at"`
	}
	out := make([]accountSummary, 0, len(rows))
	for _, row := range rows {
		out = append(out, accountSummary{row.Username, row.Enabled, row.UpdatedAt, row.LastStartedAt})
	}
	writeJSON(w, http.StatusOK, out)
}


// handleGetAccount GET /api/accounts/{username}
func (s *AnalyticsServer) handleGetAccount(w http.ResponseWriter, r *http.Request) {
	st := s.getAccountStore()
	if st == nil {
		http.Error(w, "DB mode not enabled", http.StatusNotImplemented)
		return
	}

	username := r.PathValue("username")
	if username == "" {
		http.Error(w, "username is required", http.StatusBadRequest)
		return
	}

	row, err := st.GetAccount(username)
	if err != nil {
		writeInternalError(w, s.log, "GET /api/accounts/{username}", err)
		return
	}
	if row == nil {
		http.Error(w, "account not found", http.StatusNotFound)
		return
	}

	cfg, err := config.AccountConfigFromJSON(row.Username, row.ConfigJSON)
	if err != nil {
		writeInternalError(w, s.log, "GET /api/accounts/{username}", err)
		return
	}

	type accountDetail struct {
		Username      string                `json:"username"`
		Enabled       bool                  `json:"enabled"`
		UpdatedAt     time.Time             `json:"updated_at"`
		LastStartedAt *time.Time            `json:"last_started_at"`
		Config        *config.AccountConfig `json:"config"`
	}
	writeJSON(w, http.StatusOK, accountDetail{
		Username:      row.Username,
		Enabled:       row.Enabled,
		UpdatedAt:     row.UpdatedAt,
		LastStartedAt: row.LastStartedAt,
		Config:        cfg,
	})
}

// handleCreateAccount POST /api/accounts
func (s *AnalyticsServer) handleCreateAccount(w http.ResponseWriter, r *http.Request) {
	st := s.getAccountStore()
	if st == nil {
		http.Error(w, "DB mode not enabled", http.StatusNotImplemented)
		return
	}

	var cfg config.AccountConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	if cfg.Username == "" {
		http.Error(w, "username is required", http.StatusBadRequest)
		return
	}
	if err := config.Validate(&cfg); err != nil {
		http.Error(w, "invalid config: "+err.Error(), http.StatusUnprocessableEntity)
		return
	}

	blob, err := config.AccountConfigToJSON(&cfg)
	if err != nil {
		writeInternalError(w, s.log, "POST /api/accounts", err)
		return
	}

	row := store.AccountRow{
		Username:   cfg.Username,
		ConfigJSON: blob,
		Enabled:    cfg.IsEnabled(),
		UpdatedAt:  time.Now(),
	}
	if err := st.UpsertAccount(row); err != nil {
		writeInternalError(w, s.log, "POST /api/accounts", err)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// handleUpdateAccount PUT /api/accounts/{username}
func (s *AnalyticsServer) handleUpdateAccount(w http.ResponseWriter, r *http.Request) {
	st := s.getAccountStore()
	if st == nil {
		http.Error(w, "DB mode not enabled", http.StatusNotImplemented)
		return
	}

	username := r.PathValue("username")
	if username == "" {
		http.Error(w, "username is required", http.StatusBadRequest)
		return
	}

	var cfg config.AccountConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	cfg.Username = username

	if err := config.Validate(&cfg); err != nil {
		http.Error(w, "invalid config: "+err.Error(), http.StatusUnprocessableEntity)
		return
	}

	blob, err := config.AccountConfigToJSON(&cfg)
	if err != nil {
		writeInternalError(w, s.log, "PUT /api/accounts/{username}", err)
		return
	}

	row := store.AccountRow{
		Username:   cfg.Username,
		ConfigJSON: blob,
		Enabled:    cfg.IsEnabled(),
		UpdatedAt:  time.Now(),
	}
	if err := st.UpsertAccount(row); err != nil {
		writeInternalError(w, s.log, "PUT /api/accounts/{username}", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleDeleteAccount DELETE /api/accounts/{username}
func (s *AnalyticsServer) handleDeleteAccount(w http.ResponseWriter, r *http.Request) {
	st := s.getAccountStore()
	if st == nil {
		http.Error(w, "DB mode not enabled", http.StatusNotImplemented)
		return
	}

	username := r.PathValue("username")
	if username == "" {
		http.Error(w, "username is required", http.StatusBadRequest)
		return
	}

	if err := st.DeleteAccount(username); err != nil {
		writeInternalError(w, s.log, "DELETE /api/accounts/{username}", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *AnalyticsServer) getAccountStore() store.Store {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.accountStore
}
