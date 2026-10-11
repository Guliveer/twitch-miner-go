package config

import "github.com/Guliveer/twitch-miner-go/internal/model"

// filterWhere, filterBy, webhookMethods and notificationProviders have no
// model types yet; they are kept here as the single source of truth for the
// schema payload until they are promoted into internal/model.
var (
	filterWhere           = []string{"GT", "LT", "GTE", "LTE"}
	filterBy              = []string{"total_users", "total_points", "percentage_users", "odds_percentage", "odds", "top_points", "decision_users", "decision_points"}
	webhookMethods        = []string{"GET", "POST"}
	notificationProviders = []string{"telegram", "discord", "webhook", "matrix", "pushover", "gotify"}
)

// rules surfaces lightweight validation metadata for client-side warnings.
// It is not a full validation rewrite — config.Validate() remains authoritative.
var rules = []map[string]any{
	{"field": "streamer_defaults.make_predictions", "condition": "requires", "target": "streamer_defaults.bet", "message": "make_predictions requires bet config"},
	{"field": "badge_watcher", "condition": "requires_priority", "value": "BADGES", "message": "badge_watcher enabled requires BADGES in priority"},
	{"field": "badge_watcher.streamer_limit", "condition": "range", "min": 1, "max": 2, "message": "streamer_limit must be 1-2"},
	{"field": "max_watch_streams", "condition": "min", "value": 0, "message": "must be non-negative"},
	{"field": "category_watcher.poll_interval", "condition": "duration", "message": "must be a Go duration string"},
	{"field": "team_watcher.poll_interval", "condition": "duration", "message": "must be a Go duration string"},
	{"field": "badge_watcher.poll_interval", "condition": "duration", "message": "must be a Go duration string"},
}

// BuildSchemaPayload returns the config schema served to clients: enum lists
// derived from the model constants, defaults, and validation metadata.
func BuildSchemaPayload() map[string]any {
	return map[string]any{
		"strategies":             stringValues(model.AllStrategies()),
		"chat_modes":             stringValues(model.AllChatModes()),
		"priorities":             stringValues(model.AllPriorities()),
		"followers_order":        stringValues(model.AllFollowersOrders()),
		"delay_modes":            stringValues(model.AllDelayModes()),
		"filter_where":           filterWhere,
		"filter_by":              filterBy,
		"outcome_keys":           filterBy,
		"webhook_methods":        webhookMethods,
		"notification_events":    stringValues(model.AllEvents()),
		"notification_providers": notificationProviders,
		"defaults":               BuildDefaults(),
		"rules":                  rules,
	}
}

// BuildDefaults returns the default values a zero-value AccountConfig gets
// from applyDefaults, as a JSON-serializable map. Values are replicated here
// so the builder has no side effects and no dependency beyond model.
func BuildDefaults() map[string]any {
	return map[string]any{
		"max_watch_streams":              2,
		"streak_watch_streams":           2,
		"watch_streak_minutes":           10,
		"priority":                       []string{"STREAK", "DROPS", "ORDER"},
		"poll_interval":                  "120s",
		"category_watcher_poll_interval": "120s",
		"team_watcher_poll_interval":     "120s",
		"badge_watcher_poll_interval":    "5m",
		"badge_watcher_streamer_limit":   1,
		"followers_order":                "ASC",
		"chat":                           "ONLINE",
		"make_predictions":               true,
		"follow_raid":                    true,
		"claim_drops":                    true,
		"claim_moments":                  true,
		"watch_streak":                   true,
		"bet": map[string]any{
			"strategy":       "SMART",
			"percentage":     5,
			"percentage_gap": 20,
			"max_points":     50000,
			"minimum_points": 0,
			"stealth_mode":   false,
			"delay":          6,
			"delay_mode":     "FROM_END",
		},
	}
}

func stringValues[T interface{ String() string }](in []T) []string {
	out := make([]string, 0, len(in))
	for _, v := range in {
		out = append(out, v.String())
	}
	return out
}