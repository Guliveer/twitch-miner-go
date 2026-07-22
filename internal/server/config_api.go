package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Guliveer/twitch-miner-go/internal/config"
	"gopkg.in/yaml.v3"
)

// configSchema represents the static validation schema returned to clients.
type configSchema struct {
	Strategies        []string         `json:"strategies"`
	ChatModes         []string         `json:"chat_modes"`
	Priorities        []string         `json:"priorities"`
	FollowersOrders   []string         `json:"followers_order"`
	DelayModes        []string         `json:"delay_modes"`
	FilterWhere       []string         `json:"filter_where"`
	FilterBy          []string         `json:"filter_by"`
	WebhookMethods    []string         `json:"webhook_methods"`
	NotificationEvents []string        `json:"notification_events"`
	Defaults          map[string]any   `json:"defaults"`
}

// handleConfigSchema GET /api/config/schema
// Returns the static validation schema and default values for config generation.
func (s *AnalyticsServer) handleConfigSchema(w http.ResponseWriter, _ *http.Request) {
	schema := configSchema{
		Strategies: []string{
			"MOST_VOTED", "HIGH_ODDS", "PERCENTAGE", "SMART_MONEY", "SMART",
			"NUMBER_1", "NUMBER_2", "NUMBER_3", "NUMBER_4",
			"NUMBER_5", "NUMBER_6", "NUMBER_7", "NUMBER_8",
		},
		ChatModes:      []string{"ALWAYS", "NEVER", "ONLINE", "OFFLINE"},
		Priorities:     []string{"STREAK", "DROPS", "ORDER", "SUBSCRIBED", "POINTS_ASCENDING", "POINTS_DESCENDING"},
		FollowersOrders: []string{"ASC", "DESC"},
		DelayModes:     []string{"FROM_START", "FROM_END", "PERCENTAGE"},
		FilterWhere:    []string{"GT", "LT", "GTE", "LTE"},
		FilterBy:       []string{"total_users", "total_points"},
		WebhookMethods: []string{"GET", "POST"},
		NotificationEvents: []string{
			"STREAMER_ONLINE", "STREAMER_OFFLINE",
			"GAIN_FOR_RAID", "GAIN_FOR_CLAIM", "GAIN_FOR_WATCH", "GAIN_FOR_WATCH_STREAK",
			"BET_WIN", "BET_LOSE", "BET_REFUND", "BET_FILTERS", "BET_GENERAL", "BET_FAILED", "BET_START",
			"BONUS_CLAIM", "MOMENT_CLAIM", "JOIN_RAID",
			"DROP_CLAIM", "DROP_CLAIM_AVAILABLE", "DROP_STATUS", "DROP_MILESTONE",
			"CHAT_MENTION", "GIFTED_SUB",
			"MINER_STARTED", "MINER_STOPPED", "MINER_CRASHED",
			"ACCOUNT_CONFIG_RELOADED",
			"TEST",
		},
		Defaults: map[string]any{
			"max_watch_streams": 2,
			"priority":          []string{"STREAK", "DROPS", "ORDER"},
			"poll_interval":     "120s",
			"followers_order":   "ASC",
			"chat":              "ONLINE",
		},
	}
	writeJSON(w, http.StatusOK, schema)
}

// handleConfigValidate POST /api/config/validate
// Accepts an AccountConfig JSON and returns validation results without saving.
func (s *AnalyticsServer) handleConfigValidate(w http.ResponseWriter, r *http.Request) {
	var cfg config.AccountConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"valid":  false,
			"errors": []string{"invalid JSON: " + err.Error()},
		})
		return
	}

	if cfg.Username == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"valid":  false,
			"errors": []string{"username is required"},
		})
		return
	}

	if err := config.Validate(&cfg); err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]any{
			"valid":  false,
			"errors": []string{err.Error()},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"valid":  true,
		"errors": []string{},
	})
}

// handleConfigGenerate POST /api/config/generate
// Accepts an AccountConfig JSON and returns the generated YAML content.
// The YAML is suitable for direct use as configs/<username>.yaml.
func (s *AnalyticsServer) handleConfigGenerate(w http.ResponseWriter, r *http.Request) {
	var cfg config.AccountConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": "invalid JSON: " + err.Error(),
		})
		return
	}

	if cfg.Username == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": "username is required",
		})
		return
	}

	if err := config.Validate(&cfg); err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]any{
			"error": "invalid config: " + err.Error(),
		})
		return
	}

	yamlData, err := yaml.Marshal(&cfg)
	if err != nil {
		writeInternalError(w, s.log, "POST /api/config/generate", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"username": cfg.Username,
		"filename": cfg.Username + ".yaml",
		"yaml":     string(yamlData),
		"generated_at": time.Now().UTC().Format(time.RFC3339),
	})
}
