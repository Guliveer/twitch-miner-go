package server

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/Guliveer/twitch-miner-go/internal/logger"
)

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
