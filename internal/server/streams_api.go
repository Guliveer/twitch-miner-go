package server

import (
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Guliveer/twitch-miner-go/internal/logger"
	"github.com/Guliveer/twitch-miner-go/internal/model"
)

// ---- response helpers ----

type errorResponse struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.Encode(data) //nolint:errcheck // JSON to ResponseWriter; error not actionable
}

// writeInternalError logs the full error server-side and returns a generic
// error message to the client. This prevents leaking internal details
// (database errors, file paths, etc.) to API consumers.
func writeInternalError(w http.ResponseWriter, log *logger.Logger, route string, err error) {
	log.Error("API error", "route", route, "error", err)
	writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
}

// ---- pagination ----

type paginationParams struct {
	Offset int
	Limit  int
}

func parsePagination(r *http.Request) paginationParams {
	p := paginationParams{}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			p.Offset = n
		}
	}
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			p.Limit = n
		}
	}
	return p
}

type paginatedResponse struct {
	Total  int `json:"total"`
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
	Data   any `json:"data"`
}

func applyPagination[T any](items []T, p paginationParams) paginatedResponse {
	total := len(items)
	if p.Offset > total {
		p.Offset = total
	}
	items = items[p.Offset:]
	if p.Limit > 0 && p.Limit < len(items) {
		items = items[:p.Limit]
	}
	return paginatedResponse{
		Total:  total,
		Offset: p.Offset,
		Limit:  p.Limit,
		Data:   items,
	}
}

// ---- streamer types ----

type streamerSummary struct {
	Account           string              `json:"account"`
	Username          string              `json:"username"`
	DisplayName       string              `json:"display_name,omitempty"`
	ChannelID         string              `json:"channel_id"`
	IsOnline          bool                `json:"is_online"`
	IsCategoryWatched bool                `json:"is_category_watched"`
	ChannelPoints     int                 `json:"channel_points"`
	StreamerURL       string              `json:"streamer_url"`
	Game              string              `json:"game,omitempty"`
	ViewersCount      int                 `json:"viewers_count"`
	Title             string              `json:"title,omitempty"`
	ActiveDrops       []dropProgressShort `json:"active_drops,omitempty"`
}

type dropProgressShort struct {
	Game            string `json:"game,omitempty"`
	Reward          string `json:"reward,omitempty"`
	ProgressPercent int    `json:"progress_percent"`
	WatchedMinutes  int    `json:"watched_minutes"`
	RequiredMinutes int    `json:"required_minutes"`
	IsClaimable     bool   `json:"is_claimable"`
	IsClaimed       bool   `json:"is_claimed"`
}

type streamerDetail struct {
	Account           string                         `json:"account"`
	Username          string                         `json:"username"`
	DisplayName       string                         `json:"display_name,omitempty"`
	ChannelID         string                         `json:"channel_id"`
	IsOnline          bool                           `json:"is_online"`
	IsCategoryWatched bool                           `json:"is_category_watched"`
	CategorySlug      string                         `json:"category_slug,omitempty"`
	ChannelPoints     int                            `json:"channel_points"`
	StreamerURL       string                         `json:"streamer_url"`
	ViewerIsMod       bool                           `json:"viewer_is_mod"`
	Stream            *streamInfo                    `json:"stream,omitempty"`
	Multipliers       []float64                      `json:"multipliers,omitempty"`
	History           map[string]*model.HistoryEntry `json:"history,omitempty"`
	DropCampaigns     []campaignInfo                 `json:"drop_campaigns,omitempty"`
}

type streamInfo struct {
	BroadcastID  string `json:"broadcast_id,omitempty"`
	Title        string `json:"title,omitempty"`
	Game         string `json:"game,omitempty"`
	ViewersCount int    `json:"viewers_count"`
	HasDropsTag  bool   `json:"drops_tags"`
}

type campaignInfo struct {
	Game     string     `json:"game,omitempty"`
	Name     string     `json:"name,omitempty"`
	Status   string     `json:"status,omitempty"`
	EndAt    string     `json:"end_at,omitempty"`
	Drops    []dropInfo `json:"drops,omitempty"`
}

