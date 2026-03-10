package runtimecfg

import (
	"strings"
	"testing"
)

func TestTwitchValidateMissingValues(t *testing.T) {
	t.Parallel()

	cfg := &Twitch{
		ClientIDTV:      "tv",
		ClientIDBrowser: "",
		ClientIDMobile:  "mobile",
		ClientIDAndroid: "",
		ClientIDIOS:     "ios",
		ClientVersion:   "",
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected validation error for missing values")
	}

	for _, key := range []string{
		"TWITCH_CLIENT_ID_BROWSER",
		"TWITCH_CLIENT_VERSION",
	} {
		if !strings.Contains(err.Error(), key) {
			t.Fatalf("expected error to mention %s, got %q", key, err.Error())
		}
	}
}

func TestTwitchValidateSuccess(t *testing.T) {
	t.Parallel()

	cfg := &Twitch{
		ClientIDTV:      "tv",
		ClientIDBrowser: "browser",
		ClientIDMobile:  "mobile",
		ClientIDAndroid: "android",
		ClientIDIOS:     "ios",
		ClientVersion:   "version",
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected config to validate, got %v", err)
	}
}
