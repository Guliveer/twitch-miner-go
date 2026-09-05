package server

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Guliveer/twitch-miner-go/internal/version"
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
	checks := map[string]string{"status": "ok", "version": version.String()}

	s.mu.RLock()
	st := s.accountStore
	countFn := s.minerCountFunc
	s.mu.RUnlock()

	if st != nil {
		if err := st.Ping(); err != nil {
			checks["db"] = "unreachable"
			checks["status"] = "degraded"
		}
	}

	// A container that loaded no usable account config would otherwise report
	// healthy while mining nothing, which is indistinguishable from success in
	// an orchestrator's status view.
	if countFn != nil {
		count := countFn()
		checks["miners"] = strconv.Itoa(count)
		if count == 0 {
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