type dropInfo struct {
	Name            string `json:"name,omitempty"`
	Benefit         string `json:"benefit,omitempty"`
	ProgressPercent int    `json:"progress_percent"`
	WatchedMinutes  int    `json:"watched_minutes"`
	RequiredMinutes int    `json:"required_minutes"`
	IsClaimable     bool   `json:"is_claimable"`
	IsClaimed       bool   `json:"is_claimed"`
}

// ---- stats types ----

type overallStats struct {
	TotalStreamers  int                         `json:"total_streamers"`
	OnlineStreamers int                         `json:"online_streamers"`
	TotalPoints     int                         `json:"total_points"`
	History         map[string]historyAggregate `json:"history"`
}

type historyAggregate struct {
	Counter int `json:"counter"`
	Amount  int `json:"amount"`
}

// ---- event types ----

type eventLogEntry struct {
	Account  string `json:"account"`
	Streamer string `json:"streamer"`
	Event    string `json:"event"`
	Count    int    `json:"count"`
	Amount   int    `json:"amount"`
}

// eventCategories groups event reasons into UI-filterable buckets.
var eventCategories = map[string][]string{
	"drops":   {"DROP_CLAIM", "DROP_CLAIM_AVAILABLE", "DROP_STATUS", "DROP_MILESTONE"},
	"points":  {"GAIN_FOR_WATCH", "GAIN_FOR_WATCH_STREAK", "GAIN_FOR_CLAIM", "GAIN_FOR_RAID", "BONUS_CLAIM"},
	"bets":    {"BET_START", "BET_WIN", "BET_LOSE", "BET_REFUND", "BET_FILTERS", "BET_GENERAL", "BET_FAILED"},
	"raids":   {"JOIN_RAID"},
	"streams": {"STREAMER_ONLINE", "STREAMER_OFFLINE"},
	"other":   {"MOMENT_CLAIM", "CHAT_MENTION", "GIFTED_SUB"},
}

// ---- helpers ----

func sortedKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// filterStreamers applies query-parameter filters to a streamer list.
func filterStreamers(streamers []*model.Streamer, r *http.Request) []*model.Streamer {
	accountFilter := strings.ToLower(r.URL.Query().Get("account"))
	channelFilter := strings.ToLower(r.URL.Query().Get("channel"))
	categoryFilter := strings.ToLower(r.URL.Query().Get("category"))
	onlineFilter := r.URL.Query().Get("online")

	if accountFilter == "" && channelFilter == "" && categoryFilter == "" && onlineFilter == "" {
		return streamers
	}

	filtered := make([]*model.Streamer, 0, len(streamers))
	for _, st := range streamers {
		st.Mu.RLock()

		if accountFilter != "" && !strings.EqualFold(st.AccountUsername, accountFilter) {
			st.Mu.RUnlock()
			continue
		}
		if channelFilter != "" {
			usernameMatch := strings.Contains(strings.ToLower(st.Username), channelFilter)
			displayNameMatch := strings.Contains(strings.ToLower(st.DisplayName), channelFilter)
			if !usernameMatch && !displayNameMatch {
				st.Mu.RUnlock()
				continue
			}
		}
		if categoryFilter != "" {
			gameName := ""
			if st.Stream != nil && st.Stream.Game != nil {
				gameName = strings.ToLower(st.Stream.Game.DisplayName)
			}
			if !strings.Contains(gameName, categoryFilter) {
				st.Mu.RUnlock()
				continue
			}
		}
		if onlineFilter != "" {
			wantOnline := onlineFilter == "true"
			if st.IsOnline != wantOnline {
				st.Mu.RUnlock()
				continue
			}
		}

		st.Mu.RUnlock()
		filtered = append(filtered, st)
	}
	return filtered
}

// ---- handlers ----

func (s *AnalyticsServer) handleDebug(w http.ResponseWriter, _ *http.Request) {
	s.mu.RLock()
	fn := s.debugFunc
	s.mu.RUnlock()

	if fn == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "debug snapshot is not configured"})
		return
	}

	writeJSON(w, http.StatusOK, fn())
}

