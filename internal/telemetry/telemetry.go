// Package telemetry sends periodic heartbeats to a telemetry server to track
// active instances, versions, and platforms.
//
// Telemetry is enabled by default. Set TELEMETRY_AGREE=false to opt out.
// The telemetry URL is compiled into the binary and can be overridden via
// TELEMETRY_URL env var (useful for forks running their own server).
//
// Heartbeat payload is anonymous: instance_id (random UUID), version, os, arch,
// deployment label, and running accounts count.
// No personal data, no channel names, no IPs.
package telemetry

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// defaultTelemetryURL is the default telemetry server URL compiled into the
// binary. Override via TELEMETRY_URL env var.
const defaultTelemetryURL = "https://twitch-miner-go-telemetry.vercel.app"

const (
	envTelemetryAgree  = "TELEMETRY_AGREE"
	envTelemetryURL    = "TELEMETRY_URL"
	envHeartbeatAPIKey = "HEARTBEAT_API_KEY"
	envDataDir         = "DATA_DIR"
	envDeployment      = "DEPLOYMENT"
	envInterval        = "TELEMETRY_INTERVAL"

	defaultInterval    = 10 * time.Minute
	instanceIDFilename = ".instance_id"
)

// Config holds the configuration for the heartbeat sender.
type Config struct {
	TelemetryURL string
	APIKey       string
	InstanceID   string
	Version      string
	Interval     time.Duration
}

// heartbeatPayload is the JSON body sent to the telemetry server.
type heartbeatPayload struct {
	InstanceID      string `json:"instance_id"`
	Version         string `json:"version"`
	OS              string `json:"os"`
	Arch            string `json:"arch"`
	Deployment      string `json:"deployment"`
	RunningAccounts int    `json:"running_accounts"`
	UptimeSeconds   int    `json:"uptime_seconds"`
}

// LoadConfigFromEnv reads heartbeat configuration from environment variables.
// Telemetry is enabled by default unless TELEMETRY_AGREE=false is set.
// The default telemetry URL is compiled into the binary; override via TELEMETRY_URL.
func LoadConfigFromEnv(log *slog.Logger) (*Config, error) {
	if os.Getenv(envTelemetryAgree) == "false" {
		return nil, nil
	}

	url := os.Getenv(envTelemetryURL)
	if url == "" {
		url = defaultTelemetryURL
	}
	url = strings.TrimRight(url, "/")

	apiKey := os.Getenv(envHeartbeatAPIKey)

	dataDir := os.Getenv(envDataDir)
	if dataDir == "" {
		dataDir = "."
	}

	instanceID, err := loadOrGenerateInstanceID(dataDir)
	if err != nil {
		return nil, fmt.Errorf("telemetry: %w", err)
	}

	interval := defaultInterval
	if s := os.Getenv(envInterval); s != "" {
		if d, err := time.ParseDuration(s); err == nil && d > 0 {
			interval = d
		} else {
			log.Warn("Telemetry: invalid interval, using default", "value", s, "fallback", defaultInterval)
		}
	}

	return &Config{
		TelemetryURL: url,
		APIKey:       apiKey,
		InstanceID:   instanceID,
		Version:      "dev",
		Interval:     interval,
	}, nil
}

// Sender sends periodic heartbeats to a telemetry server.
type Sender struct {
	cfg              *Config
	client           *http.Client
	log              *slog.Logger
	runningAccounts  func() int
	processStartTime time.Time
}

// SetRunningAccountsFunc sets a callback that returns the current number
// of running miner accounts. Called on each heartbeat.
func (s *Sender) SetRunningAccountsFunc(fn func() int) {
	s.runningAccounts = fn
}

// NewSender creates a new heartbeat sender.
func NewSender(cfg *Config, log *slog.Logger) *Sender {
	return &Sender{
		cfg:              cfg,
		client:           &http.Client{Timeout: 10 * time.Second},
		log:              log,
		processStartTime: time.Now(),
	}
}

