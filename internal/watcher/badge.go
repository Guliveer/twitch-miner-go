package watcher

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Guliveer/twitch-miner-go/internal/config"
	"github.com/Guliveer/twitch-miner-go/internal/gql"
	"github.com/Guliveer/twitch-miner-go/internal/logger"
	"github.com/Guliveer/twitch-miner-go/internal/model"
)

const (
	badgeCatalogCacheTTL = 15 * time.Minute
	// Twitch's directory query accepts at most 100 streams. AllChannels badge
	// campaigns do not require the Drops Enabled tag, so search the broadest
	// eligible page and prefer its highest-viewer candidate.
	badgeCategoryStreamLimit = 100

	// Twitch uses Special Events as a synthetic category for campaigns whose
	// allow-listed channels may stream in any real category.
	crossCategoryBadgeCampaignSlug = "special-events"
)

type badgeGQL interface {
	GetAvailableBadgeNames(context.Context) (map[string]struct{}, error)
	GetDropsInventory(context.Context) (json.RawMessage, error)
	GetStreamInfo(context.Context, string) (*gql.StreamInfoResponse, error)
	GetUserID(context.Context, string) (string, error)
	GetTopStreamsByCategory(context.Context, string, int, bool) ([]gql.TopStream, error)
	GetAvailableCampaigns(context.Context, string) ([]string, error)
}

type badgeTracked struct {
	username string
	added    bool
}

type badgeCandidate struct {
	stream      *gql.TopStream
	campaignIDs []string
}

type badgeCampaignLookup struct {
	ids []string
	err error
}

// BadgeWatcher discovers active watch-time chat badge Drops and keeps a live,
// eligible channel in the miner until Twitch reports the badge/campaign earned.
type BadgeWatcher struct {
	mu              sync.Mutex
	channelIDsMu    sync.Mutex
	cfg             config.BadgeWatcherConfig
	gql             badgeGQL
	httpClient      *http.Client
	log             *logger.Logger
	blacklist       map[string]bool
	defaults        *model.StreamerSettings
	tracked         map[string]badgeTracked
	channelIDs      map[string]string
	campaigns       []badgeCampaign
	catalogLoadedAt time.Time
}

func NewBadgeWatcher(cfg config.BadgeWatcherConfig, gqlClient badgeGQL, httpClient *http.Client, log *logger.Logger, blacklist []string, defaults *model.StreamerSettings) *BadgeWatcher {
	b := make(map[string]bool)
	for _, name := range blacklist {
		b[strings.ToLower(name)] = true
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 20 * time.Second}
	}
	return &BadgeWatcher{cfg: cfg, gql: gqlClient, httpClient: httpClient, log: log, blacklist: b, defaults: defaults, tracked: make(map[string]badgeTracked), channelIDs: make(map[string]string)}
}

func (bw *BadgeWatcher) Run(ctx context.Context, add func(context.Context, *model.Streamer), remove func(string, string), get func() []*model.Streamer) error {
	return pollLoop(ctx, bw.log, bw.cfg.PollInterval, "🏅 BadgeWatcher started", "🏅 BadgeWatcher stopping", []any{"poll_interval", bw.cfg.PollInterval, "streamer_limit", bw.cfg.StreamerLimit}, func(ctx context.Context) { bw.evaluate(ctx, add, remove, get) }, func() { bw.cleanup(remove, get) })
}

