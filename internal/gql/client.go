// Package gql provides a typed GraphQL client for the Twitch GQL API.
// It handles connection pooling, request building, runtime-configured client
// version headers, rate limiting awareness, and error handling with retries.
package gql

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/Guliveer/twitch-miner-go/internal/auth"
	"github.com/Guliveer/twitch-miner-go/internal/constants"
	"github.com/Guliveer/twitch-miner-go/internal/logger"
)

// ErrCircuitOpen is returned when the circuit breaker is open and requests
// are being skipped to avoid hammering a failing API.
var ErrCircuitOpen = errors.New("circuit breaker open: API requests temporarily suspended")

// ErrTransientGQLError marks GQL responses that failed due to a temporary
// Twitch-side condition and should not be treated like a durable application
// error.
var ErrTransientGQLError = errors.New("transient GQL error")

type operationBehavior struct {
	skipIntegrity   bool
	failOnErrors    bool
	tryAltClientIDs bool
	retryGQLErrors  bool
}

// integrityFailureOps lists GQL operations where integrity check failures are
// expected and should be logged at DEBUG instead of WARN. These operations
// sometimes fail with "failed integrity check" but may still succeed on retry.
var integrityFailureOps = map[string]bool{
	"JoinRaid":                  true,
	"ClaimCommunityPoints":      true,
	"ViewerDropsDashboard":      true,
	"DropsPage_ClaimDropRewards": true,
}

// operationBehaviors centralizes per-operation compatibility workarounds for
// Twitch's unstable internal GQL APIs.
var operationBehaviors = map[string]operationBehavior{
	// This browser-oriented query frequently fails when sent with integrity
	// headers or with stale persisted-query behavior. Treating errors as fatal
	// prevents silent "0 followers" fallbacks.
	"ChannelFollows": {
		skipIntegrity:   true,
		failOnErrors:    true,
		tryAltClientIDs: true,
	},
	"VideoPlayerStreamInfoOverlayChannel": {
		tryAltClientIDs: true,
		retryGQLErrors:  true,
	},
	"TeamPage": {
		failOnErrors: true,
	},
	// Drop claims fail with "failed integrity check" when the integrity
	// token is stale. Retrying with a fresh token typically succeeds.
	"DropsPage_ClaimDropRewards": {
		retryGQLErrors: true,
	},
}

// circuitBreaker tracks consecutive failures and backs off when the API
type circuitBreaker struct {
	mu               sync.Mutex
	consecutiveFails int
	lastFailure      time.Time
	cooldownUntil    time.Time
}

func (cb *circuitBreaker) recordSuccess() {
	cb.mu.Lock()
	cb.consecutiveFails = 0
	cb.mu.Unlock()
}

// recordFailure increments the failure counter and, after 10 consecutive
func (cb *circuitBreaker) recordFailure() {
	cb.mu.Lock()
	cb.consecutiveFails++
	cb.lastFailure = time.Now()
	if cb.consecutiveFails >= 10 {
		multiplier := cb.consecutiveFails - 9
		if multiplier > 10 {
			multiplier = 10
		}
		backoff := time.Duration(multiplier) * 30 * time.Second
		if backoff > 5*time.Minute {
			backoff = 5 * time.Minute
		}
		cb.cooldownUntil = time.Now().Add(backoff)
	}
	cb.mu.Unlock()
}

// shouldSkip returns true if the circuit breaker is open and requests
func (cb *circuitBreaker) shouldSkip() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return time.Now().Before(cb.cooldownUntil)
}

// Client is the Twitch GQL HTTP client with connection pooling, a circuit
// breaker, and retry logic.
type Client struct {
	httpClient *http.Client
	transport  *http.Transport
	auth       auth.Provider
	log        *logger.Logger
	breaker    *circuitBreaker

	maxRetries int
	mu         sync.RWMutex
}

// NewClientForTest creates a GQL Client with a caller-supplied *http.Client,
// allowing tests to inject a custom transport (e.g. mock round-tripper).
func NewClientForTest(authenticator auth.Provider, log *logger.Logger, httpClient *http.Client) *Client {
	return &Client{
		httpClient: httpClient,
		auth:       authenticator,
		log:        log,
		breaker:    &circuitBreaker{},
		maxRetries: 0, // no retries in tests
	}
}

