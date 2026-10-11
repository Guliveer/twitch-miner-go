package config

import (
	"testing"

	"github.com/Guliveer/twitch-miner-go/internal/model"
)

func TestBuildSchemaPayload_ContainsAllEvents(t *testing.T) {
	payload := BuildSchemaPayload()
	events := toStringSet(payload["notification_events"])
	for _, e := range model.AllEvents() {
		if !events[e.String()] {
			t.Errorf("notification_events missing %q", e.String())
		}
	}
}

func TestBuildSchemaPayload_ContainsAllStrategies(t *testing.T) {
	payload := BuildSchemaPayload()
	strategies := toStringSet(payload["strategies"])
	for _, s := range model.AllStrategies() {
		if !strategies[s.String()] {
			t.Errorf("strategies missing %q", s.String())
		}
	}
}

func TestBuildSchemaPayload_ContainsAllPriorities(t *testing.T) {
	payload := BuildSchemaPayload()
	priorities := toStringSet(payload["priorities"])
	for _, p := range model.AllPriorities() {
		if !priorities[p.String()] {
			t.Errorf("priorities missing %q", p.String())
		}
	}
}

func TestBuildSchemaPayload_ContainsAllDelayModes(t *testing.T) {
	payload := BuildSchemaPayload()
	modes := toStringSet(payload["delay_modes"])
	for _, m := range model.AllDelayModes() {
		if !modes[m.String()] {
			t.Errorf("delay_modes missing %q", m.String())
		}
	}
}

func TestBuildSchemaPayload_ContainsAllChatModes(t *testing.T) {
	payload := BuildSchemaPayload()
	modes := toStringSet(payload["chat_modes"])
	for _, m := range model.AllChatModes() {
		if !modes[m.String()] {
			t.Errorf("chat_modes missing %q", m.String())
		}
	}
}

func TestBuildSchemaPayload_ContainsAllFollowersOrders(t *testing.T) {
	payload := BuildSchemaPayload()
	orders := toStringSet(payload["followers_order"])
	for _, o := range model.AllFollowersOrders() {
		if !orders[o.String()] {
			t.Errorf("followers_order missing %q", o.String())
		}
	}
}

func TestBuildSchemaPayload_DefaultsNonEmpty(t *testing.T) {
	payload := BuildSchemaPayload()
	defaults, ok := payload["defaults"].(map[string]any)
	if !ok {
		t.Fatal("defaults is not a map")
	}
	if len(defaults) < 3 {
		t.Errorf("defaults has %d keys, want at least 3", len(defaults))
	}
}

func TestBuildSchemaPayload_RulesNonEmpty(t *testing.T) {
	payload := BuildSchemaPayload()
	rules, ok := payload["rules"].([]map[string]any)
	if !ok {
		t.Fatal("rules is not a slice of maps")
	}
	if len(rules) < 1 {
		t.Error("rules is empty, want at least 1 entry")
	}
}

func TestBuildSchemaPayload_NotificationProviders(t *testing.T) {
	payload := BuildSchemaPayload()
	providers, ok := payload["notification_providers"].([]string)
	if !ok {
		t.Fatal("notification_providers is not a string slice")
	}
	if len(providers) == 0 {
		t.Error("notification_providers is empty")
	}
	found := false
	for _, p := range providers {
		if p == "telegram" {
			found = true
			break
		}
	}
	if !found {
		t.Error("notification_providers does not contain telegram")
	}
}

func toStringSet(v any) map[string]bool {
	out := map[string]bool{}
	if list, ok := v.([]string); ok {
		for _, s := range list {
			out[s] = true
		}
	}
	return out
}