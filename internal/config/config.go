// Package config handles loading, parsing, and validating YAML configuration
// files for the Twitch miner. It supports per-account configuration with
// environment variable overrides for secrets.
package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Guliveer/twitch-miner-go/internal/constants"
	"gopkg.in/yaml.v3"
)

// parseProxyURL validates and parses a proxy URL string.
func parseProxyURL(raw string) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}
	if u.Scheme != "http" && u.Scheme != "https" && u.Scheme != "socks5" {
		return nil, fmt.Errorf("unsupported proxy scheme %q (use http, https, or socks5)", u.Scheme)
	}
	if u.Host == "" {
		return nil, fmt.Errorf("proxy URL missing host")
	}
	return u, nil
}

// ProxyURL returns the parsed proxy URL, or nil if no proxy is configured.
func (ac *AccountConfig) ProxyURL() *url.URL {
	if ac.Proxy == "" {
		return nil
	}
	u, _ := parseProxyURL(ac.Proxy)
	return u
}

// DefaultConfigDir is the default directory for account configuration files.
const DefaultConfigDir = "configs"

// LoadAccountConfig loads a single account configuration from a YAML file,
// then overlays environment variables for secrets.
func LoadAccountConfig(path string) (*AccountConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file %s: %w", path, err)
	}

	var cfg AccountConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file %s: %w", path, err)
	}

	filename := filepath.Base(path)
	ext := filepath.Ext(filename)
	cfg.Username = strings.TrimSuffix(filename, ext)

	applyDefaults(&cfg)
	applyEnvOverrides(&cfg)

	return &cfg, nil
}

// LoadAllAccountConfigs loads all .yaml/.yml files from the given directory.
// Each file is expected to contain a single AccountConfig.
// Only files ending in .yaml or .yml are loaded; everything else (including
// .yaml.example) is ignored by the extension check.
// The username for each account is derived from the config filename.
func LoadAllAccountConfigs(dir string) ([]*AccountConfig, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading config directory %s: %w", dir, err)
	}

	var configs []*AccountConfig
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		ext := filepath.Ext(name)
		if ext != ".yaml" && ext != ".yml" {
			continue
		}

		cfg, err := LoadAccountConfig(filepath.Join(dir, name))
		if err != nil {
			return nil, fmt.Errorf("loading %s: %w", name, err)
		}

		configs = append(configs, cfg)
	}

	if len(configs) == 0 {
		return nil, fmt.Errorf("no account config files found in %s", dir)
	}

	return configs, nil
}

func applyDefaults(cfg *AccountConfig) {
	if cfg.MaxWatchStreams <= 0 {
		cfg.MaxWatchStreams = constants.MaxWatchStreams
	}

	if len(cfg.Priority) == 0 {
		cfg.Priority = []string{"STREAK", "DROPS", "ORDER"}
	}

	if cfg.CategoryWatcher.PollInterval == 0 {
		cfg.CategoryWatcher.PollInterval = 120 * time.Second
	}

	if cfg.TeamWatcher.PollInterval == 0 {
		cfg.TeamWatcher.PollInterval = 120 * time.Second
	}

	if cfg.Followers.Order == "" {
		cfg.Followers.Order = "ASC"
	}
}

// getEnv looks up an environment variable with a per-account suffix.
// Falls back to the global key (without suffix) if the per-account one is not set.
func getEnv(key, username string) string {
	if v := os.Getenv(key + "_" + strings.ToUpper(username)); v != "" {
		return v
	}
	return os.Getenv(key)
}

