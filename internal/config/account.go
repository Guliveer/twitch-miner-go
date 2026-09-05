package config

import (
	"encoding/json"
	"time"

	"github.com/Guliveer/twitch-miner-go/internal/model"
)

// jsonDuration is a time.Duration that marshals/unmarshals as a human-readable string (e.g. "2m0s").
type jsonDuration time.Duration

func (d jsonDuration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

func (d *jsonDuration) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	dur, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*d = jsonDuration(dur)
	return nil
}

// AccountConfig represents the full configuration for a single Twitch account.
// It is loaded from a YAML file and optionally overlaid with environment variables.
type AccountConfig struct {
	Username string `yaml:"-" json:"username"`

	Enabled *bool `yaml:"enabled,omitempty" json:"enabled,omitempty"`

	Auth AuthConfig `yaml:"-" json:"-"`

	Features FeaturesConfig `yaml:"features" json:"features"`

	MaxWatchStreams *int `yaml:"max_watch_streams,omitempty" json:"max_watch_streams,omitempty"`

	// StreakWatchStreams narrows the watch set while any channel still has a
	// watch streak pending. Twitch only credits about two concurrent streams,
	// so spreading minute-watched events across every live channel leaves the
	// choice to Twitch and no streak reliably lands. Once nothing is pending,
	// the set widens back to MaxWatchStreams. Zero disables the narrowing.
	StreakWatchStreams *int `yaml:"streak_watch_streams,omitempty" json:"streak_watch_streams,omitempty"`

	// WatchStreakMinutes is how long a channel may hold a streak slot before
	// it gives way to the next one, whether or not the streak arrived.
	WatchStreakMinutes *float64 `yaml:"watch_streak_minutes,omitempty" json:"watch_streak_minutes,omitempty"`

	// PreferredStreamers lists channel logins that the PREFERRED priority
	// picks first, in the order given.
	PreferredStreamers []string `yaml:"preferred_streamers,omitempty" json:"preferred_streamers,omitempty"`

	Priority []string `yaml:"priority" json:"priority"`

	Proxy string `yaml:"proxy,omitempty" json:"proxy,omitempty"`

	CategoryWatcher CategoryWatcherConfig `yaml:"category_watcher" json:"category_watcher"`

	TeamWatcher TeamWatcherConfig `yaml:"team_watcher" json:"team_watcher"`

	StreamerDefaults StreamerSettingsConfig `yaml:"streamer_defaults" json:"streamer_defaults"`

	Streamers []StreamerConfig `yaml:"streamers" json:"streamers"`

	Blacklist []string `yaml:"blacklist" json:"blacklist"`

	CategoryBlacklist []string `yaml:"category_blacklist" json:"category_blacklist"`

	Followers FollowersConfig `yaml:"followers" json:"followers"`

	Notifications NotificationsConfig `yaml:"notifications" json:"notifications"`
}

// AuthConfig holds authentication-related settings.
type AuthConfig struct {
	AuthToken string `yaml:"auth_token" json:"-"`
	Password  string `yaml:"password" json:"-"`
}

// FeaturesConfig holds global feature toggles for an account.
type FeaturesConfig struct {
	ClaimDropsStartup bool `yaml:"claim_drops_startup" json:"claim_drops_startup"`
	EnableAnalytics   bool `yaml:"enable_analytics" json:"enable_analytics"`
}

// CategoryWatcherConfig holds settings for the category watcher.
type CategoryWatcherConfig struct {
	Enabled      bool             `yaml:"enabled" json:"enabled"`
	PollInterval time.Duration    `yaml:"poll_interval" json:"-"`
	DropsOnly    bool             `yaml:"drops_only" json:"drops_only"`
	Categories   []CategoryConfig `yaml:"categories" json:"categories"`
}

