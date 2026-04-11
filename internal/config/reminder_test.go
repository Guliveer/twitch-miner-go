package config

import (
	"testing"
	"time"
)

func TestParseReminderDuration_Days(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected time.Duration
	}{
		{"1d", 24 * time.Hour},
		{"3d", 72 * time.Hour},
		{"7d", 168 * time.Hour},
	}

	for _, tt := range tests {
		got, err := parseReminderDuration(tt.input)
		if err != nil {
			t.Fatalf("parseReminderDuration(%q): unexpected error: %v", tt.input, err)
		}
		if got != tt.expected {
			t.Fatalf("parseReminderDuration(%q) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}

func TestParseReminderDuration_Standard(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected time.Duration
	}{
		{"15m", 15 * time.Minute},
		{"1h", time.Hour},
		{"30s", 30 * time.Second},
	}

	for _, tt := range tests {
		got, err := parseReminderDuration(tt.input)
		if err != nil {
			t.Fatalf("parseReminderDuration(%q): unexpected error: %v", tt.input, err)
		}
		if got != tt.expected {
			t.Fatalf("parseReminderDuration(%q) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}

func TestParseReminderDuration_Invalid(t *testing.T) {
	t.Parallel()

	for _, input := range []string{"abc", "xd", ""} {
		_, err := parseReminderDuration(input)
		if err == nil {
			t.Fatalf("parseReminderDuration(%q): expected error, got nil", input)
		}
	}
}

func TestParseCampaignReminders_Full(t *testing.T) {
	t.Parallel()

	cfg := ParseCampaignReminders([]string{"on_detection", "3d", "1d", "15m"})
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if !cfg.OnDetection {
		t.Fatal("expected OnDetection to be true")
	}
	if len(cfg.Durations) != 3 {
		t.Fatalf("expected 3 durations, got %d", len(cfg.Durations))
	}
	// Should be sorted descending.
	if cfg.Durations[0] != 72*time.Hour {
		t.Fatalf("expected first duration 72h (3d), got %v", cfg.Durations[0])
	}
	if cfg.Durations[1] != 24*time.Hour {
		t.Fatalf("expected second duration 24h (1d), got %v", cfg.Durations[1])
	}
	if cfg.Durations[2] != 15*time.Minute {
		t.Fatalf("expected third duration 15m, got %v", cfg.Durations[2])
	}
}

func TestParseCampaignReminders_OnDetectionOnly(t *testing.T) {
	t.Parallel()

	cfg := ParseCampaignReminders([]string{"on_detection"})
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if !cfg.OnDetection {
		t.Fatal("expected OnDetection to be true")
	}
	if len(cfg.Durations) != 0 {
		t.Fatalf("expected 0 durations, got %d", len(cfg.Durations))
	}
}

func TestParseCampaignReminders_DurationsOnly(t *testing.T) {
	t.Parallel()

	cfg := ParseCampaignReminders([]string{"1d"})
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if cfg.OnDetection {
		t.Fatal("expected OnDetection to be false")
	}
	if len(cfg.Durations) != 1 || cfg.Durations[0] != 24*time.Hour {
		t.Fatalf("expected [24h], got %v", cfg.Durations)
	}
}

func TestParseCampaignReminders_Empty(t *testing.T) {
	t.Parallel()

	cfg := ParseCampaignReminders([]string{})
	if cfg != nil {
		t.Fatal("expected nil config for empty input")
	}
}

func TestParseCampaignReminders_InvalidEntriesIgnored(t *testing.T) {
	t.Parallel()

	cfg := ParseCampaignReminders([]string{"on_detection", "invalid", "1d"})
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if !cfg.OnDetection {
		t.Fatal("expected OnDetection to be true")
	}
	if len(cfg.Durations) != 1 {
		t.Fatalf("expected 1 valid duration (invalid skipped), got %d", len(cfg.Durations))
	}
}

func TestEffectiveCampaignReminders_GlobalDefault(t *testing.T) {
	t.Parallel()

	cwc := CategoryWatcherConfig{
		CampaignReminders: []string{"on_detection", "1d"},
	}
	cat := CategoryConfig{Slug: "test"}

	cfg := cwc.EffectiveCampaignReminders(cat)
	if cfg == nil {
		t.Fatal("expected non-nil config from global default")
	}
	if !cfg.OnDetection {
		t.Fatal("expected OnDetection from global")
	}
	if len(cfg.Durations) != 1 {
		t.Fatalf("expected 1 duration from global, got %d", len(cfg.Durations))
	}
}

func TestEffectiveCampaignReminders_PerCategoryOverride(t *testing.T) {
	t.Parallel()

	cwc := CategoryWatcherConfig{
		CampaignReminders: []string{"on_detection", "1d"},
	}
	cat := CategoryConfig{
		Slug:              "test",
		CampaignReminders: []string{"3d", "15m"},
	}

	cfg := cwc.EffectiveCampaignReminders(cat)
	if cfg == nil {
		t.Fatal("expected non-nil config from per-category override")
	}
	if cfg.OnDetection {
		t.Fatal("expected OnDetection false (per-category has no on_detection)")
	}
	if len(cfg.Durations) != 2 {
		t.Fatalf("expected 2 durations from per-category, got %d", len(cfg.Durations))
	}
}

func TestEffectiveCampaignReminders_NeitherSet(t *testing.T) {
	t.Parallel()

	cwc := CategoryWatcherConfig{}
	cat := CategoryConfig{Slug: "test"}

	cfg := cwc.EffectiveCampaignReminders(cat)
	if cfg != nil {
		t.Fatal("expected nil config when neither global nor per-category set")
	}
}
