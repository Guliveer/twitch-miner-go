package runtimecfg

import (
	"fmt"
	"os"
	"strings"
)

const (
	envTwitchClientIDTV      = "TWITCH_CLIENT_ID_TV"
	envTwitchClientIDBrowser = "TWITCH_CLIENT_ID_BROWSER"
	envTwitchClientIDMobile  = "TWITCH_CLIENT_ID_MOBILE"
	envTwitchClientIDAndroid = "TWITCH_CLIENT_ID_ANDROID"
	envTwitchClientIDIOS     = "TWITCH_CLIENT_ID_IOS"
	envTwitchClientVersion   = "TWITCH_CLIENT_VERSION"
)

// Twitch holds Twitch runtime identifiers that must be supplied by the
// deployment environment instead of being hardcoded in the binary.
type Twitch struct {
	ClientIDTV      string
	ClientIDBrowser string
	ClientIDMobile  string
	ClientIDAndroid string
	ClientIDIOS     string
	ClientVersion   string
}

// LoadTwitchFromEnv loads all required Twitch identifiers from environment
// variables and returns a validation error when any value is missing.
func LoadTwitchFromEnv() (*Twitch, error) {
	cfg := &Twitch{
		ClientIDTV:      strings.TrimSpace(os.Getenv(envTwitchClientIDTV)),
		ClientIDBrowser: strings.TrimSpace(os.Getenv(envTwitchClientIDBrowser)),
		ClientIDMobile:  strings.TrimSpace(os.Getenv(envTwitchClientIDMobile)),
		ClientIDAndroid: strings.TrimSpace(os.Getenv(envTwitchClientIDAndroid)),
		ClientIDIOS:     strings.TrimSpace(os.Getenv(envTwitchClientIDIOS)),
		ClientVersion:   strings.TrimSpace(os.Getenv(envTwitchClientVersion)),
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate ensures all required Twitch identifiers are present.
func (c *Twitch) Validate() error {
	if c == nil {
		return fmt.Errorf("twitch runtime config is required")
	}

	required := map[string]string{
		envTwitchClientIDTV:      c.ClientIDTV,
		envTwitchClientIDBrowser: c.ClientIDBrowser,
		envTwitchClientVersion:   c.ClientVersion,
	}

	var missing []string
	for key, value := range required {
		if value == "" {
			missing = append(missing, key)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required Twitch runtime environment variables: %s", strings.Join(missing, ", "))
	}

	return nil
}