// Run starts the heartbeat loop. It waits for ready (if non-nil) to be
// closed, then waits until at least one account is running before sending
// the first heartbeat. Repeats at the configured interval. Blocks until
// ctx is cancelled.
func (s *Sender) Run(ctx context.Context, ready <-chan struct{}) {
	if s.cfg == nil {
		return
	}

	if ready != nil {
		select {
		case <-ready:
		case <-ctx.Done():
			return
		}
	}

	s.waitForFirstRunningAccount(ctx, 5*time.Minute)

	s.sendHeartbeat(ctx)

	ticker := time.NewTicker(s.cfg.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.sendHeartbeat(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (s *Sender) waitForFirstRunningAccount(ctx context.Context, timeout time.Duration) {
	if s.runningAccounts == nil {
		return
	}

	deadline := time.After(timeout)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		if s.runningAccounts() > 0 {
			return
		}
		select {
		case <-ctx.Done():
			return
		case <-deadline:
			s.log.Warn("Telemetry: no running accounts after timeout, sending heartbeat anyway")
			return
		case <-ticker.C:
		}
	}
}

func (s *Sender) sendHeartbeat(ctx context.Context) {
	runningAccounts := 0
	if s.runningAccounts != nil {
		runningAccounts = s.runningAccounts()
	}

	payload := heartbeatPayload{
		InstanceID:      s.cfg.InstanceID,
		Version:         s.cfg.Version,
		OS:              runtime.GOOS,
		Arch:            runtime.GOARCH,
		Deployment:      detectDeployment(),
		RunningAccounts: runningAccounts,
		UptimeSeconds:   int(time.Since(s.processStartTime).Seconds()),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		s.log.Warn("Telemetry: failed to marshal heartbeat", "error", err)
		return
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		s.cfg.TelemetryURL+"/api/heartbeat", bytes.NewReader(body))
	if err != nil {
		s.log.Warn("Telemetry: failed to create request", "error", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	if s.cfg.APIKey != "" {
		req.Header.Set("X-API-Key", s.cfg.APIKey)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		s.log.Warn("Telemetry: heartbeat failed", "error", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		s.log.Warn("Telemetry: unexpected response", "status", resp.StatusCode)
		return
	}

	s.log.Debug("Telemetry: heartbeat sent",
		"version", payload.Version,
		"deployment", payload.Deployment,
		"os", payload.OS,
		"arch", payload.Arch,
		"running_accounts", payload.RunningAccounts,
		"uptime_seconds", payload.UptimeSeconds,
	)
}

// loadOrGenerateInstanceID reads the instance ID from DATA_DIR/.instance_id,
// or generates a UUID v4 and persists it.
func loadOrGenerateInstanceID(dataDir string) (string, error) {
	path := filepath.Join(dataDir, instanceIDFilename)

	if data, err := os.ReadFile(path); err == nil {
		id := strings.TrimSpace(string(data))
		if id != "" {
			return id, nil
		}
	}

	uuid, err := newUUID()
	if err != nil {
		return "", fmt.Errorf("generate instance id: %w", err)
	}

	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return "", fmt.Errorf("create data dir %q: %w", dataDir, err)
	}

	// 0o444 keeps the file read-only from the moment it is created. The
	// leading dot in the filename already hides it on Unix; on Windows the
	// HIDDEN attribute is applied explicitly (see applyInstanceFileAttrs).
	if err := os.WriteFile(path, []byte(uuid+"\n"), 0o444); err != nil {
		return "", fmt.Errorf("write instance id: %w", err)
	}
	if err := applyInstanceFileAttrs(path); err != nil {
		return "", fmt.Errorf("apply instance id attributes: %w", err)
	}

	return uuid, nil
}

// newUUID generates a RFC 4122 v4 UUID using crypto/rand.
func newUUID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	// Set version 4 bits.
	b[6] = (b[6] & 0x0f) | 0x40
	// Set variant bits (RFC 4122).
	b[8] = (b[8] & 0x3f) | 0x80

	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}

// detectDeployment returns a deployment label based on environment hints.
func detectDeployment() string {
	switch {
	case os.Getenv(envDeployment) != "":
		return os.Getenv(envDeployment)
	case os.Getenv("FLY_APP_NAME") != "":
		return "fly-io"
	case os.Getenv("DYNO") != "":
		return "heroku"
	case os.Getenv("K_SERVICE") != "" || os.Getenv("CLOUD_RUN_JOB") != "":
		return "cloud-run"
	case os.Getenv("RAILWAY_ENVIRONMENT") != "":
		return "railway"
	case os.Getenv("RENDER") != "":
		return "render"
	case os.Getenv("KUBERNETES_SERVICE_HOST") != "":
		return "kubernetes"
	case os.Getenv("WEBSITE_SITE_NAME") != "":
		return "azure"
	case os.Getenv("ECS_CONTAINER_METADATA_URI_V4") != "":
		return "aws-ecs"
	case os.Getenv("KOYEB_APP_NAME") != "":
		return "koyeb"
	case os.Getenv("CYCLIC_URL") != "":
		return "cyclic"
	default:
		if _, err := os.Stat("/.dockerenv"); err == nil {
			return "docker"
		}
		return "self-hosted"
	}
}
