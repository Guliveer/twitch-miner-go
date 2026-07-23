package configeditor

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

var validName = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

var secretFields = map[string][]string{
	"telegram": {"token", "chat_id"},
	"discord":  {"webhook_url"},
	"webhook":  {"endpoint"},
	"matrix":   {"homeserver", "room_id", "access_token"},
	"pushover": {"api_token", "user_key"},
	"gotify":   {"url", "token"},
}

// schema is the static validation schema returned to the frontend.
var schema = map[string]any{
	"strategies": []string{
		"MOST_VOTED", "HIGH_ODDS", "PERCENTAGE", "SMART_MONEY", "SMART",
		"NUMBER_1", "NUMBER_2", "NUMBER_3", "NUMBER_4",
		"NUMBER_5", "NUMBER_6", "NUMBER_7", "NUMBER_8",
	},
	"chat_modes":      []string{"ALWAYS", "NEVER", "ONLINE", "OFFLINE"},
	"priorities":      []string{"STREAK", "DROPS", "ORDER", "SUBSCRIBED", "POINTS_ASCENDING", "POINTS_DESCENDING"},
	"followers_order": []string{"ASC", "DESC"},
	"delay_modes":     []string{"FROM_START", "FROM_END", "PERCENTAGE"},
	"filter_where":    []string{"GT", "LT", "GTE", "LTE"},
	"filter_by":       []string{"total_users", "total_points"},
	"webhook_methods": []string{"GET", "POST"},
	"notification_events": []string{
		"STREAMER_ONLINE", "STREAMER_OFFLINE",
		"GAIN_FOR_RAID", "GAIN_FOR_CLAIM", "GAIN_FOR_WATCH", "GAIN_FOR_WATCH_STREAK",
		"BET_WIN", "BET_LOSE", "BET_REFUND", "BET_FILTERS", "BET_GENERAL", "BET_FAILED", "BET_START",
		"BONUS_CLAIM", "MOMENT_CLAIM", "JOIN_RAID",
		"DROP_CLAIM", "DROP_STATUS", "DROP_MILESTONE",
		"CHAT_MENTION", "GIFTED_SUB",
		"MINER_STARTED", "MINER_STOPPED", "MINER_CRASHED",
		"ACCOUNT_CONFIG_RELOADED",
		"TEST",
	},
	"defaults": map[string]any{
		"max_watch_streams":              2,
		"priority":                       []string{"STREAK", "DROPS", "ORDER"},
		"category_watcher_poll_interval": "120s",
		"team_watcher_poll_interval":     "120s",
		"followers_order":                "ASC",
	},
}

// Server handles HTTP requests for the config editor web UI.
type Server struct {
	configDir string
	mux       *http.ServeMux
}

// NewServer creates a new Server serving config files from configDir.
func NewServer(configDir string) *Server {
	s := &Server{configDir: configDir, mux: http.NewServeMux()}
	s.registerRoutes()
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) registerRoutes() {
	sub, _ := fs.Sub(webFS, "web")
	staticFS := http.FileServer(http.FS(sub))

	s.mux.HandleFunc("/api/schema", s.handleCORS(s.handleSchema))
	s.mux.HandleFunc("/api/accounts", s.handleCORS(s.handleAccounts))
	s.mux.HandleFunc("/api/accounts/", s.handleCORS(s.handleAccount))
	s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		s.setCORSHeaders(w)
		staticFS.ServeHTTP(w, r)
	})
}

func (s *Server) setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func (s *Server) handleCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.setCORSHeaders(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next(w, r)
	}
}

func (s *Server) handleSchema(w http.ResponseWriter, _ *http.Request) {
	sendJSON(w, http.StatusOK, schema)
}

