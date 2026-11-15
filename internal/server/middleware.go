package server

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Guliveer/twitch-miner-go/internal/logger"
)

type contextKey string

const requestIDKey contextKey = "request_id"

func RequestIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}

func withRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = generateRequestID()
		}
		ctx := context.WithValue(r.Context(), requestIDKey, id)
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func generateRequestID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// withAuth enforces authentication via either X-API-Key (machine clients) or
// HTTP Basic Auth (browser users). Health endpoint and static assets are excluded.
func withAuth(creds *DashboardAuth, apiKey string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" || strings.HasPrefix(r.URL.Path, "/static/") {
			next.ServeHTTP(w, r)
			return
		}

		if apiKey != "" {
			key := r.Header.Get("X-API-Key")
			if subtle.ConstantTimeCompare([]byte(key), []byte(apiKey)) == 1 {
				next.ServeHTTP(w, r)
				return
			}
		}

		if creds != nil {
			user, pass, ok := r.BasicAuth()
			if ok && checkCredentials(user, pass, creds) {
				next.ServeHTTP(w, r)
				return
			}
		}

		w.Header().Set("WWW-Authenticate", `Basic realm="Twitch Miner Dashboard"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})
}

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

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

type rateLimiter struct {
	clients   map[string]*tokenBucket
	mu        sync.Mutex
	now       func() time.Time
	lastSweep time.Time
	rate      int
	period    time.Duration
}

type tokenBucket struct {
	tokens   int
	lastTime time.Time
}

func newRateLimiter(rps int, period time.Duration) *rateLimiter {
	if rps <= 0 {
		panic("server: rate limiter rate must be positive")
	}
	if period <= 0 {
		panic("server: rate limiter period must be positive")
	}

	return &rateLimiter{
		clients: make(map[string]*tokenBucket),
		now:     time.Now,
		rate:    rps,
		period:  period,
	}
}

func (rl *rateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := rl.now()
	rl.evictIdleLocked(now)

	bucket, exists := rl.clients[key]
	if !exists {
		rl.clients[key] = &tokenBucket{tokens: rl.rate - 1, lastTime: now}
		return true
	}

	// Advancing lastTime by only the time actually converted into tokens keeps
	// the remainder, so clients polling faster than one token per period still
	// refill instead of having their elapsed time truncated away.
	perToken := rl.period / time.Duration(rl.rate)
	refill := int(now.Sub(bucket.lastTime) / perToken)
	if refill > 0 {
		bucket.tokens += refill
		if bucket.tokens > rl.rate {
			bucket.tokens = rl.rate
		}
		bucket.lastTime = bucket.lastTime.Add(time.Duration(refill) * perToken)
	}

	if bucket.tokens <= 0 {
		return false
	}
	bucket.tokens--
	return true
}

// A bucket that has been idle for a full period is refilled to its cap, so
// dropping it is equivalent to keeping it and stops the map from growing with
// every client address that ever connected.
func (rl *rateLimiter) evictIdleLocked(now time.Time) {
	if !rl.lastSweep.IsZero() && now.Sub(rl.lastSweep) < rl.period {
		return
	}
	rl.lastSweep = now

	for key, bucket := range rl.clients {
		if now.Sub(bucket.lastTime) >= rl.period {
			delete(rl.clients, key)
		}
	}
}

func withRateLimit(limiter *rateLimiter, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" || strings.HasPrefix(r.URL.Path, "/static/") {
			next.ServeHTTP(w, r)
			return
		}

		key := r.RemoteAddr
		if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
			key = strings.Split(fwd, ",")[0]
		}

		if !limiter.allow(strings.TrimSpace(key)) {
			w.Header().Set("Retry-After", "1")
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