// NewClient creates a new GQL Client with a shared HTTP client configured
// for connection pooling and the given authenticator.
// An optional proxyURL routes all requests through the specified proxy.
func NewClient(authenticator auth.Provider, log *logger.Logger, proxyURL *url.URL) *Client {
	transport := &http.Transport{
		MaxIdleConns:        20,
		MaxIdleConnsPerHost: 5,
		IdleConnTimeout:     90 * time.Second,
	}
	if proxyURL != nil {
		transport.Proxy = http.ProxyURL(proxyURL)
		log.Info("GQL client using proxy", "proxy", proxyURL.Host)
	}

	httpClient := &http.Client{
		Transport: transport,
		Timeout:   constants.DefaultHTTPTimeout,
	}

	return &Client{
		httpClient: httpClient,
		transport:  transport,
		auth:       authenticator,
		log:        log,
		breaker:    &circuitBreaker{},
		maxRetries: constants.DefaultMaxRetries,
	}
}

// SetStartupMode configures the client for fast startup with reduced
// timeout and retries. Call SetNormalMode to restore defaults.
func (c *Client) SetStartupMode() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.httpClient.Timeout = constants.StartupHTTPTimeout
	c.maxRetries = constants.StartupMaxRetries
	c.log.Debug("GQL client switched to startup mode",
		"timeout", constants.StartupHTTPTimeout,
		"max_retries", constants.StartupMaxRetries)
}

// SetNormalMode restores the client to normal operating mode with
// default timeout and retries.
func (c *Client) SetNormalMode() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.httpClient.Timeout = constants.DefaultHTTPTimeout
	c.maxRetries = constants.DefaultMaxRetries
	c.log.Debug("GQL client switched to normal mode",
		"timeout", constants.DefaultHTTPTimeout,
		"max_retries", constants.DefaultMaxRetries)
}

func (c *Client) getMaxRetries() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.maxRetries
}

// HTTPClient returns the underlying *http.Client for reuse by other packages
// (e.g., minute-watched events that need the same connection pool).
func (c *Client) HTTPClient() *http.Client {
	return c.httpClient
}

type gqlRequest struct {
	OperationName string         `json:"operationName"`
	Variables     map[string]any `json:"variables,omitempty"`
	Extensions    *gqlExtensions `json:"extensions,omitempty"`
	Query         string         `json:"query,omitempty"`
}

type gqlExtensions struct {
	PersistedQuery *persistedQuery `json:"persistedQuery"`
}

type persistedQuery struct {
	Version    int    `json:"version"`
	SHA256Hash string `json:"sha256Hash"`
}

type gqlResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []gqlError      `json:"errors,omitempty"`
}

type gqlError struct {
	Message string `json:"message"`
	Path    []any  `json:"path,omitempty"`
}

// PostGQL sends a single GQL operation and returns the "data" portion of the response.
// It builds the request body, adds auth and version headers, handles errors, and
// retries on transient failures (429, 5xx) with exponential backoff.
func (c *Client) PostGQL(ctx context.Context, op constants.GQLOperation, variables map[string]any) (json.RawMessage, error) {
	reqBody := c.buildRequestBody(op, variables)
	return c.doGQLRequest(ctx, reqBody, op.OperationName)
}

// PostGQLBatch sends multiple GQL operations in a single HTTP request (batch).
// Twitch supports batched GQL requests as a JSON array.
func (c *Client) PostGQLBatch(ctx context.Context, ops []constants.GQLOperation, varsList []map[string]any) ([]json.RawMessage, error) {
	if len(ops) != len(varsList) {
		return nil, fmt.Errorf("ops and varsList must have the same length")
	}

	batch := make([]gqlRequest, len(ops))
	for i, op := range ops {
		batch[i] = c.buildRequestBody(op, varsList[i])
	}

	jsonBody, err := json.Marshal(batch)
	if err != nil {
		return nil, fmt.Errorf("marshaling batch GQL request: %w", err)
	}

	respBody, err := c.doHTTPRequest(ctx, jsonBody, "batch", false, "")
	if err != nil {
		return nil, err
	}

	var responses []gqlResponse
	if err := json.Unmarshal(respBody, &responses); err != nil {
		return nil, fmt.Errorf("parsing batch GQL response: %w", err)
	}

	results := make([]json.RawMessage, len(responses))
	for i, response := range responses {
		if len(response.Errors) > 0 {
			c.log.Warn("GQL batch error",
				"index", i,
				"error", response.Errors[0].Message)
		}
		results[i] = response.Data
	}

	return results, nil
}

