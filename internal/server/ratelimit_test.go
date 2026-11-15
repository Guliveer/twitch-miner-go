package server

import "testing"

// The dashboard requests /api/auth-status once per account card every 5s, which
// is 12 requests per minute per account. The configured limit must leave room for
// that load, otherwise the limiter throttles a single well-behaved dashboard.
func TestDashboardRateLimitCoversAccountPolling(t *testing.T) {
	const (
		accounts          = 16
		pollsPerMinute    = 12
		requestsPerMinute = accounts * pollsPerMinute
		trackedClientAddr = "203.0.113.10"
	)

	rl := newRateLimiter(dashboardRateLimit, dashboardRateLimitPeriod)

	for i := 0; i < requestsPerMinute; i++ {
		if !rl.allow(trackedClientAddr) {
			t.Fatalf("dashboard polling (%d req/min) was throttled at request %d with a limit of %d per %s",
				requestsPerMinute, i+1, dashboardRateLimit, dashboardRateLimitPeriod)
		}
	}
}