func (s *Server) handleAccounts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		sendJSON(w, http.StatusOK, s.listAccounts())
	case http.MethodPost:
		s.createAccount(w, r)
	default:
		sendError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) handleAccount(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/api/accounts/")
	if !validName.MatchString(name) {
		sendError(w, http.StatusBadRequest, "invalid account name")
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.getAccount(w, name)
	case http.MethodPut:
		s.updateAccount(w, r, name)
	case http.MethodDelete:
		s.deleteAccount(w, name)
	default:
		sendError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

type accountMeta struct {
	Name               string `json:"name"`
	Enabled            bool   `json:"enabled"`
	StreamerCount      int    `json:"streamer_count"`
	HasCategoryWatcher bool   `json:"has_category_watcher"`
	HasTeamWatcher     bool   `json:"has_team_watcher"`
	HasFollowers       bool   `json:"has_followers"`
	HasNotifications   bool   `json:"has_notifications"`
	Error              bool   `json:"error,omitempty"`
}

func (s *Server) listAccounts() []accountMeta {
	entries, err := os.ReadDir(s.configDir)
	if err != nil {
		return []accountMeta{}
	}
	accounts := []accountMeta{}
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || strings.HasSuffix(name, ".example") {
			continue
		}
		if !strings.HasSuffix(name, ".yaml") && !strings.HasSuffix(name, ".yml") {
			continue
		}
		accountName := strings.TrimSuffix(strings.TrimSuffix(name, ".yaml"), ".yml")
		cfg, err := s.loadRaw(accountName)
		if err != nil {
			accounts = append(accounts, accountMeta{Name: accountName, Error: true})
			continue
		}
		meta := accountMeta{Name: accountName, Enabled: true}
		if v, ok := cfg["enabled"].(bool); ok {
			meta.Enabled = v
		}
		if streamers, ok := cfg["streamers"].([]any); ok {
			meta.StreamerCount = len(streamers)
		}
		if cw, ok := cfg["category_watcher"].(map[string]any); ok {
			meta.HasCategoryWatcher = cw["enabled"] == true
		}
		if tw, ok := cfg["team_watcher"].(map[string]any); ok {
			meta.HasTeamWatcher = tw["enabled"] == true
		}
		if fol, ok := cfg["followers"].(map[string]any); ok {
			meta.HasFollowers = fol["enabled"] == true
		}
		if notif, ok := cfg["notifications"].(map[string]any); ok {
			providers := []string{"telegram", "discord", "webhook", "matrix", "pushover", "gotify"}
			for _, p := range providers {
				if pCfg, ok := notif[p].(map[string]any); ok && pCfg["enabled"] == true {
					meta.HasNotifications = true
					break
				}
			}
		}
		accounts = append(accounts, meta)
	}
	return accounts
}

func (s *Server) getAccount(w http.ResponseWriter, name string) {
	cfg, err := s.loadRaw(name)
	if err != nil {
		if os.IsNotExist(err) {
			sendError(w, http.StatusNotFound, fmt.Sprintf("account %q not found", name))
			return
		}
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	sendJSON(w, http.StatusOK, stripSecrets(cfg))
}

func (s *Server) createAccount(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name   string         `json:"name"`
		Config map[string]any `json:"config"`
	}
	if err := readJSON(r.Body, &body); err != nil {
		sendError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if !validName.MatchString(body.Name) {
		sendError(w, http.StatusBadRequest, "invalid account name: use only letters, numbers, underscores, hyphens")
		return
	}
	if _, err := os.Stat(s.configPath(body.Name)); err == nil {
		sendError(w, http.StatusConflict, fmt.Sprintf("account %q already exists", body.Name))
		return
	}
	if errs := validateConfig(body.Config); len(errs) > 0 {
		sendJSON(w, http.StatusUnprocessableEntity, map[string]any{"errors": errs})
		return
	}
	if err := s.saveRaw(body.Name, cleanConfig(body.Config)); err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	sendJSON(w, http.StatusCreated, map[string]string{"name": body.Name, "message": "Account created"})
}

func (s *Server) updateAccount(w http.ResponseWriter, r *http.Request, name string) {
	existing, err := s.loadRaw(name)
	if err != nil {
		if os.IsNotExist(err) {
			sendError(w, http.StatusNotFound, fmt.Sprintf("account %q not found", name))
			return
		}
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	var body struct {
		Config map[string]any `json:"config"`
	}
	if err := readJSON(r.Body, &body); err != nil {
		sendError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	cfg := body.Config
	if cfg == nil {
		sendError(w, http.StatusBadRequest, "missing config field")
		return
	}
	if errs := validateConfig(cfg); len(errs) > 0 {
		sendJSON(w, http.StatusUnprocessableEntity, map[string]any{"errors": errs})
		return
	}
	merged := mergeSecretsBack(cfg, existing)
	if err := s.saveRaw(name, cleanConfig(merged)); err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	sendJSON(w, http.StatusOK, map[string]string{"name": name, "message": "Account updated"})
}

func (s *Server) deleteAccount(w http.ResponseWriter, name string) {
	path := s.configPath(name)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		sendError(w, http.StatusNotFound, fmt.Sprintf("account %q not found", name))
		return
	}
	if err := os.Remove(path); err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	sendJSON(w, http.StatusOK, map[string]string{"name": name, "message": "Account deleted"})
}

func (s *Server) configPath(name string) string {
	return filepath.Join(s.configDir, name+".yaml")
}

func (s *Server) loadRaw(name string) (map[string]any, error) {
	data, err := os.ReadFile(s.configPath(name))
	if err != nil {
		return nil, err
	}
	var out map[string]any
	if err := yaml.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	if out == nil {
		out = map[string]any{}
	}
	return out, nil
}

func (s *Server) saveRaw(name string, cfg map[string]any) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	path := s.configPath(name)
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func stripSecrets(cfg map[string]any) map[string]any {
	out := shallowCopy(cfg)
	notif, ok := out["notifications"].(map[string]any)
	if !ok {
		return out
	}
	notifCopy := shallowCopy(notif)
	for provider, fields := range secretFields {
		pCfg, ok := notifCopy[provider].(map[string]any)
		if !ok {
			continue
		}
		pCopy := shallowCopy(pCfg)
		for _, f := range fields {
			delete(pCopy, f)
		}
		notifCopy[provider] = pCopy
	}
	out["notifications"] = notifCopy
	return out
}

func mergeSecretsBack(newCfg, existing map[string]any) map[string]any {
	out := shallowCopy(newCfg)
	existingNotif, ok := existing["notifications"].(map[string]any)
	if !ok {
		return out
	}
	newNotif, ok := out["notifications"].(map[string]any)
	if !ok {
		return out
	}
	notifCopy := shallowCopy(newNotif)
	for provider, fields := range secretFields {
		newPCfg, ok := notifCopy[provider].(map[string]any)
		if !ok {
			continue
		}
		existingPCfg, ok := existingNotif[provider].(map[string]any)
		if !ok {
			continue
		}
		pCopy := shallowCopy(newPCfg)
		for _, f := range fields {
			if v, ok := existingPCfg[f]; ok {
				pCopy[f] = v
			}
		}
		notifCopy[provider] = pCopy
	}
	out["notifications"] = notifCopy
	return out
}

func cleanConfig(cfg map[string]any) map[string]any {
	result := removeEmpty(cfg)
	if result == nil {
		return map[string]any{}
	}
	if m, ok := result.(map[string]any); ok {
		return m
	}
	return map[string]any{}
}

func removeEmpty(v any) any {
	if v == nil {
		return nil
	}
	switch val := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(val))
		for k, child := range val {
			if child == nil || child == "" {
				continue
			}
			cleaned := removeEmpty(child)
			if cleaned != nil {
				out[k] = cleaned
			}
		}
		if len(out) == 0 {
			return nil
		}
		return out
	case []any:
		var out []any
		for _, item := range val {
			cleaned := removeEmpty(item)
			if cleaned != nil {
				out = append(out, cleaned)
			}
		}
		if len(out) == 0 {
			return nil
		}
		return out
	}
	return v
}

func validateConfig(cfg map[string]any) []string {
	var errs []string

	if mws, ok := cfg["max_watch_streams"].(float64); ok && mws < 0 {
		errs = append(errs, "max_watch_streams must be non-negative (0 = unlimited)")
	}

	hasStreamers := func() bool {
		s, ok := cfg["streamers"].([]any)
		return ok && len(s) > 0
	}
	hasCW := func() bool {
		cw, ok := cfg["category_watcher"].(map[string]any)
		return ok && cw["enabled"] == true
	}
	hasTW := func() bool {
		tw, ok := cfg["team_watcher"].(map[string]any)
		return ok && tw["enabled"] == true
	}
	hasFollowers := func() bool {
		f, ok := cfg["followers"].(map[string]any)
		return ok && f["enabled"] == true
	}

	if !hasStreamers() && !hasFollowers() && !hasCW() && !hasTW() {
		errs = append(errs, "at least one of streamers, followers, category_watcher, or team_watcher must be configured")
	}

	if streamers, ok := cfg["streamers"].([]any); ok {
		for i, s := range streamers {
			sm, ok := s.(map[string]any)
			if !ok || strings.TrimSpace(fmt.Sprint(sm["username"])) == "" {
				errs = append(errs, fmt.Sprintf("streamer at index %d has empty username", i))
			}
		}
	}

	if cw, ok := cfg["category_watcher"].(map[string]any); ok && cw["enabled"] == true {
		cats, _ := cw["categories"].([]any)
		if len(cats) == 0 {
			errs = append(errs, "category_watcher is enabled but no categories are configured")
		}
	}

	if tw, ok := cfg["team_watcher"].(map[string]any); ok && tw["enabled"] == true {
		teams, _ := tw["teams"].([]any)
		if len(teams) == 0 {
			errs = append(errs, "team_watcher is enabled but no teams are configured")
		}
	}

	if sd, ok := cfg["streamer_defaults"].(map[string]any); ok {
		if sd["make_predictions"] == true {
			if _, hasBet := sd["bet"]; !hasBet {
				errs = append(errs, "make_predictions is enabled in streamer_defaults but no bet config is set")
			}
		}
	}

	validateDuration(cfg, "category_watcher", "poll_interval", &errs)
	validateDuration(cfg, "team_watcher", "poll_interval", &errs)
	if notif, ok := cfg["notifications"].(map[string]any); ok {
		if batch, ok := notif["batch"].(map[string]any); ok {
			if v, ok := batch["interval"].(string); ok && v != "" && !isValidDuration(v) {
				errs = append(errs, fmt.Sprintf(`notifications.batch.interval %q is not a valid duration`, v))
			}
		}
	}

	return errs
}

func validateDuration(cfg map[string]any, section, field string, errs *[]string) {
	sec, ok := cfg[section].(map[string]any)
	if !ok {
		return
	}
	v, ok := sec[field].(string)
	if !ok || v == "" {
		return
	}
	if !isValidDuration(v) {
		*errs = append(*errs, fmt.Sprintf(`%s.%s %q is not a valid duration (e.g. 120s, 15m, 1h30m)`, section, field, v))
	}
}

var durationRE = regexp.MustCompile(`^(?:\d+(?:\.\d+)?(?:ns|us|µs|ms|s|m|h))+$`)

func isValidDuration(s string) bool {
	return durationRE.MatchString(strings.TrimSpace(s))
}

func shallowCopy(m map[string]any) map[string]any {
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func sendJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func sendError(w http.ResponseWriter, status int, msg string) {
	sendJSON(w, status, map[string]string{"error": msg})
}

func readJSON(r io.Reader, v any) error {
	return json.NewDecoder(r).Decode(v)
}