func (c CategoryWatcherConfig) MarshalJSON() ([]byte, error) {
	type alias struct {
		Enabled      bool             `json:"enabled"`
		PollInterval jsonDuration     `json:"poll_interval,omitempty"`
		DropsOnly    bool             `json:"drops_only"`
		Categories   []CategoryConfig `json:"categories"`
	}
	return json.Marshal(alias{c.Enabled, jsonDuration(c.PollInterval), c.DropsOnly, c.Categories})
}

func (c *CategoryWatcherConfig) UnmarshalJSON(b []byte) error {
	type alias struct {
		Enabled      bool             `json:"enabled"`
		PollInterval jsonDuration     `json:"poll_interval,omitempty"`
		DropsOnly    bool             `json:"drops_only"`
		Categories   []CategoryConfig `json:"categories"`
	}
	var a alias
	if err := json.Unmarshal(b, &a); err != nil {
		return err
	}
	c.Enabled = a.Enabled
	c.PollInterval = time.Duration(a.PollInterval)
	c.DropsOnly = a.DropsOnly
	c.Categories = a.Categories
	return nil
}

// CategoryConfig holds settings for a single game category.
type CategoryConfig struct {
	Slug      string `yaml:"slug" json:"slug"`
	DropsOnly *bool  `yaml:"drops_only,omitempty" json:"drops_only,omitempty"`
}

// TeamWatcherConfig holds settings for the team watcher.
type TeamWatcherConfig struct {
	Enabled      bool          `yaml:"enabled" json:"enabled"`
	PollInterval time.Duration `yaml:"poll_interval" json:"-"`
	Teams        []TeamConfig  `yaml:"teams" json:"teams"`
}

func (t TeamWatcherConfig) MarshalJSON() ([]byte, error) {
	type alias struct {
		Enabled      bool         `json:"enabled"`
		PollInterval jsonDuration `json:"poll_interval,omitempty"`
		Teams        []TeamConfig `json:"teams"`
	}
	return json.Marshal(alias{t.Enabled, jsonDuration(t.PollInterval), t.Teams})
}

func (t *TeamWatcherConfig) UnmarshalJSON(b []byte) error {
	type alias struct {
		Enabled      bool         `json:"enabled"`
		PollInterval jsonDuration `json:"poll_interval,omitempty"`
		Teams        []TeamConfig `json:"teams"`
	}
	var a alias
	if err := json.Unmarshal(b, &a); err != nil {
		return err
	}
	t.Enabled = a.Enabled
	t.PollInterval = time.Duration(a.PollInterval)
	t.Teams = a.Teams
	return nil
}

// TeamConfig holds settings for a single Twitch team.
type TeamConfig struct {
	Name string `yaml:"name" json:"name"`
}

// StreamerSettingsConfig is the YAML representation of per-streamer settings.
type StreamerSettingsConfig struct {
	MakePredictions *bool              `yaml:"make_predictions,omitempty" json:"make_predictions,omitempty"`
	FollowRaid      *bool              `yaml:"follow_raid,omitempty" json:"follow_raid,omitempty"`
	ClaimDrops      *bool              `yaml:"claim_drops,omitempty" json:"claim_drops,omitempty"`
	ClaimMoments    *bool              `yaml:"claim_moments,omitempty" json:"claim_moments,omitempty"`
	WatchStreak     *bool              `yaml:"watch_streak,omitempty" json:"watch_streak,omitempty"`
	CommunityGoals  *bool              `yaml:"community_goals,omitempty" json:"community_goals,omitempty"`
	DropsOnly       *bool              `yaml:"drops_only,omitempty" json:"drops_only,omitempty"`
	Chat            string             `yaml:"chat,omitempty" json:"chat,omitempty"`
	Bet             *BetSettingsConfig `yaml:"bet,omitempty" json:"bet,omitempty"`
}