func (s *AnalyticsServer) handleAuthStatus(w http.ResponseWriter, r *http.Request) {
	username := strings.ToLower(r.PathValue("username"))
	if username == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "missing username"})
		return
	}

	s.mu.RLock()
	fn := s.authStatusFunc
	s.mu.RUnlock()

	if fn == nil {
		writeJSON(w, http.StatusOK, map[string]string{"status": "no_pending"})
		return
	}

	result := fn(username)
	if result == nil {
		writeJSON(w, http.StatusOK, map[string]string{"status": "no_pending"})
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (s *AnalyticsServer) handleStreamers(w http.ResponseWriter, r *http.Request) {
	streamers := filterStreamers(s.getStreamers(), r)
	result := make([]streamerSummary, 0, len(streamers))

	for _, streamer := range streamers {
		streamer.Mu.RLock()
		summary := streamerSummary{
			Account:           streamer.AccountUsername,
			Username:          streamer.Username,
			DisplayName:       streamer.DisplayName,
			ChannelID:         streamer.ChannelID,
			IsOnline:          streamer.IsOnline,
			IsCategoryWatched: streamer.IsCategoryWatched,
			ChannelPoints:     streamer.ChannelPoints,
			StreamerURL:       streamer.StreamerURL,
		}
		if streamer.Stream != nil && streamer.Stream.Game != nil {
			summary.Game = streamer.Stream.Game.DisplayName
		}
		if streamer.Stream != nil {
			summary.ViewersCount = streamer.Stream.ViewersCount
			summary.Title = streamer.Stream.Title

			if len(streamer.Stream.Campaigns) > 0 {
				drops := make([]dropProgressShort, 0)
				for _, c := range streamer.Stream.Campaigns {
					gameName := ""
					if c.Game != nil {
						gameName = c.Game.DisplayName
					}
					for _, d := range c.Drops {
						if d == nil {
							continue
						}
						reward := d.Benefit
						if reward == "" {
							reward = d.Name
						}
						drops = append(drops, dropProgressShort{
							Game:            gameName,
							Reward:          reward,
							ProgressPercent: d.PercentageProgress,
							WatchedMinutes:  d.CurrentMinutesWatched,
							RequiredMinutes: d.MinutesRequired,
							IsClaimable:     d.IsClaimable,
							IsClaimed:       d.IsClaimed,
						})
					}
				}
				if len(drops) > 0 {
					summary.ActiveDrops = drops
				}
			}
		}
		streamer.Mu.RUnlock()
		result = append(result, summary)
	}

	sortBy := r.URL.Query().Get("sort")
	order := strings.ToLower(r.URL.Query().Get("order"))
	desc := order == "desc"

	switch sortBy {
	case "points":
		sort.Slice(result, func(i, j int) bool {
			if desc {
				return result[i].ChannelPoints > result[j].ChannelPoints
			}
			return result[i].ChannelPoints < result[j].ChannelPoints
		})
	case "viewers":
		sort.Slice(result, func(i, j int) bool {
			if desc {
				return result[i].ViewersCount > result[j].ViewersCount
			}
			return result[i].ViewersCount < result[j].ViewersCount
		})
	default:
		sort.Slice(result, func(i, j int) bool {
			if desc {
				return strings.ToLower(result[i].Username) > strings.ToLower(result[j].Username)
			}
			return strings.ToLower(result[i].Username) < strings.ToLower(result[j].Username)
		})
	}

	p := parsePagination(r)
	if p.Limit > 0 || p.Offset > 0 {
		writeJSON(w, http.StatusOK, applyPagination(result, p))
	} else {
		writeJSON(w, http.StatusOK, result)
	}
}

func (s *AnalyticsServer) handleStreamer(w http.ResponseWriter, r *http.Request) {
	name := strings.ToLower(r.PathValue("name"))
	if name == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "missing streamer name"})
		return
	}

	streamers := s.getStreamers()
	for _, streamer := range streamers {
		streamer.Mu.RLock()
		if strings.EqualFold(streamer.Username, name) {
			detail := streamerDetail{
				Account:           streamer.AccountUsername,
				Username:          streamer.Username,
				DisplayName:       streamer.DisplayName,
				ChannelID:         streamer.ChannelID,
				IsOnline:          streamer.IsOnline,
				IsCategoryWatched: streamer.IsCategoryWatched,
				CategorySlug:      streamer.CategorySlug,
				ChannelPoints:     streamer.ChannelPoints,
				StreamerURL:       streamer.StreamerURL,
				ViewerIsMod:       streamer.ViewerIsMod,
				History:           streamer.History,
			}
			if streamer.Stream != nil {
				detail.Stream = &streamInfo{
					BroadcastID:  streamer.Stream.BroadcastID,
					Title:        streamer.Stream.Title,
					ViewersCount: streamer.Stream.ViewersCount,
					HasDropsTag:  streamer.Stream.HasDropsTag,
				}
				if streamer.Stream.Game != nil {
					detail.Stream.Game = streamer.Stream.Game.DisplayName
				}
			}
			if len(streamer.ActiveMultipliers) > 0 {
				detail.Multipliers = make([]float64, 0, len(streamer.ActiveMultipliers))
				for _, m := range streamer.ActiveMultipliers {
					detail.Multipliers = append(detail.Multipliers, m.Factor)
				}
			}
			if streamer.Stream != nil && len(streamer.Stream.Campaigns) > 0 {
				detail.DropCampaigns = make([]campaignInfo, 0, len(streamer.Stream.Campaigns))
				for _, c := range streamer.Stream.Campaigns {
					ci := campaignInfo{
						Name:   c.Name,
						Status: c.Status,
						EndAt:  c.EndAt.Format(time.RFC3339),
						Drops:  make([]dropInfo, 0, len(c.Drops)),
					}
					if c.Game != nil {
						ci.Game = c.Game.DisplayName
					}
					for _, d := range c.Drops {
						if d == nil {
							continue
						}
						benefit := d.Benefit
						if benefit == "" {
							benefit = d.Name
						}
						ci.Drops = append(ci.Drops, dropInfo{
							Name:            d.Name,
							Benefit:         benefit,
							ProgressPercent: d.PercentageProgress,
							WatchedMinutes:  d.CurrentMinutesWatched,
							RequiredMinutes: d.MinutesRequired,
							IsClaimable:     d.IsClaimable,
							IsClaimed:       d.IsClaimed,
						})
					}
					if len(ci.Drops) > 0 {
						detail.DropCampaigns = append(detail.DropCampaigns, ci)
					}
				}
				if len(detail.DropCampaigns) == 0 {
					detail.DropCampaigns = nil
				}
			}
			streamer.Mu.RUnlock()
			writeJSON(w, http.StatusOK, detail)
			return
		}
		streamer.Mu.RUnlock()
	}

	writeJSON(w, http.StatusNotFound, errorResponse{Error: "streamer not found"})
}

