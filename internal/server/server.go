// Package server provides a lightweight HTTP analytics server that exposes
// streamer data, statistics, and a simple dashboard.
package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/pprof"
	"sync"
	"time"

	"github.com/Guliveer/twitch-miner-go/internal/constants"
	"github.com/Guliveer/twitch-miner-go/internal/logger"
	"github.com/Guliveer/twitch-miner-go/internal/model"
	"github.com/Guliveer/twitch-miner-go/internal/store"
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

// AuthStatusFunc returns the pending device code auth status for a given miner
// username, or nil if the miner is not found or has no pending flow.
type AuthStatusFunc func(username string) any

// MinerCountFunc returns the number of miners the manager currently runs.
type MinerCountFunc func() int

// DashboardAuth holds credentials for HTTP Basic Auth on the dashboard.
// The password is stored as a SHA-256 hex digest for constant-time comparison.
type DashboardAuth struct {
	Username     string
	PasswordHash string // hex-encoded SHA-256
}

// AnalyticsServer serves the analytics dashboard and JSON API endpoints.
type AnalyticsServer struct {
	addr   string
	log    *logger.Logger
	srv    *http.Server
	auth   *DashboardAuth
	apiKey string

	mu             sync.RWMutex
	streamers      []*model.Streamer
	streamerFunc   StreamerFunc
	notifyTestFunc NotifyTestFunc
	debugFunc      DebugSnapshotFunc
	authStatusFunc AuthStatusFunc
	minerCountFunc MinerCountFunc
	accountStore   store.Store
}

// NewAnalyticsServer creates a new AnalyticsServer bound to the given address.
// If auth or apiKey is provided, all endpoints (except /health and /static)
// require either HTTP Basic Auth (browser users) or an X-API-Key header (machine clients).
func NewAnalyticsServer(addr string, log *logger.Logger, auth *DashboardAuth, apiKey string) *AnalyticsServer {
	s := &AnalyticsServer{
		addr:   addr,
		log:    log,
		auth:   auth,
		apiKey: apiKey,
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

	mux.HandleFunc("GET /api/auth-status/{username}", s.handleAuthStatus)

	mux.HandleFunc("GET /api/accounts", s.handleListAccounts)
	mux.HandleFunc("POST /api/accounts", s.handleCreateAccount)
	mux.HandleFunc("GET /api/accounts/{username}", s.handleGetAccount)
	mux.HandleFunc("PUT /api/accounts/{username}", s.handleUpdateAccount)
	mux.HandleFunc("DELETE /api/accounts/{username}", s.handleDeleteAccount)

	mux.HandleFunc("GET /api/config/schema", s.handleConfigSchema)
	mux.HandleFunc("POST /api/config/validate", s.handleConfigValidate)
	mux.HandleFunc("POST /api/config/generate", s.handleConfigGenerate)

	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServerFS(staticFS)))

	// Heap and goroutine dumps expose live Twitch auth tokens held in process
	// memory, so the profiling endpoints only exist once the server is actually
	// protected. Registering them unconditionally leaked those tokens to anyone
	// who could reach the published port.
	authConfigured := auth != nil || apiKey != ""
	if authConfigured {
		mux.HandleFunc("GET /debug/pprof/", pprof.Index)
		mux.HandleFunc("GET /debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("GET /debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("GET /debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("GET /debug/pprof/trace", pprof.Trace)
		mux.Handle("GET /debug/pprof/heap", pprof.Handler("heap"))
		mux.Handle("GET /debug/pprof/goroutine", pprof.Handler("goroutine"))
		mux.Handle("GET /debug/pprof/allocs", pprof.Handler("allocs"))
	} else {
		log.Warn("Dashboard auth is not configured — profiling endpoints disabled and all endpoints are public",
			"hint", "set DASHBOARD_USER + DASHBOARD_PASSWORD_SHA256, or DASHBOARD_API_KEY")
	}

	var handler http.Handler = mux
	if authConfigured {
		handler = withAuth(auth, apiKey, mux)
	}

	limiter := newRateLimiter(100, time.Minute)
	handler = withRateLimit(limiter, handler)
	handler = withRequestID(handler)

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

// SetAuthStatusFunc sets a function that returns the device-code auth status
// for a given username. Thread-safe.
func (s *AnalyticsServer) SetAuthStatusFunc(fn AuthStatusFunc) {
	s.mu.Lock()
	s.authStatusFunc = fn
	s.mu.Unlock()
}

// SetMinerCountFunc sets a function returning how many miners are running.
// When set, /health reports "degraded" while the count is zero, so a container
// that loaded no usable config is not reported as healthy. Thread-safe.
func (s *AnalyticsServer) SetMinerCountFunc(fn MinerCountFunc) {
	s.mu.Lock()
	s.minerCountFunc = fn
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