// BetSettingsConfig is the YAML representation of bet settings.
type BetSettingsConfig struct {
	Strategy        string                 `yaml:"strategy,omitempty" json:"strategy,omitempty"`
	Percentage      *int                   `yaml:"percentage,omitempty" json:"percentage,omitempty"`
	PercentageGap   *int                   `yaml:"percentage_gap,omitempty" json:"percentage_gap,omitempty"`
	MaxPoints       *int                   `yaml:"max_points,omitempty" json:"max_points,omitempty"`
	MinimumPoints   *int                   `yaml:"minimum_points,omitempty" json:"minimum_points,omitempty"`
	StealthMode     *bool                  `yaml:"stealth_mode,omitempty" json:"stealth_mode,omitempty"`
	Delay           *float64               `yaml:"delay,omitempty" json:"delay,omitempty"`
	DelayMode       string                 `yaml:"delay_mode,omitempty" json:"delay_mode,omitempty"`
	FilterCondition *FilterConditionConfig `yaml:"filter_condition,omitempty" json:"filter_condition,omitempty"`
}

// FilterConditionConfig is the YAML representation of a filter condition.
type FilterConditionConfig struct {
	By    string  `yaml:"by" json:"by"`
	Where string  `yaml:"where" json:"where"`
	Value float64 `yaml:"value" json:"value"`
}

// StreamerConfig holds per-streamer configuration from YAML.
type StreamerConfig struct {
	Username string                  `yaml:"username" json:"username"`
	Settings *StreamerSettingsConfig `yaml:"settings,omitempty" json:"settings,omitempty"`
}

// FollowersConfig holds settings for watching followed channels.
type FollowersConfig struct {
	Enabled bool   `yaml:"enabled" json:"enabled"`
	Order   string `yaml:"order" json:"order"`
}

// BatchConfig holds notification batching settings.
// When set at the notifications level, it provides global defaults.
// When set per-provider, it overrides the global defaults.
type BatchConfig struct {
	Enabled         *bool         `yaml:"enabled,omitempty" json:"enabled,omitempty"`
	Interval        time.Duration `yaml:"interval,omitempty" json:"-"`
	MaxEntries      int           `yaml:"max_entries,omitempty" json:"max_entries,omitempty"`
	ImmediateEvents []string      `yaml:"immediate_events,omitempty" json:"immediate_events,omitempty"`
}

func (bc BatchConfig) MarshalJSON() ([]byte, error) {
	type alias struct {
		Enabled         *bool        `json:"enabled,omitempty"`
		Interval        jsonDuration `json:"interval,omitempty"`
		MaxEntries      int          `json:"max_entries,omitempty"`
		ImmediateEvents []string     `json:"immediate_events,omitempty"`
	}
	return json.Marshal(alias{bc.Enabled, jsonDuration(bc.Interval), bc.MaxEntries, bc.ImmediateEvents})
}

func (bc *BatchConfig) UnmarshalJSON(b []byte) error {
	type alias struct {
		Enabled         *bool        `json:"enabled,omitempty"`
		Interval        jsonDuration `json:"interval,omitempty"`
		MaxEntries      int          `json:"max_entries,omitempty"`
		ImmediateEvents []string     `json:"immediate_events,omitempty"`
	}
	var a alias
	if err := json.Unmarshal(b, &a); err != nil {
		return err
	}
	bc.Enabled = a.Enabled
	bc.Interval = time.Duration(a.Interval)
	bc.MaxEntries = a.MaxEntries
	bc.ImmediateEvents = a.ImmediateEvents
	return nil
}

// NotificationsConfig holds all notification provider configurations.
type NotificationsConfig struct {
	Batch    *BatchConfig    `yaml:"batch,omitempty" json:"batch,omitempty"`
	Telegram *TelegramConfig `yaml:"telegram,omitempty" json:"telegram,omitempty"`
	Discord  *DiscordConfig  `yaml:"discord,omitempty" json:"discord,omitempty"`
	Webhook  *WebhookConfig  `yaml:"webhook,omitempty" json:"webhook,omitempty"`
	Matrix   *MatrixConfig   `yaml:"matrix,omitempty" json:"matrix,omitempty"`
	Pushover *PushoverConfig `yaml:"pushover,omitempty" json:"pushover,omitempty"`
	Gotify   *GotifyConfig   `yaml:"gotify,omitempty" json:"gotify,omitempty"`
}