func (bw *BadgeWatcher) evaluate(ctx context.Context, add func(context.Context, *model.Streamer), remove func(string, string), get func() []*model.Streamer) {
	campaigns, err := bw.loadCampaigns(ctx)
	if err != nil {
		bw.log.Warn("Failed to load badge campaign catalog", "error", err)
		return
	}
	owned, err := bw.gql.GetAvailableBadgeNames(ctx)
	if err != nil {
		bw.log.Warn("Failed to load earned Twitch badges", "error", err)
		return
	}
	raw, err := bw.gql.GetDropsInventory(ctx)
	if err != nil {
		bw.log.Warn("Failed to load Drops inventory for badges", "error", err)
		return
	}
	completed, err := completedCampaigns(raw)
	if err != nil {
		bw.log.Warn("Failed to parse Drops inventory for badges", "error", err)
		return
	}
	eligible := make(map[string]badgeCampaign)
	eligibleOrdered := make([]badgeCampaign, 0, len(campaigns))
	for _, c := range campaigns {
		if c.EndsAt.After(time.Now()) && !ownsCampaignBadge(c, owned) && !campaignCompleted(c, completed) {
			eligible[c.ID] = c
			eligibleOrdered = append(eligibleOrdered, c)
		}
	}
	sort.SliceStable(eligibleOrdered, func(i, j int) bool {
		if eligibleOrdered[i].EndsAt.Equal(eligibleOrdered[j].EndsAt) {
			return eligibleOrdered[i].Name < eligibleOrdered[j].Name
		}
		return eligibleOrdered[i].EndsAt.Before(eligibleOrdered[j].EndsAt)
	})

	// Retire completed campaigns and stale channels before filling free slots.
	bw.mu.Lock()
	for id, tr := range bw.tracked {
		campaign, eligibleNow := eligible[id]
		streamer := findStreamer(get(), tr.username)
		valid := eligibleNow && streamer != nil && badgeStreamerValid(streamer, campaign)
		if valid {
			continue
		}
		if tr.added {
			reason := "badge_stream_no_longer_eligible"
			if !eligibleNow {
				reason = "badge_campaign_completed_or_expired"
			}
			remove(tr.username, reason)
		} else {
			unmarkBadge(get(), tr.username)
		}
		delete(bw.tracked, id)
	}
	used := len(bw.tracked)
	bw.mu.Unlock()
	limit := bw.cfg.StreamerLimit
	if limit < 1 {
		limit = 1
	} else if limit > 2 {
		limit = 2
	}
	bw.mu.Lock()
	reserved := make(map[string]bool, len(bw.tracked))
	for _, tr := range bw.tracked {
		reserved[strings.ToLower(tr.username)] = true
	}
	bw.mu.Unlock()
	bw.log.Info("🏅 Badge campaigns evaluated", "catalog", len(campaigns), "eligible", len(eligibleOrdered), "active", used, "owned_badges", len(owned), "completed_campaigns", len(completed))
	campaignLookupCache := make(map[string]badgeCampaignLookup)
	for _, c := range eligibleOrdered {
		id := c.ID
		if used >= limit || ctx.Err() != nil {
			break
		}
		bw.mu.Lock()
		_, exists := bw.tracked[id]
		bw.mu.Unlock()
		if exists {
			continue
		}
		candidate, err := bw.findCampaignCandidate(ctx, c, reserved, campaignLookupCache)
		if err != nil {
			bw.log.Warn("Failed to find badge stream", "campaign", c.Name, "category", c.GameSlug, "error", err)
			continue
		}
		if candidate == nil {
			bw.log.Info("No live eligible stream for badge campaign", "campaign", c.Name, "category", c.GameSlug, "all_channels", c.AllChannels)
			continue
		}
		if existing := findStreamer(get(), candidate.stream.Username); existing != nil {
			existing.Mu.Lock()
			existing.IsBadgeWatched = true
			existing.BadgeCampaign = c.Name
			if len(candidate.campaignIDs) > 0 {
				existing.Stream.CampaignIDs = append([]string(nil), candidate.campaignIDs...)
			}
			existing.Mu.Unlock()
			bw.mu.Lock()
			bw.tracked[id] = badgeTracked{candidate.stream.Username, false}
			bw.mu.Unlock()
			reserved[strings.ToLower(candidate.stream.Username)] = true
			used++
			bw.log.Info("🏅 Using existing streamer for badge campaign", "streamer", candidate.stream.Username, "campaign", c.Name, "category", c.GameSlug, "badges", strings.Join(c.BadgeNames, ", "))
			continue
		}
		s := bw.badgeStreamer(ctx, candidate.stream, c, candidate.campaignIDs)
		add(ctx, s)
		bw.mu.Lock()
		bw.tracked[id] = badgeTracked{candidate.stream.Username, true}
		bw.mu.Unlock()
		reserved[strings.ToLower(candidate.stream.Username)] = true
		used++
		bw.log.Info("🏅 Discovered badge campaign stream", "streamer", candidate.stream.Username, "campaign", c.Name, "category", c.GameSlug, "badges", strings.Join(c.BadgeNames, ", "))
	}
}

