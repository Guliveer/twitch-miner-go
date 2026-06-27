package configeditor

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"
)

// RunTUI launches the interactive terminal config editor.
func RunTUI(configDir string) error {
	accounts := listAccountNames(configDir)

	if len(accounts) == 0 {
		fmt.Println("No config files found in " + configDir)
		fmt.Println("Create a new account:")
		return runCreateAccount(configDir)
	}

	accounts = append(accounts, "[+ Create new account]")

	var choice string
	if err := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Config Editor — Select account").
				Description(configDir).
				Options(huh.NewOptions(accounts...)...).
				Value(&choice),
		),
	).Run(); err != nil {
		return err
	}

	if choice == "[+ Create new account]" {
		return runCreateAccount(configDir)
	}
	return runEditAccount(configDir, choice)
}

func runCreateAccount(configDir string) error {
	var name string
	if err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("New account name").
				Description("Enter the Twitch username (letters, numbers, _ and -)").
				Validate(func(s string) error {
					if !validName.MatchString(s) {
						return fmt.Errorf("use only letters, numbers, _ and -")
					}
					if _, err := os.Stat(filepath.Join(configDir, s+".yaml")); err == nil {
						return fmt.Errorf("account %q already exists", s)
					}
					return nil
				}).
				Value(&name),
		),
	).Run(); err != nil {
		return err
	}

	defaultCfg := map[string]any{
		"streamers": []any{map[string]any{"username": "placeholder"}},
		"features":  map[string]any{"enable_analytics": true},
	}
	s := &Server{configDir: configDir}
	if err := s.saveRaw(name, defaultCfg); err != nil {
		return fmt.Errorf("failed to create account: %w", err)
	}
	fmt.Printf("Account %q created.\n", name)
	return runEditAccount(configDir, name)
}

