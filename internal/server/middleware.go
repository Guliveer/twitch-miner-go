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
	clients map[string]*tokenBucket
	mu      sync.Mutex
	rate    int
	period  time.Duration
}

type tokenBucket struct {
	tokens   int
	lastTime time.Time
}

func newRateLimiter(rps int, period time.Duration) *rateLimiter {
	return &rateLimiter{
		clients: make(map[string]*tokenBucket),
		rate:    rps,
		period:  period,
	}
}

func (rl *rateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	bucket, exists := rl.clients[key]
	if !exists {
		rl.clients[key] = &tokenBucket{tokens: rl.rate - 1, lastTime: now}
		return true
	}

	elapsed := now.Sub(bucket.lastTime)
	refill := int(elapsed.Seconds() * float64(rl.rate) / rl.period.Seconds())
	bucket.tokens += refill
	if bucket.tokens > rl.rate {
		bucket.tokens = rl.rate
	}
	bucket.lastTime = now

	if bucket.tokens <= 0 {
		return false
	}
	bucket.tokens--
	return true
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