// applyEnvOverrides overlays environment variables for secrets.
// Each variable is first looked up with the per-account suffix KEY_<UPPERCASE_USERNAME>,
// then falls back to the global key KEY (without suffix).
func applyEnvOverrides(cfg *AccountConfig) {
	username := cfg.Username

	if cfg.Notifications.Telegram != nil {
		if envValue := getEnv("TELEGRAM_TOKEN", username); envValue != "" {
			cfg.Notifications.Telegram.Token = envValue
		}
		if envValue := getEnv("TELEGRAM_CHAT_ID", username); envValue != "" {
			cfg.Notifications.Telegram.ChatID = envValue
		}
	}

	if cfg.Notifications.Discord != nil {
		if envValue := getEnv("DISCORD_WEBHOOK", username); envValue != "" {
			cfg.Notifications.Discord.WebhookURL = envValue
		}
	}

	if cfg.Notifications.Webhook != nil {
		if envValue := getEnv("WEBHOOK_URL", username); envValue != "" {
			cfg.Notifications.Webhook.Endpoint = envValue
		}
	}

	if cfg.Notifications.Matrix != nil {
		if envValue := getEnv("MATRIX_HOMESERVER", username); envValue != "" {
			cfg.Notifications.Matrix.Homeserver = envValue
		}
		if envValue := getEnv("MATRIX_ROOM_ID", username); envValue != "" {
			cfg.Notifications.Matrix.RoomID = envValue
		}
		if envValue := getEnv("MATRIX_ACCESS_TOKEN", username); envValue != "" {
			cfg.Notifications.Matrix.AccessToken = envValue
		}
	}

	if cfg.Notifications.Pushover != nil {
		if envValue := getEnv("PUSHOVER_TOKEN", username); envValue != "" {
			cfg.Notifications.Pushover.APIToken = envValue
		}
		if envValue := getEnv("PUSHOVER_USER_KEY", username); envValue != "" {
			cfg.Notifications.Pushover.UserKey = envValue
		}
	}

	if cfg.Notifications.Gotify != nil {
		if envValue := getEnv("GOTIFY_URL", username); envValue != "" {
			cfg.Notifications.Gotify.URL = envValue
		}
		if envValue := getEnv("GOTIFY_TOKEN", username); envValue != "" {
			cfg.Notifications.Gotify.Token = envValue
		}
	}
}

// Validate checks the configuration for common errors and contradictory settings.
func Validate(cfg *AccountConfig) error {
	if cfg.Username == "" {
		return fmt.Errorf("username is required")
	}

	if cfg.MaxWatchStreams < 1 {
		return fmt.Errorf("account %s: max_watch_streams must be at least 1", cfg.Username)
	}

	if len(cfg.Streamers) == 0 && !cfg.Followers.Enabled && !cfg.CategoryWatcher.Enabled && !cfg.TeamWatcher.Enabled {
		return fmt.Errorf("account %s: at least one of streamers, followers, category_watcher, or team_watcher must be configured", cfg.Username)
	}

	for i, streamerCfg := range cfg.Streamers {
		if streamerCfg.Username == "" {
			return fmt.Errorf("account %s: streamer at index %d has empty username", cfg.Username, i)
		}
	}

	if cfg.Notifications.Telegram != nil && cfg.Notifications.Telegram.Enabled {
		if cfg.Notifications.Telegram.Token == "" || cfg.Notifications.Telegram.ChatID == "" {
			u := strings.ToUpper(cfg.Username)
			return fmt.Errorf("account %s: telegram enabled but token or chat_id not set (use env vars TELEGRAM_TOKEN_%s/TELEGRAM_CHAT_ID_%s or global TELEGRAM_TOKEN/TELEGRAM_CHAT_ID)", cfg.Username, u, u)
		}
	}

	if cfg.Notifications.Discord != nil && cfg.Notifications.Discord.Enabled {
		if cfg.Notifications.Discord.WebhookURL == "" {
			return fmt.Errorf("account %s: discord enabled but webhook_url not set (use env var DISCORD_WEBHOOK_%s or global DISCORD_WEBHOOK)", cfg.Username, strings.ToUpper(cfg.Username))
		}
	}

	// Detect contradictory settings: predictions enabled but no bet config.
	if cfg.StreamerDefaults.MakePredictions != nil && *cfg.StreamerDefaults.MakePredictions && cfg.StreamerDefaults.Bet == nil {
		return fmt.Errorf("account %s: make_predictions is enabled in streamer_defaults but no bet config is set", cfg.Username)
	}

	// Category watcher enabled but no categories configured.
	if cfg.CategoryWatcher.Enabled && len(cfg.CategoryWatcher.Categories) == 0 {
		return fmt.Errorf("account %s: category_watcher is enabled but no categories are configured", cfg.Username)
	}

	// Team watcher enabled but no teams configured.
	if cfg.TeamWatcher.Enabled && len(cfg.TeamWatcher.Teams) == 0 {
		return fmt.Errorf("account %s: team_watcher is enabled but no teams are configured", cfg.Username)
	}

	// Batch interval must be positive when enabled.
	if cfg.Notifications.Batch != nil && cfg.Notifications.Batch.IsBatchEnabled() {
		if cfg.Notifications.Batch.Interval < 0 {
			return fmt.Errorf("account %s: batch interval must be positive", cfg.Username)
		}
	}

	// Proxy URL must be well-formed if set.
	if cfg.Proxy != "" {
		if _, err := parseProxyURL(cfg.Proxy); err != nil {
			return fmt.Errorf("account %s: invalid proxy URL %q: %w", cfg.Username, cfg.Proxy, err)
		}
	}

	return nil
}
