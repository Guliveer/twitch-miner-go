package config

import (
	"testing"
	"time"
)

func boolPtr(b bool) *bool { return &b }

func TestResolveBatchConfig_BothNil(t *testing.T) {
	if ResolveBatchConfig(nil, nil) != nil {
		t.Error("expected nil when both global and provider are nil")
	}
}

func TestResolveBatchConfig_GlobalOnly(t *testing.T) {
	global := &BatchConfig{
		Enabled:         boolPtr(true),
		Interval:        15 * time.Minute,
		MaxEntries:      10,
		ImmediateEvents: []string{"BET_WIN"},
	}

	resolved := ResolveBatchConfig(global, nil)
	if !resolved.IsBatchEnabled() {
		t.Error("expected batch enabled")
	}
	if resolved.Interval != 15*time.Minute {
		t.Errorf("expected 15m interval, got %v", resolved.Interval)
	}
	if resolved.MaxEntries != 10 {
		t.Errorf("expected 10 max entries, got %d", resolved.MaxEntries)
	}
	if len(resolved.ImmediateEvents) != 1 || resolved.ImmediateEvents[0] != "BET_WIN" {
		t.Error("expected immediate events from global")
	}
}

func TestResolveBatchConfig_ProviderOverridesGlobal(t *testing.T) {
	global := &BatchConfig{
		Enabled:         boolPtr(true),
		Interval:        15 * time.Minute,
		MaxEntries:      10,
		ImmediateEvents: []string{"BET_WIN"},
	}
	provider := &BatchConfig{
		Interval:        30 * time.Minute,
		ImmediateEvents: []string{"DROP_CLAIM", "BET_LOSE"},
	}

	resolved := ResolveBatchConfig(global, provider)
	if !resolved.IsBatchEnabled() {
		t.Error("expected enabled from global")
	}
	if resolved.Interval != 30*time.Minute {
		t.Errorf("expected 30m override, got %v", resolved.Interval)
	}
	if resolved.MaxEntries != 10 {
		t.Errorf("expected 10 from global, got %d", resolved.MaxEntries)
	}
	if len(resolved.ImmediateEvents) != 2 {
		t.Errorf("expected 2 immediate events from provider, got %d", len(resolved.ImmediateEvents))
	}
}

func TestResolveBatchConfig_ProviderDisablesGlobal(t *testing.T) {
	global := &BatchConfig{
		Enabled:  boolPtr(true),
		Interval: 15 * time.Minute,
	}
	provider := &BatchConfig{
		Enabled: boolPtr(false),
	}

	resolved := ResolveBatchConfig(global, provider)
	if resolved.IsBatchEnabled() {
		t.Error("expected provider to disable batching")
	}
}

func TestIsBatchEnabled_Nil(t *testing.T) {
	var bc *BatchConfig
	if bc.IsBatchEnabled() {
		t.Error("nil config should not be enabled")
	}
}