func (s *AnalyticsServer) handleStats(w http.ResponseWriter, r *http.Request) {
	streamers := filterStreamers(s.getStreamers(), r)

	eventFilter := r.URL.Query().Get("event")
	var allowedEvents map[string]bool
	if eventFilter != "" {
		allowedEvents = make(map[string]bool)
		for _, e := range strings.Split(eventFilter, ",") {
			allowedEvents[strings.TrimSpace(strings.ToUpper(e))] = true
		}
	}

	stats := overallStats{
		TotalStreamers: len(streamers),
		History:        make(map[string]historyAggregate),
	}

	for _, streamer := range streamers {
		streamer.Mu.RLock()
		stats.TotalPoints += streamer.ChannelPoints
		if streamer.IsOnline {
			stats.OnlineStreamers++
		}
		for reason, entry := range streamer.History {
			if allowedEvents != nil && !allowedEvents[reason] {
				continue
			}
			agg := stats.History[reason]
			agg.Counter += entry.Counter
			agg.Amount += entry.Amount
			stats.History[reason] = agg
		}
		streamer.Mu.RUnlock()
	}

	writeJSON(w, http.StatusOK, stats)
}

func (s *AnalyticsServer) handleFilters(w http.ResponseWriter, _ *http.Request) {
	streamers := s.getStreamers()

	accountSet := make(map[string]bool)
	channelSet := make(map[string]bool)
	categorySet := make(map[string]bool)
	eventSet := make(map[string]bool)

	for _, st := range streamers {
		st.Mu.RLock()
		if st.AccountUsername != "" {
			accountSet[st.AccountUsername] = true
		}
		channelSet[st.Username] = true
		if st.Stream != nil && st.Stream.Game != nil && st.Stream.Game.DisplayName != "" {
			categorySet[st.Stream.Game.DisplayName] = true
		}
		for reason := range st.History {
			eventSet[reason] = true
		}
		st.Mu.RUnlock()
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"accounts":   sortedKeys(accountSet),
		"channels":   sortedKeys(channelSet),
		"categories": sortedKeys(categorySet),
		"events":     sortedKeys(eventSet),
	})
}