func (bw *BadgeWatcher) badgeStreamer(ctx context.Context, candidate *gql.TopStream, campaign badgeCampaign, campaignIDs []string) *model.Streamer {
	s := model.NewStreamer(candidate.Username)
	s.ChannelID = candidate.ChannelID
	s.DisplayName = candidate.DisplayName
	s.IsOnline = true
	s.OnlineAt = time.Now()
	s.IsBadgeWatched = true
	s.BadgeCampaign = campaign.Name
	if candidate.GameID != "" || candidate.GameSlug != "" || candidate.GameName != "" {
		gameSlug := candidate.GameSlug
		if gameSlug == "" {
			gameSlug = campaign.GameSlug
		}
		s.Stream.Game = &model.GameInfo{ID: candidate.GameID, Slug: gameSlug, Name: candidate.GameName}
	}
	s.Stream.ViewersCount = candidate.ViewersCount
	if campaignIDs != nil {
		s.Stream.CampaignIDs = append([]string(nil), campaignIDs...)
	} else if ids, err := bw.gql.GetAvailableCampaigns(ctx, candidate.ChannelID); err == nil {
		s.Stream.CampaignIDs = ids
	} else {
		bw.log.Warn("Failed to load campaigns for badge streamer", "streamer", candidate.Username, "error", err)
	}
	settings := *bw.defaults
	if settings.Bet != nil {
		copyBet := *settings.Bet
		settings.Bet = &copyBet
	}
	settings.FollowRaid = false
	settings.ClaimDrops = true
	// A restricted campaign's explicit allow-list is authoritative even when
	// Twitch's generic Drops endpoint does not expose campaign IDs for it.
	settings.DropsOnly = campaign.AllChannels
	settings.Chat = model.ChatNever
	s.Settings = &settings
	return s
}

func (bw *BadgeWatcher) findCampaignCandidate(ctx context.Context, campaign badgeCampaign, reserved map[string]bool, lookupCache map[string]badgeCampaignLookup) (*badgeCandidate, error) {
	if !campaign.AllChannels {
		streams, err := bw.findCampaignStreams(ctx, campaign)
		if err != nil {
			return nil, err
		}
		allowed := make(map[string]bool, len(campaign.Channels))
		for _, name := range campaign.Channels {
			allowed[strings.ToLower(strings.TrimSpace(name))] = true
		}
		candidate := bw.pickCandidate(streams, reserved, allowed, false)
		if candidate == nil {
			return nil, nil
		}
		return &badgeCandidate{stream: candidate}, nil
	}
	lookupFailures := 0
	var lastLookupErr error
	var directoryErrors []error
	for _, dropsOnly := range []bool{true, false} {
		streams, err := bw.gql.GetTopStreamsByCategory(ctx, campaign.GameSlug, badgeCategoryStreamLimit, dropsOnly)
		if err != nil {
			directoryErrors = append(directoryErrors, err)
			continue
		}
		candidate, failures, lookupErr := bw.pickEligibleCampaignCandidate(ctx, streams, reserved, lookupCache)
		lookupFailures += failures
		if lookupErr != nil {
			lastLookupErr = lookupErr
		}
		if candidate != nil {
			bw.logCampaignCandidateWarnings(campaign, lookupFailures, lastLookupErr, directoryErrors)
			return candidate, nil
		}
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
	}
	bw.logCampaignCandidateWarnings(campaign, lookupFailures, lastLookupErr, directoryErrors)
	if len(directoryErrors) == 2 {
		return nil, fmt.Errorf("both badge category searches failed: %v; %v", directoryErrors[0], directoryErrors[1])
	}
	return nil, nil
}

