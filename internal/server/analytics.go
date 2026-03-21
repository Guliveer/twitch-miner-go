// Package server provides a lightweight HTTP analytics server that exposes
// streamer data, statistics, and a simple dashboard.
package server

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"net/http/pprof"
	"strings"
	"sync"
	"time"

	"github.com/Guliveer/twitch-miner-go/internal/constants"
	"github.com/Guliveer/twitch-miner-go/internal/logger"
	"github.com/Guliveer/twitch-miner-go/internal/model"
	"github.com/Guliveer/twitch-miner-go/internal/utils"
)

// StreamerFunc is a function that returns the current list of streamers
// across all miners. Used to dynamically fetch streamer data.
type StreamerFunc func() []*model.Streamer

// NotifyTestFunc is a function that sends a test notification to all configured
// notifiers across all miners. Returns any errors encountered.
type NotifyTestFunc func(ctx context.Context) []error

// DebugSnapshotFunc returns a debug snapshot that can be serialized as JSON.
type DebugSnapshotFunc func() any

// DashboardAuth holds credentials for HTTP Basic Auth on the dashboard.
// The password is stored as a SHA-256 hex digest for constant-time comparison.
type DashboardAuth struct {
	Username     string
	PasswordHash string // hex-encoded SHA-256
}

// AnalyticsServer serves the analytics dashboard and JSON API endpoints.
type AnalyticsServer struct {
	addr string
	log  *logger.Logger
	srv  *http.Server
	auth *DashboardAuth

	mu             sync.RWMutex
	streamers      []*model.Streamer
	streamerFunc   StreamerFunc
	notifyTestFunc NotifyTestFunc
	debugFunc      DebugSnapshotFunc
}

// NewAnalyticsServer creates a new AnalyticsServer bound to the given address.
// If auth is non-nil, all endpoints (except /health and /static) require
// HTTP Basic Auth with the configured username and SHA-256 hashed password.
func NewAnalyticsServer(addr string, log *logger.Logger, auth *DashboardAuth) *AnalyticsServer {
	s := &AnalyticsServer{
		addr: addr,
		log:  log,
		auth: auth,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", s.handleDashboard)
	mux.HandleFunc("GET /logs", s.handleLogs)
	mux.HandleFunc("GET /health", s.handleHealth)
	mux.HandleFunc("GET /api/streamers", s.handleStreamers)
	mux.HandleFunc("GET /api/streamer/{name}", s.handleStreamer)
	mux.HandleFunc("GET /api/stats", s.handleStats)
	mux.HandleFunc("GET /api/filters", s.handleFilters)
	mux.HandleFunc("GET /api/events", s.handleEventLogs)
	mux.HandleFunc("GET /api/event-filters", s.handleEventFilters)
	mux.HandleFunc("GET /api/debug", s.handleDebug)

	mux.HandleFunc("POST /api/test-notification", s.handleTestNotification)
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServerFS(staticFS)))

	// pprof endpoints for remote memory profiling
	mux.HandleFunc("GET /debug/pprof/", pprof.Index)
	mux.HandleFunc("GET /debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("GET /debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("GET /debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("GET /debug/pprof/trace", pprof.Trace)
	mux.Handle("GET /debug/pprof/heap", pprof.Handler("heap"))
	mux.Handle("GET /debug/pprof/goroutine", pprof.Handler("goroutine"))
	mux.Handle("GET /debug/pprof/allocs", pprof.Handler("allocs"))

	var handler http.Handler = mux
	if auth != nil {
		handler = withBasicAuth(auth, mux)
	}

	s.srv = &http.Server{
		Addr:              addr,
		Handler:           withLogging(log, handler),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
		BaseContext: func(_ net.Listener) context.Context {
			return context.Background()
		},
	}

	return s
}

// SetStreamers updates the streamer list reference. Thread-safe.
func (s *AnalyticsServer) SetStreamers(streamers []*model.Streamer) {
	s.mu.Lock()
	s.streamers = streamers
	s.mu.Unlock()
}

// SetNotifyTestFunc sets a function that sends test notifications to all
// configured notifiers. Thread-safe.
func (s *AnalyticsServer) SetNotifyTestFunc(fn NotifyTestFunc) {
	s.mu.Lock()
	s.notifyTestFunc = fn
	s.mu.Unlock()
}

// SetDebugFunc sets a function that returns a debug snapshot across all miners.
func (s *AnalyticsServer) SetDebugFunc(fn DebugSnapshotFunc) {
	s.mu.Lock()
	s.debugFunc = fn
	s.mu.Unlock()
}

// SetStreamerFunc sets a function that dynamically returns all streamers
// across all miners. When set, getStreamers() calls this function instead
// of returning the static list.
func (s *AnalyticsServer) SetStreamerFunc(fn StreamerFunc) {
	s.mu.Lock()
	s.streamerFunc = fn
	s.mu.Unlock()
}

// getStreamers returns the current streamer list. Thread-safe.
func (s *AnalyticsServer) getStreamers() []*model.Streamer {
	s.mu.RLock()
	fn := s.streamerFunc
	s.mu.RUnlock()

	if fn != nil {
		return fn()
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.streamers
}

// Run starts the HTTP server and blocks until the context is cancelled.
// It performs graceful shutdown when the context is done.
func (s *AnalyticsServer) Run(ctx context.Context) error {
	s.log.Info("Analytics server started", "address", "http://"+s.addr)

	errCh := make(chan error, 1)
	utils.SafeGo(func() {
		if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("analytics server: %w", err)
		}
		close(errCh)
	})

	select {
	case <-ctx.Done():
		s.log.Info("Analytics server shutting down")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), constants.DefaultGracefulShutdownTimeout)
		defer cancel()
		if err := s.srv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("analytics server shutdown: %w", err)
		}
		return ctx.Err()
	case err := <-errCh:
		return err
	}
}

// withBasicAuth enforces HTTP Basic Auth with SHA-256 password comparison.
// Health endpoint and static assets are excluded.
func withBasicAuth(creds *DashboardAuth, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Always allow health checks and static assets without auth.
		if r.URL.Path == "/health" || strings.HasPrefix(r.URL.Path, "/static/") {
			next.ServeHTTP(w, r)
			return
		}

		user, pass, ok := r.BasicAuth()
		if !ok || !checkCredentials(user, pass, creds) {
			w.Header().Set("WWW-Authenticate", `Basic realm="Twitch Miner Dashboard"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// checkCredentials verifies username and password against stored credentials.
// The password is hashed with SHA-256 and compared in constant time.
func checkCredentials(user, pass string, creds *DashboardAuth) bool {
	userMatch := subtle.ConstantTimeCompare([]byte(user), []byte(creds.Username)) == 1

	hash := sha256.Sum256([]byte(pass))
	passHash := hex.EncodeToString(hash[:])
	passMatch := subtle.ConstantTimeCompare([]byte(passHash), []byte(creds.PasswordHash)) == 1

	return userMatch && passMatch
}

func withLogging(log *logger.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rw, r)
		log.Debug("HTTP request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rw.statusCode,
			"duration", time.Since(start).String(),
		)
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code before writing it.
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
