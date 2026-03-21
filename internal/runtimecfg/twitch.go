package runtimecfg

import (
	"log/slog"
	"os"
	"strings"

	"github.com/Guliveer/twitch-miner-go/internal/constants"
)

const (
	envTwitchClientIDTV      = "TWITCH_CLIENT_ID_TV"
	envTwitchClientIDBrowser = "TWITCH_CLIENT_ID_BROWSER"
	envTwitchClientIDMobile  = "TWITCH_CLIENT_ID_MOBILE"
	envTwitchClientIDAndroid = "TWITCH_CLIENT_ID_ANDROID"
	envTwitchClientIDIOS     = "TWITCH_CLIENT_ID_IOS"
	envTwitchClientVersion   = "TWITCH_CLIENT_VERSION"
)

// Twitch holds Twitch runtime identifiers loaded from environment variables
// with built-in defaults from constants. Environment variables take priority.
type Twitch struct {
	ClientIDTV      string
	ClientIDBrowser string
	ClientIDMobile  string
	ClientIDAndroid string
	ClientIDIOS     string
	ClientVersion   string
}

// ClientIDsForGQL returns a de-duplicated list of client IDs ordered from the
// most browser-like context to the most device-like context. This is useful for
// trying alternative Twitch client contexts when an internal operation becomes
// client-sensitive.
func (c *Twitch) ClientIDsForGQL() []string {
	if c == nil {
		return nil
	}

	candidates := []string{
		c.ClientIDBrowser,
		c.ClientIDMobile,
		c.ClientIDAndroid,
		c.ClientIDIOS,
		c.ClientIDTV,
	}

	seen := make(map[string]struct{}, len(candidates))
	ids := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		if _, ok := seen[candidate]; ok {
			continue
		}
		seen[candidate] = struct{}{}
		ids = append(ids, candidate)
	}

	return ids
}

// envOrDefault returns the trimmed environment variable value, or the fallback
// default if the variable is not set or empty.
func envOrDefault(envKey, fallback string) (value string, fromEnv bool) {
	v := strings.TrimSpace(os.Getenv(envKey))
	if v != "" {
		return v, true
	}
	return fallback, false
}

// LoadTwitchFromEnv loads Twitch identifiers from environment variables,
// falling back to built-in defaults from constants. Logs a warning when
// defaults are used so operators know to configure fresh values.
func LoadTwitchFromEnv(log *slog.Logger) *Twitch {
	var usedDefaults []string

	load := func(envKey, fallback string) string {
		v, fromEnv := envOrDefault(envKey, fallback)
		if !fromEnv && fallback != "" {
			usedDefaults = append(usedDefaults, envKey)
		}
		return v
	}

	cfg := &Twitch{
		ClientIDTV:      load(envTwitchClientIDTV, constants.ClientID),
		ClientIDBrowser: load(envTwitchClientIDBrowser, constants.ClientIDBrowser),
		ClientIDMobile:  load(envTwitchClientIDMobile, constants.ClientIDMobile),
		ClientIDAndroid: load(envTwitchClientIDAndroid, constants.ClientIDAndroid),
		ClientIDIOS:     load(envTwitchClientIDIOS, constants.ClientIDiOS),
		ClientVersion:   load(envTwitchClientVersion, constants.ClientVersion),
	}

	if len(usedDefaults) > 0 {
		log.Warn("Using built-in default Twitch client identifiers (may be outdated). "+
			"Set environment variables for fresh values.",
			"defaults_used", strings.Join(usedDefaults, ", "))
	}

	return cfg
}