func (bw *BadgeWatcher) logCampaignCandidateWarnings(campaign badgeCampaign, lookupFailures int, lastLookupErr error, directoryErrors []error) {
	if lookupFailures > 0 {
		bw.log.Warn("Failed to verify some badge campaign candidates", "campaign", campaign.Name, "failures", lookupFailures, "last_error", lastLookupErr)
	}
	if len(directoryErrors) == 1 {
		bw.log.Warn("One badge category search failed", "campaign", campaign.Name, "error", directoryErrors[0])
	}
}

func (bw *BadgeWatcher) pickEligibleCampaignCandidate(ctx context.Context, streams []gql.TopStream, reserved map[string]bool, cache map[string]badgeCampaignLookup) (*badgeCandidate, int, error) {
	failures := 0
	var lastErr error
	for i := range streams {
		if ctx.Err() != nil {
			return nil, failures, lastErr
		}
		stream := &streams[i]
		login := strings.ToLower(stream.Username)
		if bw.blacklist[login] || reserved[login] || stream.ChannelID == "" {
			continue
		}
		lookup, found := cache[stream.ChannelID]
		if !found {
			lookup.ids, lookup.err = bw.gql.GetAvailableCampaigns(ctx, stream.ChannelID)
			cache[stream.ChannelID] = lookup
			if lookup.err != nil {
				failures++
				lastErr = lookup.err
			}
		}
		if lookup.err != nil {
			continue
		}
		// Catalog IDs (such as twitchdrops-app-*) and Twitch campaign IDs are
		// different namespaces. The catalog identifies the badge campaign for
		// this game; Twitch confirms that the channel exposes active Drops.
		if len(lookup.ids) > 0 {
			return &badgeCandidate{stream: stream, campaignIDs: lookup.ids}, failures, lastErr
		}
	}
	return nil, failures, lastErr
}

func (bw *BadgeWatcher) findCampaignStreams(ctx context.Context, campaign badgeCampaign) ([]gql.TopStream, error) {
	if campaign.AllChannels {
		return bw.gql.GetTopStreamsByCategory(ctx, campaign.GameSlug, badgeCategoryStreamLimit, false)
	}

	streams := make([]gql.TopStream, 0, len(campaign.Channels))
	for _, username := range campaign.Channels {
		displayName := strings.TrimSpace(username)
		login := strings.ToLower(displayName)
		if login == "" || bw.blacklist[login] {
			continue
		}
		stream, err := bw.restrictedCampaignStream(ctx, campaign, login, displayName)
		if err != nil {
			bw.log.Warn("Failed to check restricted badge channel", "campaign", campaign.Name, "streamer", login, "error", err)
			continue
		}
		if stream != nil {
			streams = append(streams, *stream)
		}
	}
	sort.SliceStable(streams, func(i, j int) bool { return streams[i].ViewersCount > streams[j].ViewersCount })
	return streams, nil
}

func (bw *BadgeWatcher) restrictedCampaignStream(ctx context.Context, campaign badgeCampaign, login, displayName string) (*gql.TopStream, error) {
	info, err := bw.gql.GetStreamInfo(ctx, login)
	if err != nil || info == nil {
		return nil, err
	}
	if !badgeCampaignGameMatches(info.Game, campaign) {
		return nil, nil
	}
	channelID, err := bw.channelID(ctx, login)
	if err != nil {
		return nil, err
	}
	stream := &gql.TopStream{Username: login, ChannelID: channelID, DisplayName: displayName, ViewersCount: info.ViewersCount}
	if info.Game != nil {
		stream.GameID = info.Game.ID
		stream.GameName = info.Game.Name
		stream.GameSlug = info.Game.Slug
	}
	return stream, nil
}

func (bw *BadgeWatcher) channelID(ctx context.Context, login string) (string, error) {
	bw.channelIDsMu.Lock()
	channelID := bw.channelIDs[login]
	bw.channelIDsMu.Unlock()
	if channelID != "" {
		return channelID, nil
	}
	channelID, err := bw.gql.GetUserID(ctx, login)
	if err != nil {
		return "", err
	}
	bw.channelIDsMu.Lock()
	bw.channelIDs[login] = channelID
	bw.channelIDsMu.Unlock()
	return channelID, nil
}