func (c *Client) buildRequestBody(op constants.GQLOperation, variables map[string]any) gqlRequest {
	req := gqlRequest{
		OperationName: op.OperationName,
		Variables:     variables,
	}

	if op.Query != "" {
		req.Query = op.Query
	} else {
		req.Extensions = &gqlExtensions{
			PersistedQuery: &persistedQuery{
				Version:    1,
				SHA256Hash: op.SHA256Hash,
			},
		}
	}

	return req
}

func (c *Client) doGQLRequest(ctx context.Context, reqBody gqlRequest, opName string) (json.RawMessage, error) {
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshaling GQL request: %w", err)
	}

	behavior := operationBehaviors[opName]

	// Raw queries (non-persisted) should not send the integrity token —
	// Twitch's integrity system is designed for persisted queries and
	// sending it with raw queries causes "service error" for some categories.
	skipIntegrity := reqBody.Query != "" || behavior.skipIntegrity
	clientIDs := []string{""}
	if behavior.tryAltClientIDs {
		ids := c.auth.ClientIDsForGQL()
		if len(ids) > 0 {
			clientIDs = ids
		}
	}

	var lastErr error
	for idx, clientID := range clientIDs {
		semanticRetries := 0
		for {
			respBody, err := c.doHTTPRequest(ctx, jsonBody, opName, skipIntegrity, clientID)
			if err != nil {
				lastErr = err
				break
			}

			var response gqlResponse
			if err := json.Unmarshal(respBody, &response); err != nil {
				return nil, fmt.Errorf("parsing GQL response for %s: %w", opName, err)
			}

			if len(response.Errors) > 0 {
				errMsg := response.Errors[0].Message
				if strings.Contains(errMsg, "integrity check") && integrityFailureOps[opName] {
					c.log.Debug("GQL integrity check failure (expected)",
						"operation", opName,
						"error", errMsg)
				} else {
					c.log.Warn("GQL operation returned errors",
						"operation", opName,
						"error", errMsg,
						"client_id_attempt", idx+1)
				}

				if behavior.retryGQLErrors && isRetryableGQLError(errMsg) && semanticRetries < 2 {
					// If the error was an integrity check failure, invalidate
					// the cached token so the retry fetches a fresh one.
					if strings.Contains(errMsg, "integrity check") {
						if clearer, ok := c.auth.(interface{ ClearIntegrityToken() }); ok {
							clearer.ClearIntegrityToken()
						}
					}
					semanticRetries++
					backoff := time.Duration(semanticRetries) * time.Second
					c.log.Info("Retrying GQL operation after transient response error",
						"operation", opName,
						"retry", semanticRetries,
						"backoff", backoff)
					select {
					case <-ctx.Done():
						return nil, ctx.Err()
					case <-time.After(backoff):
					}
					continue
				}

				if behavior.tryAltClientIDs && idx < len(clientIDs)-1 {
					c.log.Info("Retrying GQL operation with alternate client ID",
						"operation", opName,
						"attempt", idx+2,
						"total_attempts", len(clientIDs))
					lastErr = wrapTransientGQLError(opName, errMsg)
					break
				}

				if behavior.failOnErrors {
					return nil, wrapTransientGQLError(opName, errMsg)
				}
			}

			return response.Data, nil
		}
	}

	if lastErr != nil {
		return nil, lastErr
	}

	return nil, fmt.Errorf("GQL operation %s failed without a response", opName)
}

func wrapTransientGQLError(opName, errMsg string) error {
	if isRetryableGQLError(errMsg) {
		return fmt.Errorf("%w: GQL operation %s returned error: %s", ErrTransientGQLError, opName, errMsg)
	}
	return fmt.Errorf("GQL operation %s returned error: %s", opName, errMsg)
}