func runEditAccount(configDir string, name string) error {
	s := &Server{configDir: configDir}
	cfg, err := s.loadRaw(name)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// ── General ──
	enabled := boolVal(cfg, "enabled", true)
	maxWatch := intVal(cfg, "max_watch_streams", 2)
	maxWatchStr := strconv.Itoa(maxWatch)
	proxy := stringVal(cfg, "proxy", "")

	priorityOptions := []string{"STREAK", "DROPS", "ORDER", "SUBSCRIBED", "POINTS_ASCENDING", "POINTS_DESCENDING"}
	selectedPriorities := stringSliceVal(cfg, "priority", []string{"STREAK", "DROPS", "ORDER"})

	// ── Features ──
	features, _ := cfg["features"].(map[string]any)
	if features == nil {
		features = map[string]any{}
	}
	claimDropsStartup := boolVal(features, "claim_drops_startup", false)
	enableAnalytics := boolVal(features, "enable_analytics", false)

	// ── Category Watcher ──
	cw, _ := cfg["category_watcher"].(map[string]any)
	if cw == nil {
		cw = map[string]any{}
	}
	cwEnabled := boolVal(cw, "enabled", false)
	cwInterval := stringVal(cw, "poll_interval", "120s")
	cwCategoriesRaw, _ := cw["categories"].([]any)
	cwCategoriesSlugs := make([]string, 0, len(cwCategoriesRaw))
	for _, c := range cwCategoriesRaw {
		if cm, ok := c.(map[string]any); ok {
			if slug, ok := cm["slug"].(string); ok && slug != "" {
				cwCategoriesSlugs = append(cwCategoriesSlugs, slug)
			}
		}
	}
	cwCategoriesStr := strings.Join(cwCategoriesSlugs, ", ")

	// ── Team Watcher ──
	tw, _ := cfg["team_watcher"].(map[string]any)
	if tw == nil {
		tw = map[string]any{}
	}
	twEnabled := boolVal(tw, "enabled", false)
	twInterval := stringVal(tw, "poll_interval", "120s")
	twTeamsRaw, _ := tw["teams"].([]any)
	twTeamNames := make([]string, 0, len(twTeamsRaw))
	for _, t := range twTeamsRaw {
		if tm, ok := t.(map[string]any); ok {
			if n, ok := tm["name"].(string); ok && n != "" {
				twTeamNames = append(twTeamNames, n)
			}
		}
	}
	twTeamsStr := strings.Join(twTeamNames, ", ")

	// ── Streamers ──
	streamersRaw, _ := cfg["streamers"].([]any)
	streamerNames := make([]string, 0, len(streamersRaw))
	for _, st := range streamersRaw {
		if sm, ok := st.(map[string]any); ok {
			if u, ok := sm["username"].(string); ok && u != "" {
				streamerNames = append(streamerNames, u)
			}
		}
	}
	streamersStr := strings.Join(streamerNames, ", ")

	// ── Action choice ──
	var action string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title("Editing: "+name).
				Description("Use Tab/Enter to navigate, Esc to cancel."),
			huh.NewConfirm().
				Title("Enabled").
				Description("Enable this account").
				Value(&enabled),
		).Title("General"),

		huh.NewGroup(
			huh.NewInput().
				Title("Max watch streams").
				Validate(func(s string) error {
					n, err := strconv.Atoi(s)
					if err != nil || n < 1 {
						return fmt.Errorf("must be a positive integer")
					}
					return nil
				}).
				Value(&maxWatchStr),
			huh.NewInput().
				Title("Proxy").
				Description("e.g. socks5://127.0.0.1:1080 (leave blank to disable)").
				Value(&proxy),
			huh.NewMultiSelect[string]().
				Title("Priority").
				Description("Select and order priorities (first wins)").
				Options(huh.NewOptions(priorityOptions...)...).
				Value(&selectedPriorities),
		).Title("General (continued)"),

		huh.NewGroup(
			huh.NewConfirm().
				Title("Claim drops on startup").
				Value(&claimDropsStartup),
			huh.NewConfirm().
				Title("Enable analytics").
				Value(&enableAnalytics),
		).Title("Features"),

		huh.NewGroup(
			huh.NewConfirm().
				Title("Category Watcher enabled").
				Value(&cwEnabled),
			huh.NewInput().
				Title("Category Watcher poll interval").
				Description("e.g. 120s, 5m, 1h").
				Validate(func(s string) error {
					if s != "" && !isValidDuration(s) {
						return fmt.Errorf("invalid duration (e.g. 120s, 5m, 1h30m)")
					}
					return nil
				}).
				Value(&cwInterval),
			huh.NewInput().
				Title("Categories (comma-separated slugs)").
				Description("e.g. just-chatting, science-and-technology").
				Value(&cwCategoriesStr),
		).Title("Category Watcher"),

		huh.NewGroup(
			huh.NewConfirm().
				Title("Team Watcher enabled").
				Value(&twEnabled),
			huh.NewInput().
				Title("Team Watcher poll interval").
				Description("e.g. 120s, 5m").
				Validate(func(s string) error {
					if s != "" && !isValidDuration(s) {
						return fmt.Errorf("invalid duration (e.g. 120s, 5m, 1h30m)")
					}
					return nil
				}).
				Value(&twInterval),
			huh.NewInput().
				Title("Teams (comma-separated names)").
				Description("e.g. nrg, sentinels").
				Value(&twTeamsStr),
		).Title("Team Watcher"),

		huh.NewGroup(
			huh.NewInput().
				Title("Streamers (comma-separated usernames)").
				Description("e.g. streamer1, streamer2").
				Value(&streamersStr),
		).Title("Streamers"),

		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Save changes?").
				Options(
					huh.NewOption("Save and exit", "save"),
					huh.NewOption("Discard changes", "discard"),
					huh.NewOption("Delete this account", "delete"),
				).
				Value(&action),
		).Title("Confirm"),
	)

	if err := form.Run(); err != nil {
		return err
	}

	switch action {
	case "discard":
		fmt.Println("Changes discarded.")
		return nil
	case "delete":
		var confirm bool
		if err := huh.NewForm(huh.NewGroup(
			huh.NewConfirm().
				Title(fmt.Sprintf("Delete account %q?", name)).
				Description("This will permanently remove the config file.").
				Value(&confirm),
		)).Run(); err != nil {
			return err
		}
		if confirm {
			if err := os.Remove(filepath.Join(configDir, name+".yaml")); err != nil {
				return fmt.Errorf("failed to delete: %w", err)
			}
			fmt.Printf("Account %q deleted.\n", name)
		}
		return nil
	}

	// Build updated config
	maxWatchInt, _ := strconv.Atoi(maxWatchStr)
	if maxWatchInt < 1 {
		maxWatchInt = 2
	}
	if !enabled {
		cfg["enabled"] = false
	} else {
		delete(cfg, "enabled")
	}
	cfg["max_watch_streams"] = maxWatchInt
	if proxy != "" {
		cfg["proxy"] = proxy
	} else {
		delete(cfg, "proxy")
	}
	if len(selectedPriorities) > 0 {
		cfg["priority"] = selectedPriorities
	}

	featuresMap := map[string]any{}
	if claimDropsStartup {
		featuresMap["claim_drops_startup"] = true
	}
	if enableAnalytics {
		featuresMap["enable_analytics"] = true
	}
	if len(featuresMap) > 0 {
		cfg["features"] = featuresMap
	} else {
		delete(cfg, "features")
	}

	// Category watcher
	cwMap := map[string]any{}
	if cwEnabled {
		cwMap["enabled"] = true
	}
	if cwInterval != "" && cwInterval != "120s" {
		cwMap["poll_interval"] = cwInterval
	}
	cwSlugs := parseCSSV(cwCategoriesStr)
	if len(cwSlugs) > 0 {
		cats := make([]any, len(cwSlugs))
		for i, slug := range cwSlugs {
			cats[i] = map[string]any{"slug": slug}
		}
		cwMap["categories"] = cats
	}
	if len(cwMap) > 0 {
		cfg["category_watcher"] = cwMap
	} else {
		delete(cfg, "category_watcher")
	}

	// Team watcher
	twMap := map[string]any{}
	if twEnabled {
		twMap["enabled"] = true
	}
	if twInterval != "" && twInterval != "120s" {
		twMap["poll_interval"] = twInterval
	}
	twNames := parseCSSV(twTeamsStr)
	if len(twNames) > 0 {
		teams := make([]any, len(twNames))
		for i, n := range twNames {
			teams[i] = map[string]any{"name": n}
		}
		twMap["teams"] = teams
	}
	if len(twMap) > 0 {
		cfg["team_watcher"] = twMap
	} else {
		delete(cfg, "team_watcher")
	}

	// Streamers
	streamerList := parseCSSV(streamersStr)
	if len(streamerList) > 0 {
		streamers := make([]any, len(streamerList))
		for i, u := range streamerList {
			streamers[i] = map[string]any{"username": u}
		}
		cfg["streamers"] = streamers
	} else {
		delete(cfg, "streamers")
	}

	if errs := validateConfig(cfg); len(errs) > 0 {
		fmt.Println("Validation errors:")
		for _, e := range errs {
			fmt.Println("  - " + e)
		}
		return fmt.Errorf("config not saved due to validation errors")
	}

	if err := s.saveRaw(name, cleanConfig(cfg)); err != nil {
		return fmt.Errorf("failed to save: %w", err)
	}
	fmt.Printf("Account %q saved.\n", name)
	return nil
}

