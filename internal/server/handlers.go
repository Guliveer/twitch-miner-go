package server

import (
	"net/http"
	"time"
)

func (s *AnalyticsServer) handleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(dashboardHTML) //nolint:errcheck // static HTML to ResponseWriter; error not actionable
}

func (s *AnalyticsServer) handleLogs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(logsHTML) //nolint:errcheck // static HTML to ResponseWriter; error not actionable
}

func (s *AnalyticsServer) handleHealth(w http.ResponseWriter, _ *http.Request) {
	checks := map[string]string{"status": "ok"}

	s.mu.RLock()
	st := s.accountStore
	s.mu.RUnlock()

	if st != nil {
		if err := st.Ping(); err != nil {
			checks["db"] = "unreachable"
			checks["status"] = "degraded"
		}
	}

	status := http.StatusOK
	if checks["status"] != "ok" {
		status = http.StatusServiceUnavailable
	}

	checks["timestamp"] = time.Now().UTC().Format(time.RFC3339)
	writeJSON(w, status, checks)
}