// TelegramConfig holds Telegram notification settings.
type TelegramConfig struct {
	Enabled             bool         `yaml:"enabled" json:"enabled"`
	Token               string       `yaml:"token,omitempty" json:"token,omitempty"`
	ChatID              string       `yaml:"chat_id,omitempty" json:"chat_id,omitempty"`
	Events              []string     `yaml:"events" json:"events"`
	DisableNotification bool         `yaml:"disable_notification" json:"disable_notification"`
	Batch               *BatchConfig `yaml:"batch,omitempty" json:"batch,omitempty"`
}

// DiscordConfig holds Discord notification settings.
type DiscordConfig struct {
	Enabled    bool         `yaml:"enabled" json:"enabled"`
	WebhookURL string       `yaml:"webhook_url,omitempty" json:"webhook_url,omitempty"`
	Events     []string     `yaml:"events" json:"events"`
	Batch      *BatchConfig `yaml:"batch,omitempty" json:"batch,omitempty"`
}

// WebhookConfig holds generic webhook notification settings.
type WebhookConfig struct {
	Enabled  bool         `yaml:"enabled" json:"enabled"`
	Endpoint string       `yaml:"endpoint,omitempty" json:"endpoint,omitempty"`
	Method   string       `yaml:"method" json:"method"`
	Events   []string     `yaml:"events" json:"events"`
	Batch    *BatchConfig `yaml:"batch,omitempty" json:"batch,omitempty"`
}

// MatrixConfig holds Matrix notification settings.
type MatrixConfig struct {
	Enabled     bool         `yaml:"enabled" json:"enabled"`
	Homeserver  string       `yaml:"homeserver,omitempty" json:"homeserver,omitempty"`
	RoomID      string       `yaml:"room_id,omitempty" json:"room_id,omitempty"`
	AccessToken string       `yaml:"access_token,omitempty" json:"access_token,omitempty"`
	Events      []string     `yaml:"events" json:"events"`
	Batch       *BatchConfig `yaml:"batch,omitempty" json:"batch,omitempty"`
}

// PushoverConfig holds Pushover notification settings.
type PushoverConfig struct {
	Enabled  bool         `yaml:"enabled" json:"enabled"`
	UserKey  string       `yaml:"user_key,omitempty" json:"user_key,omitempty"`
	APIToken string       `yaml:"api_token,omitempty" json:"api_token,omitempty"`
	Events   []string     `yaml:"events" json:"events"`
	Batch    *BatchConfig `yaml:"batch,omitempty" json:"batch,omitempty"`
}

// GotifyConfig holds Gotify notification settings.
type GotifyConfig struct {
	Enabled bool         `yaml:"enabled" json:"enabled"`
	URL     string       `yaml:"url,omitempty" json:"url,omitempty"`
	Token   string       `yaml:"token,omitempty" json:"token,omitempty"`
	Events  []string     `yaml:"events" json:"events"`
	Batch   *BatchConfig `yaml:"batch,omitempty" json:"batch,omitempty"`
}

// ResolveBatchConfig merges a provider-level BatchConfig with the global defaults.
// Provider fields take precedence; nil/zero fields fall back to global.
func ResolveBatchConfig(global, provider *BatchConfig) *BatchConfig {
	if global == nil && provider == nil {
		return nil
	}

	resolved := &BatchConfig{}

	// Start from global
	if global != nil {
		resolved.Enabled = global.Enabled
		resolved.Interval = global.Interval
		resolved.MaxEntries = global.MaxEntries
		resolved.ImmediateEvents = global.ImmediateEvents
	}

	// Override with provider-specific values
	if provider != nil {
		if provider.Enabled != nil {
			resolved.Enabled = provider.Enabled
		}
		if provider.Interval != 0 {
			resolved.Interval = provider.Interval
		}
		if provider.MaxEntries != 0 {
			resolved.MaxEntries = provider.MaxEntries
		}
		if len(provider.ImmediateEvents) > 0 {
			resolved.ImmediateEvents = provider.ImmediateEvents
		}
	}

	return resolved
}