func isRetryableGQLError(errMsg string) bool {
	errMsg = strings.ToLower(errMsg)
	return strings.Contains(errMsg, "service timeout") ||
		strings.Contains(errMsg, "service unavailable") ||
		strings.Contains(errMsg, "temporarily unavailable") ||
		strings.Contains(errMsg, "timed out") ||
		strings.Contains(errMsg, "integrity check")
}

// IsTransientError reports whether the error indicates a temporary Twitch-side
// failure where callers should preserve current state and try again later.
func IsTransientError(err error) bool {
	return errors.Is(err, ErrTransientGQLError) ||
		errors.Is(err, context.DeadlineExceeded) ||
		errors.Is(err, context.Canceled)
}

// doHTTPRequest performs the actual HTTP POST with auth headers, client version,
// integrity token, and retry logic for transient errors. The number of retries
// is controlled by the client's maxRetries setting (configurable via
// SetStartupMode/SetNormalMode).
//
// Retry logging strategy: individual retries are logged at DEBUG level to
// reduce noise. Only the final failure (after all retries exhausted) is
// logged at WARN level. Known-flaky operations (e.g., VideoPlayerStreamInfoOverlayChannel)
func (c *Client) doHTTPRequest(ctx context.Context, jsonBody []byte, opName string, skipIntegrity bool, clientIDOverride string) ([]byte, error) {
	if c.breaker.shouldSkip() {
		c.log.Debug("Circuit breaker open, skipping request", "operation", opName)
		return nil, ErrCircuitOpen
	}

	maxRetries := c.getMaxRetries()

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(math.Pow(2, float64(attempt-1))) * time.Second
			c.log.Debug("Retrying GQL request",
				"operation", opName,
				"attempt", fmt.Sprintf("%d/%d", attempt, maxRetries),
				"backoff", backoff)

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, constants.GQLURL,
			bytes.NewReader(jsonBody))
		if err != nil {
			return nil, fmt.Errorf("creating GQL request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		for k, v := range c.auth.GetAuthHeaders() {
			req.Header.Set(k, v)
		}
		if clientIDOverride != "" {
			req.Header.Set("Client-Id", clientIDOverride)
		}
		req.Header.Set("Client-Version", c.auth.ClientVersion())

		if !skipIntegrity {
			if integrityToken, err := c.auth.FetchIntegrityToken(ctx); err != nil {
				c.log.Debug("Failed to fetch integrity token, proceeding without it",
					"operation", opName, "error", err)
			} else if integrityToken != "" {
				req.Header.Set("Client-Integrity", integrityToken)
			}
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			if attempt < maxRetries {
				c.log.Debug("GQL request failed, will retry",
					"operation", opName,
					"attempt", fmt.Sprintf("%d/%d", attempt+1, maxRetries),
					"error", err)
				continue
			}
			c.log.Warn("GQL request failed after all retries",
				"operation", opName,
				"attempts", maxRetries+1,
				"error", err)
			return nil, fmt.Errorf("GQL request for %s failed: %w", opName, err)
		}

		body, readErr := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		resp.Body.Close()

		if readErr != nil {
			if attempt < maxRetries {
				c.log.Debug("Failed to read GQL response, will retry",
					"operation", opName,
					"attempt", fmt.Sprintf("%d/%d", attempt+1, maxRetries),
					"error", readErr)
				continue
			}
			return nil, fmt.Errorf("reading GQL response for %s: %w", opName, readErr)
		}

		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
			if attempt < maxRetries {
				c.log.Debug("GQL request returned retryable status, will retry",
					"operation", opName,
					"status", resp.StatusCode,
					"attempt", fmt.Sprintf("%d/%d", attempt+1, maxRetries))
				continue
			}
			c.log.Warn("GQL request returned retryable status after all retries",
				"operation", opName,
				"status", resp.StatusCode,
				"attempts", maxRetries+1)
			return nil, fmt.Errorf("GQL request for %s returned status %d after %d retries",
				opName, resp.StatusCode, maxRetries)
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("GQL request for %s returned status %d: %s",
				opName, resp.StatusCode, string(body))
		}

		c.breaker.recordSuccess()
		c.log.Debug("GQL request completed",
			"operation", opName,
			"status", resp.StatusCode)

		return body, nil
	}

	c.breaker.recordFailure()
	return nil, fmt.Errorf("GQL request for %s exhausted retries", opName)
}