func (bw *BadgeWatcher) pickCandidate(streams []gql.TopStream, reserved, allowed map[string]bool, allChannels bool) *gql.TopStream {
	for i := range streams {
		login := strings.ToLower(streams[i].Username)
		if !bw.blacklist[login] && !reserved[login] && (allChannels || allowed[login]) {
			return &streams[i]
		}
	}
	return nil
}

func (bw *BadgeWatcher) loadCampaigns(ctx context.Context) ([]badgeCampaign, error) {
	if len(bw.campaigns) > 0 && time.Since(bw.catalogLoadedAt) < badgeCatalogCacheTTL {
		return bw.campaigns, nil
	}
	campaigns, err := loadBadgeCampaigns(ctx, bw.httpClient, bw.cfg.DropsCatalogURL, bw.cfg.BadgesCatalogURL, time.Now(), func(campaign, reward, reason string, matches []string) {
		fields := []any{"campaign", campaign, "reason", reason}
		if reward != "" {
			fields = append(fields, "reward", reward)
		}
		if len(matches) > 0 {
			fields = append(fields, "badge_matches", strings.Join(matches, ", "))
		}
		bw.log.Warn("Badge catalog watch reward skipped", fields...)
	})
	if err != nil {
		if len(bw.campaigns) > 0 {
			bw.log.Warn("Badge catalogs unavailable; using cached campaigns", "error", err, "cache_age", time.Since(bw.catalogLoadedAt).Round(time.Second))
			return bw.campaigns, nil
		}
		return nil, err
	}
	bw.campaigns, bw.catalogLoadedAt = campaigns, time.Now()
	return campaigns, nil
}

func badgeStreamerValid(s *model.Streamer, campaign badgeCampaign) bool {
	s.Mu.RLock()
	defer s.Mu.RUnlock()
	if !s.IsOnline || s.Stream == nil {
		return false
	}
	if !campaign.AllChannels && !badgeCampaignAllowsChannel(campaign, s.Username) {
		return false
	}
	// CampaignIDs are Twitch UUIDs, whereas campaign.ID comes from the
	// external catalog. An all-channel streamer is valid when Twitch reports
	// at least one active Drops campaign on it.
	if campaign.AllChannels && len(s.Stream.CampaignIDs) == 0 {
		return false
	}
	return badgeCampaignGameMatches(s.Stream.Game, campaign)
}

func badgeCampaignAllowsChannel(campaign badgeCampaign, username string) bool {
	for _, allowed := range campaign.Channels {
		if strings.EqualFold(strings.TrimSpace(allowed), username) {
			return true
		}
	}
	return false
}

func badgeCampaignGameMatches(game *model.GameInfo, campaign badgeCampaign) bool {
	if strings.EqualFold(campaign.GameSlug, crossCategoryBadgeCampaignSlug) {
		return true
	}
	if game == nil {
		return false
	}
	return strings.EqualFold(game.Slug, campaign.GameSlug) || strings.EqualFold(slugify(game.Name), campaign.GameSlug)
}

func findStreamer(streamers []*model.Streamer, username string) *model.Streamer {
	for _, s := range streamers {
		if strings.EqualFold(s.Username, username) {
			return s
		}
	}
	return nil
}
func unmarkBadge(streamers []*model.Streamer, username string) {
	if s := findStreamer(streamers, username); s != nil {
		s.Mu.Lock()
		s.IsBadgeWatched = false
		s.BadgeCampaign = ""
		s.Mu.Unlock()
	}
}
func (bw *BadgeWatcher) cleanup(remove func(string, string), get func() []*model.Streamer) {
	bw.mu.Lock()
	defer bw.mu.Unlock()
	for id, tr := range bw.tracked {
		if tr.added {
			remove(tr.username, "badge_watcher_shutdown")
		} else {
			unmarkBadge(get(), tr.username)
		}
		delete(bw.tracked, id)
	}
}