// IsBatchEnabled returns whether batching is enabled in this config.
func (bc *BatchConfig) IsBatchEnabled() bool {
	if bc == nil || bc.Enabled == nil {
		return false
	}
	return *bc.Enabled
}

// ToStreamerSettings converts a StreamerSettingsConfig to a model.StreamerSettings,
// using defaults for any unset fields.
func (ssc *StreamerSettingsConfig) ToStreamerSettings(defaults *model.StreamerSettings) *model.StreamerSettings {
	settings := *defaults // copy defaults

	if ssc == nil {
		return &settings
	}

	if ssc.MakePredictions != nil {
		settings.MakePredictions = *ssc.MakePredictions
	}
	if ssc.FollowRaid != nil {
		settings.FollowRaid = *ssc.FollowRaid
	}
	if ssc.ClaimDrops != nil {
		settings.ClaimDrops = *ssc.ClaimDrops
	}
	if ssc.ClaimMoments != nil {
		settings.ClaimMoments = *ssc.ClaimMoments
	}
	if ssc.WatchStreak != nil {
		settings.WatchStreak = *ssc.WatchStreak
	}
	if ssc.CommunityGoals != nil {
		settings.CommunityGoalsEnabled = *ssc.CommunityGoals
	}
	if ssc.DropsOnly != nil {
		settings.DropsOnly = *ssc.DropsOnly
	}
	if ssc.Chat != "" {
		settings.Chat = model.ParseChatPresence(ssc.Chat)
	}
	if ssc.Bet != nil {
		settings.Bet = ssc.Bet.ToBetSettings(defaults.Bet)
	}

	return &settings
}

// ToBetSettings converts a BetSettingsConfig to a model.BetSettings,
// using defaults for any unset fields.
func (bsc *BetSettingsConfig) ToBetSettings(defaults *model.BetSettings) *model.BetSettings {
	betSettings := *defaults // copy defaults

	if bsc == nil {
		return &betSettings
	}

	if bsc.Strategy != "" {
		betSettings.Strategy = model.ParseStrategy(bsc.Strategy)
	}
	if bsc.Percentage != nil {
		betSettings.Percentage = *bsc.Percentage
	}
	if bsc.PercentageGap != nil {
		betSettings.PercentageGap = *bsc.PercentageGap
	}
	if bsc.MaxPoints != nil {
		betSettings.MaxPoints = *bsc.MaxPoints
	}
	if bsc.MinimumPoints != nil {
		betSettings.MinimumPoints = *bsc.MinimumPoints
	}
	if bsc.StealthMode != nil {
		betSettings.StealthMode = *bsc.StealthMode
	}
	if bsc.Delay != nil {
		betSettings.Delay = *bsc.Delay
	}
	if bsc.DelayMode != "" {
		betSettings.DelayMode = model.ParseDelayMode(bsc.DelayMode)
	}
	if bsc.FilterCondition != nil {
		betSettings.FilterCondition = &model.FilterCondition{
			By:    model.OutcomeKey(bsc.FilterCondition.By),
			Where: model.ParseCondition(bsc.FilterCondition.Where),
			Value: bsc.FilterCondition.Value,
		}
	}

	return &betSettings
}

// IsEnabled returns whether this account is enabled.
// If the Enabled field is not set (nil), it defaults to true.
func (ac *AccountConfig) IsEnabled() bool {
	if ac.Enabled == nil {
		return true // default to true when not specified
	}
	return *ac.Enabled
}

// ParsedPriorities converts the string priority list to model.Priority values.
func (ac *AccountConfig) ParsedPriorities() []model.Priority {
	priorities := make([]model.Priority, 0, len(ac.Priority))
	for _, priorityStr := range ac.Priority {
		priorities = append(priorities, model.ParsePriority(priorityStr))
	}
	return priorities
}