func (s *AnalyticsServer) handleEventLogs(w http.ResponseWriter, r *http.Request) {
	accountFilter := strings.ToLower(r.URL.Query().Get("account"))
	channelFilter := strings.ToLower(r.URL.Query().Get("channel"))
	categoryFilter := strings.ToLower(r.URL.Query().Get("category"))
	eventFilter := strings.ToUpper(r.URL.Query().Get("event"))

	var allowedEvents map[string]bool
	if categoryFilter != "" {
		if events, ok := eventCategories[categoryFilter]; ok {
			allowedEvents = make(map[string]bool, len(events))
			for _, e := range events {
				allowedEvents[e] = true
			}
		}
	}

	var entries []eventLogEntry
	for _, st := range s.getStreamers() {
		if accountFilter != "" && !strings.EqualFold(st.AccountUsername, accountFilter) {
			continue
		}

		st.Mu.RLock()
		streamerName := st.DisplayName
		if streamerName == "" {
			streamerName = st.Username
		}
		if channelFilter != "" {
			usernameMatch := strings.Contains(strings.ToLower(st.Username), channelFilter)
			displayNameMatch := strings.Contains(strings.ToLower(streamerName), channelFilter)
			if !usernameMatch && !displayNameMatch {
				st.Mu.RUnlock()
				continue
			}
		}
		for event, hist := range st.History {
			if eventFilter != "" && event != eventFilter {
				continue
			}
			if allowedEvents != nil && !allowedEvents[event] {
				continue
			}
			entries = append(entries, eventLogEntry{
				Account:  st.AccountUsername,
				Streamer: streamerName,
				Event:    event,
				Count:    hist.Counter,
				Amount:   hist.Amount,
			})
		}
		st.Mu.RUnlock()
	}

	if entries == nil {
		entries = []eventLogEntry{}
	}

	p := parsePagination(r)
	if p.Limit > 0 || p.Offset > 0 {
		writeJSON(w, http.StatusOK, applyPagination(entries, p))
	} else {
		writeJSON(w, http.StatusOK, entries)
	}
}

func (s *AnalyticsServer) handleEventFilters(w http.ResponseWriter, _ *http.Request) {
	accountSet := make(map[string]bool)
	channelSet := make(map[string]bool)
	eventSet := make(map[string]bool)

	for _, st := range s.getStreamers() {
		if st.AccountUsername != "" {
			accountSet[st.AccountUsername] = true
		}
		st.Mu.RLock()
		streamerName := st.DisplayName
		if streamerName == "" {
			streamerName = st.Username
		}
		channelSet[streamerName] = true
		for event := range st.History {
			eventSet[event] = true
		}
		st.Mu.RUnlock()
	}

	categories := make([]string, 0, len(eventCategories))
	for cat := range eventCategories {
		categories = append(categories, cat)
	}
	sort.Strings(categories)

	writeJSON(w, http.StatusOK, map[string]any{
		"accounts":   sortedKeys(accountSet),
		"channels":   sortedKeys(channelSet),
		"events":     sortedKeys(eventSet),
		"categories": categories,
	})
}

func (s *AnalyticsServer) handleTestNotification(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	fn := s.notifyTestFunc
	s.mu.RUnlock()

	if fn == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "no notification dispatchers configured"})
		return
	}

	errs := fn(r.Context())
	if len(errs) > 0 {
		errMsgs := make([]string, len(errs))
		for i, err := range errs {
			errMsgs[i] = err.Error()
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"status": "partial",
			"errors": errMsgs,
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"status":  "ok",
		"message": "Test notification sent to all enabled notifiers",
	})
}
