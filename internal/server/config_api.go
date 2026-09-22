package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Guliveer/twitch-miner-go/internal/config"
	"gopkg.in/yaml.v3"
)

// handleConfigSchema GET /api/config/schema
// Returns the validation schema and default values for config generation.
func (s *AnalyticsServer) handleConfigSchema(w http.ResponseWriter, r *http.Request) {
	if s.apiKey != "" {
		if key := r.Header.Get("X-API-Key"); key != s.apiKey {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}
	writeJSON(w, http.StatusOK, config.BuildSchemaPayload())
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