func listAccountNames(configDir string) []string {
	entries, err := os.ReadDir(configDir)
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, ".example") {
			continue
		}
		if strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml") {
			names = append(names, strings.TrimSuffix(strings.TrimSuffix(name, ".yaml"), ".yml"))
		}
	}
	return names
}

func parseCSSV(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func boolVal(m map[string]any, key string, def bool) bool {
	v, ok := m[key]
	if !ok {
		return def
	}
	b, ok := v.(bool)
	if !ok {
		return def
	}
	return b
}

func intVal(m map[string]any, key string, def int) int {
	v, ok := m[key]
	if !ok {
		return def
	}
	switch n := v.(type) {
	case int:
		return n
	case float64:
		return int(n)
	}
	return def
}

func stringVal(m map[string]any, key string, def string) string {
	v, ok := m[key]
	if !ok {
		return def
	}
	s, ok := v.(string)
	if !ok {
		return def
	}
	return s
}

func stringSliceVal(m map[string]any, key string, def []string) []string {
	v, ok := m[key]
	if !ok {
		return def
	}
	raw, ok := v.([]any)
	if !ok {
		return def
	}
	out := make([]string, 0, len(raw))
	for _, item := range raw {
		if s, ok := item.(string); ok {
			out = append(out, s)
		}
	}
	return out
}
