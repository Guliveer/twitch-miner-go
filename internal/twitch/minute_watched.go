package twitch

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Guliveer/twitch-miner-go/internal/constants"
	"github.com/Guliveer/twitch-miner-go/internal/model"
)

// SendMinuteWatchedEvents sends minute-watched events for the given streamers.
// Fix #3: Each streamer is processed concurrently with a per-streamer timeout
// to prevent slow HTTP requests for one streamer from blocking others or
// causing the 20-second ticker to miss ticks.
func (c *Client) SendMinuteWatchedEvents(ctx context.Context, streamers []*model.Streamer) error {
	httpClient := c.GQL.HTTPClient()

	var wg sync.WaitGroup
	for _, streamer := range streamers {
		if err := ctx.Err(); err != nil {
			return err
		}

		wg.Add(1)
		go func(s *model.Streamer) {
			defer wg.Done()

			// Per-streamer timeout to prevent one slow streamer from blocking the tick.
			sCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
			defer cancel()

			if err := c.sendMinuteWatchedForStreamer(sCtx, httpClient, s); err != nil {
				c.Log.Debug("Failed to send minute watched",
					"streamer", s.Username,
					"error", err)
			}
		}(streamer)
	}

	wg.Wait()
	return nil
}

func (c *Client) sendMinuteWatchedForStreamer(ctx context.Context, httpClient *http.Client, streamer *model.Streamer) error {
	// Stamped before anything can fail: stall detection compares this against
	// the last credit, so it must reflect every attempt, not just the ones
	// that got as far as the spade request.
	streamer.Mu.Lock()
	streamer.Stream.MarkMinuteWatchAttempt()
	streamer.Mu.Unlock()

	streamer.Mu.RLock()
	username := streamer.Username
	spadeURL := streamer.Stream.SpadeURL
	payload := streamer.Stream.Payload
	streamer.Mu.RUnlock()

	// Fix #1: If spade URL is empty (e.g. cache expired), attempt to refresh it
	// before giving up. This prevents silent failures after the TTL expires.
	if spadeURL == "" {
		c.Log.Warn("Spade URL empty, attempting refresh", "streamer", username)
		if err := c.RefreshSpadeURL(ctx, streamer); err != nil {
			c.Log.Warn("Failed to refresh spade URL", "streamer", username, "error", err)
			return fmt.Errorf("no spade URL for %s (refresh failed: %w)", username, err)
		}
		// Re-read after refresh.
		streamer.Mu.RLock()
		spadeURL = streamer.Stream.SpadeURL
		payload = streamer.Stream.Payload
		streamer.Mu.RUnlock()
		if spadeURL == "" {
			return fmt.Errorf("no spade URL for %s after refresh", username)
		}
	}
	if payload == nil {
		return fmt.Errorf("no payload for %s", username)
	}

	token, err := c.GQL.GetPlaybackAccessToken(ctx, username)
	if err != nil {
		return fmt.Errorf("getting playback access token for %s: %w", username, err)
	}

	c.Log.Debug("Got playback access token", "streamer", username)

	manifestURL := fmt.Sprintf(
		"https://usher.ttvnw.net/api/channel/hls/%s.m3u8?sig=%s&token=%s",
		username, token.Signature, token.Value,
	)

	manifestReq, err := http.NewRequestWithContext(ctx, http.MethodGet, manifestURL, nil)
	if err != nil {
		return fmt.Errorf("creating manifest request: %w", err)
	}
	manifestReq.Header.Set("User-Agent", constants.DefaultUserAgent)

	manifestResp, err := httpClient.Do(manifestReq)
	if err != nil {
		return fmt.Errorf("fetching manifest for %s: %w", username, err)
	}
	defer manifestResp.Body.Close()

	if manifestResp.StatusCode != http.StatusOK {
		return fmt.Errorf("manifest for %s returned status %d", username, manifestResp.StatusCode)
	}

	manifestBody, err := io.ReadAll(io.LimitReader(manifestResp.Body, 64<<10))
	if err != nil {
		return fmt.Errorf("reading manifest for %s: %w", username, err)
	}

	c.Log.Debug("Got HLS manifest", "streamer", username)

	lowestQualityURL := getLastURL(string(manifestBody))
	if lowestQualityURL == "" {
		return fmt.Errorf("no stream URL found in manifest for %s", username)
	}

	streamReq, err := http.NewRequestWithContext(ctx, http.MethodGet, lowestQualityURL, nil)
	if err != nil {
		return fmt.Errorf("creating stream request: %w", err)
	}
	streamReq.Header.Set("User-Agent", constants.DefaultUserAgent)

	streamResp, err := httpClient.Do(streamReq)
	if err != nil {
		return fmt.Errorf("fetching stream URL list for %s: %w", username, err)
	}
	defer streamResp.Body.Close()

	if streamResp.StatusCode != http.StatusOK {
		return fmt.Errorf("stream URL list for %s returned status %d", username, streamResp.StatusCode)
	}

	streamBody, err := io.ReadAll(io.LimitReader(streamResp.Body, 64<<10))
	if err != nil {
		return fmt.Errorf("reading stream URL list for %s: %w", username, err)
	}

	segmentURL := getSecondLastURL(string(streamBody))
	if segmentURL == "" {
		return fmt.Errorf("no segment URL found for %s", username)
	}

	headReq, err := http.NewRequestWithContext(ctx, http.MethodHead, segmentURL, nil)
	if err != nil {
		return fmt.Errorf("creating HEAD request: %w", err)
	}
	headReq.Header.Set("User-Agent", constants.DefaultUserAgent)

	headResp, err := httpClient.Do(headReq)
	if err != nil {
		return fmt.Errorf("HEAD request for %s: %w", username, err)
	}
	headResp.Body.Close()

	if headResp.StatusCode != http.StatusOK {
		return fmt.Errorf("HEAD request for %s returned status %d", username, headResp.StatusCode)
	}

	c.Log.Debug("Simulated stream watching", "streamer", username)

	encodedPayload, err := encodePayload(payload)
	if err != nil {
		return fmt.Errorf("encoding payload for %s: %w", username, err)
	}

	spadeReq, err := http.NewRequestWithContext(ctx, http.MethodPost, spadeURL,
		strings.NewReader(encodedPayload))
	if err != nil {
		return fmt.Errorf("creating spade request: %w", err)
	}
	spadeReq.Header.Set("User-Agent", constants.DefaultUserAgent)

	spadeResp, err := httpClient.Do(spadeReq)
	if err != nil {
		return fmt.Errorf("sending spade event for %s: %w", username, err)
	}
	spadeResp.Body.Close()

	if spadeResp.StatusCode == http.StatusNoContent || spadeResp.StatusCode == http.StatusOK {
		streamer.Mu.Lock()
		streamer.Stream.UpdateMinuteWatched()
		streamer.Mu.Unlock()

		c.Log.Debug("Sent minute watched event",
			"streamer", username,
			"status", spadeResp.StatusCode)

		c.logDropProgress(ctx, streamer)
		return nil
	}

	return fmt.Errorf("spade event for %s returned status %d", username, spadeResp.StatusCode)
}

func (c *Client) logDropProgress(ctx context.Context, streamer *model.Streamer) {
	streamer.Mu.Lock()
	defer streamer.Mu.Unlock()

	for _, campaign := range streamer.Stream.Campaigns {
		for _, drop := range campaign.Drops {
			if drop.HasPreconditionsMet != nil && !*drop.HasPreconditionsMet {
				continue
			}
			if drop.IsPrintable {
				streamer.UpdateHistory(string(model.EventDropStatus), 0, 1)
				c.Log.Event(ctx, model.EventDropStatus, "Drop progress",
					"streamer", streamer.Username,
					"stream", streamer.Stream.String(),
					"campaign", campaign.String(),
					"drop", drop.String(),
					"progress", drop.ProgressBar())
			}
		}
	}
}

// encodePayload encodes the minute-watched payload as base64 JSON,
func encodePayload(payload map[string]any) (string, error) {
	wrapped := []map[string]any{payload}
	jsonData, err := json.Marshal(wrapped)
	if err != nil {
		return "", fmt.Errorf("marshaling payload: %w", err)
	}
	return base64.StdEncoding.EncodeToString(jsonData), nil
}

func getLastURL(manifest string) string {
	lines := strings.Split(strings.TrimSpace(manifest), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if strings.HasPrefix(line, "http://") || strings.HasPrefix(line, "https://") {
			return line
		}
	}
	return ""
}

// getSecondLastURL returns the second-to-last URL from an m3u8 segment playlist.
func getSecondLastURL(playlist string) string {
	lines := strings.Split(strings.TrimSpace(playlist), "\n")
	count := 0
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if strings.HasPrefix(line, "http://") || strings.HasPrefix(line, "https://") {
			count++
			if count == 1 {
				continue
			}
			return line
		}
	}
	if count == 1 {
		return getLastURL(playlist)
	}
	return ""
}

// WatchOptions configures which streamers receive minute-watched events.
type WatchOptions struct {
	// Priorities is the ordered list of selection strategies.
	Priorities []model.Priority
	// MaxWatch is the width of the watch set once no streak is pending.
	MaxWatch int
	// StreakWatch narrows the set while a streak is pending. Zero disables
	// the narrowing and MaxWatch applies throughout.
	StreakWatch int
	// StreakMinutes is how long a channel may hold a streak slot.
	StreakMinutes float64
	// Preferred lists channel logins for the PREFERRED priority, in order.
	Preferred []string
}

// WatchSet is the outcome of a selection pass.
type WatchSet struct {
	// Streamers are the channels to send minute-watched events to.
	Streamers []*model.Streamer
	// StreakHarvest reports whether the set was narrowed to chase pending
	// watch streaks rather than run at full width.
	StreakHarvest bool
	// Width is the cap that was applied.
	Width int
}

// SelectStreamersToWatch selects up to maxWatch streamers using the default
// streak window. Kept for callers that do not configure the adaptive width.
func SelectStreamersToWatch(streamers []*model.Streamer, priorities []model.Priority, maxWatch int) []*model.Streamer {
	return SelectWatchSet(streamers, WatchOptions{
		Priorities:    priorities,
		MaxWatch:      maxWatch,
		StreakMinutes: constants.WatchStreakMinutes,
	}).Streamers
}

// SelectWatchSet selects the streamers to send minute-watched events for.
//
// Twitch only credits roughly two concurrent streams. Sending to every live
// channel therefore lets Twitch pick which two count, and no watch streak
// lands reliably. So while any channel still has a streak pending, the set
// narrows to StreakWatch and the STREAK priority fills it; each channel holds
// its slot until the streak arrives or StreakMinutes elapse, then gives way to
// the next. Once nothing is pending the set widens back to MaxWatch.
func SelectWatchSet(streamers []*model.Streamer, opts WatchOptions) WatchSet {
	now := time.Now()
	onlineIndices := collectOnlineIndices(streamers, now)
	if len(onlineIndices) == 0 {
		return WatchSet{}
	}

	if opts.StreakMinutes <= 0 {
		opts.StreakMinutes = constants.WatchStreakMinutes
	}

	maxWatch := opts.MaxWatch
	streakHarvest := opts.StreakWatch > 0 && anyStreakPending(streamers, onlineIndices, opts.StreakMinutes)
	if streakHarvest {
		maxWatch = opts.StreakWatch
	}
	if maxWatch <= 0 {
		maxWatch = len(streamers)
	}

	watching := make(map[int]struct{})
	selectedIndices := make([]int, 0, maxWatch)

	add := func(idx int) bool {
		if len(selectedIndices) >= maxWatch {
			return false
		}
		if _, ok := watching[idx]; ok {
			return false
		}
		watching[idx] = struct{}{}
		selectedIndices = append(selectedIndices, idx)
		return true
	}

	for _, priority := range opts.Priorities {
		if len(watching) >= maxWatch {
			break
		}
		applyPriority(priority, streamers, onlineIndices, watching, maxWatch, add, opts)
	}

	result := make([]*model.Streamer, 0, len(selectedIndices))
	for _, idx := range selectedIndices {
		result = append(result, streamers[idx])
	}

	return WatchSet{
		Streamers:     result,
		StreakHarvest: streakHarvest,
		Width:         maxWatch,
	}
}

// isStreakPending reports whether a streamer still stands to earn a watch
// streak this session. A channel that has already been given its slot time
// without the streak arriving is no longer pending, so the narrowed watch set
// cannot get stuck on it.
func isStreakPending(s *model.Streamer, streakMinutes float64) bool {
	s.Mu.RLock()
	defer s.Mu.RUnlock()

	return s.Settings != nil && s.Settings.WatchStreak &&
		s.Stream.IsWatchStreakMissing &&
		s.Stream.MinuteWatched < streakMinutes
}

func anyStreakPending(streamers []*model.Streamer, onlineIndices []int, streakMinutes float64) bool {
	for _, idx := range onlineIndices {
		if isStreakPending(streamers[idx], streakMinutes) {
			return true
		}
	}
	return false
}

func collectOnlineIndices(streamers []*model.Streamer, now time.Time) []int {
	var out []int
	for i, s := range streamers {
		s.Mu.RLock()
		isOnline := s.IsOnline
		onlineAt := s.OnlineAt
		dropsOnly := s.Settings != nil && s.Settings.DropsOnly
		// Gate on CampaignIDs (populated at discovery and by updateStream), not
		// Campaigns, which is only refreshed by the periodic campaign sync.
		hasCampaignIDs := len(s.Stream.CampaignIDs) > 0
		frozen := s.Stream.IsMinuteWatchStalled(constants.FreezeDetectionThreshold)
		cooldownUntil := s.Stream.StalledCooldownUntil
		s.Mu.RUnlock()
		if frozen {
			if cooldownUntil.IsZero() {
				s.Mu.Lock()
				s.Stream.StalledCooldownUntil = now.Add(constants.StalledCooldownDuration)
				s.Mu.Unlock()
			}
			continue
		}
		if !cooldownUntil.IsZero() && now.Before(cooldownUntil) {
			continue
		}
		if isOnline && (onlineAt.IsZero() || now.Sub(onlineAt) > 30*time.Second) {
			if dropsOnly && !hasCampaignIDs {
				continue
			}
			out = append(out, i)
		}
	}
	return out
}

func applyPriority(
	priority model.Priority,
	streamers []*model.Streamer,
	onlineIndices []int,
	watching map[int]struct{},
	maxWatch int,
	add func(int) bool,
	opts WatchOptions,
) {
	remaining := maxWatch - len(watching)
	switch priority {
	case model.PriorityOrder:
		applyPriorityOrder(onlineIndices, remaining, add)
	case model.PriorityPreferred:
		applyPriorityPreferred(streamers, onlineIndices, watching, remaining, add, opts.Preferred)
	case model.PriorityStreak:
		applyPriorityStreak(streamers, onlineIndices, watching, remaining, add, opts.StreakMinutes)
	case model.PriorityDrops:
		applyPriorityDrops(streamers, onlineIndices, watching, remaining, add)
	case model.PrioritySubscribed:
		applyPrioritySubscribed(streamers, onlineIndices, watching, remaining, add)
	case model.PriorityPointsAscending, model.PriorityPointsDescending:
		applyPriorityPoints(priority, streamers, onlineIndices, watching, remaining, add)
	case model.PriorityEndingSoonest:
		applyPriorityEndingSoonest(streamers, onlineIndices, watching, remaining, add)
	case model.PriorityLowAvailabilityFirst:
		applyPriorityLowAvailability(streamers, onlineIndices, watching, remaining, add)
	}
}

func applyPriorityOrder(onlineIndices []int, remaining int, add func(int) bool) {
	for _, idx := range onlineIndices {
		if add(idx) {
			remaining--
			if remaining <= 0 {
				break
			}
		}
	}
}

// applyPriorityStreak fills slots with channels still owed a watch streak,
// longest-served last: a channel keeps its slot until the streak arrives or
// streakMinutes of watch time have gone by, then the next one takes over.
func applyPriorityStreak(streamers []*model.Streamer, onlineIndices []int, watching map[int]struct{}, remaining int, add func(int) bool, streakMinutes float64) {
	if streakMinutes <= 0 {
		streakMinutes = constants.WatchStreakMinutes
	}

	pending := make([]int, 0, len(onlineIndices))
	for _, idx := range onlineIndices {
		if _, ok := watching[idx]; ok {
			continue
		}
		if isStreakPending(streamers[idx], streakMinutes) {
			pending = append(pending, idx)
		}
	}

	// Most watch time first, so a channel part-way through its slot finishes
	// before a fresh one starts. Without this the set would churn every tick
	// and nobody would accumulate the minutes a streak needs.
	sort.SliceStable(pending, func(i, j int) bool {
		return minuteWatchedOf(streamers[pending[i]]) > minuteWatchedOf(streamers[pending[j]])
	})

	for _, idx := range pending {
		if add(idx) {
			remaining--
			if remaining <= 0 {
				break
			}
		}
	}
}

func minuteWatchedOf(s *model.Streamer) float64 {
	s.Mu.RLock()
	defer s.Mu.RUnlock()
	return s.Stream.MinuteWatched
}

// applyPriorityPreferred fills slots from the configured preferred_streamers
// list, in the order it is written.
func applyPriorityPreferred(streamers []*model.Streamer, onlineIndices []int, watching map[int]struct{}, remaining int, add func(int) bool, preferred []string) {
	if len(preferred) == 0 {
		return
	}

	online := make(map[string]int, len(onlineIndices))
	for _, idx := range onlineIndices {
		if _, ok := watching[idx]; ok {
			continue
		}
		streamers[idx].Mu.RLock()
		username := strings.ToLower(streamers[idx].Username)
		streamers[idx].Mu.RUnlock()
		online[username] = idx
	}

	for _, name := range preferred {
		idx, ok := online[strings.ToLower(strings.TrimSpace(name))]
		if !ok {
			continue
		}
		if add(idx) {
			remaining--
			if remaining <= 0 {
				break
			}
		}
	}
}

func applyPriorityDrops(streamers []*model.Streamer, onlineIndices []int, watching map[int]struct{}, remaining int, add func(int) bool) {
	for _, idx := range onlineIndices {
		if _, ok := watching[idx]; ok {
			continue
		}
		s := streamers[idx]
		s.Mu.RLock()
		dropsCondition := s.DropsCondition()
		s.Mu.RUnlock()
		if dropsCondition && add(idx) {
			remaining--
			if remaining <= 0 {
				break
			}
		}
	}
}

func applyPrioritySubscribed(streamers []*model.Streamer, onlineIndices []int, watching map[int]struct{}, remaining int, add func(int) bool) {
	type indexMultiplier struct {
		index      int
		multiplier float64
	}
	var candidates []indexMultiplier
	for _, idx := range onlineIndices {
		if _, ok := watching[idx]; ok {
			continue
		}
		s := streamers[idx]
		s.Mu.RLock()
		hasMultiplier := s.HasPointsMultiplier()
		total := s.TotalPointsMultiplier()
		s.Mu.RUnlock()
		if hasMultiplier {
			candidates = append(candidates, indexMultiplier{idx, total})
		}
	}
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].multiplier > candidates[j].multiplier
	})
	for _, c := range candidates {
		if add(c.index) {
			remaining--
		}
		if remaining <= 0 {
			break
		}
	}
}

func applyPriorityPoints(priority model.Priority, streamers []*model.Streamer, onlineIndices []int, watching map[int]struct{}, remaining int, add func(int) bool) {
	type indexPoints struct {
		index  int
		points int
	}
	items := make([]indexPoints, 0, len(onlineIndices))
	for _, idx := range onlineIndices {
		if _, ok := watching[idx]; ok {
			continue
		}
		s := streamers[idx]
		s.Mu.RLock()
		points := s.ChannelPoints
		s.Mu.RUnlock()
		items = append(items, indexPoints{idx, points})
	}
	if priority == model.PriorityPointsAscending {
		sort.Slice(items, func(i, j int) bool { return items[i].points < items[j].points })
	} else {
		sort.Slice(items, func(i, j int) bool { return items[i].points > items[j].points })
	}
	for _, item := range items {
		if add(item.index) {
			remaining--
			if remaining <= 0 {
				break
			}
		}
	}
}

func applyPriorityEndingSoonest(streamers []*model.Streamer, onlineIndices []int, watching map[int]struct{}, remaining int, add func(int) bool) {
	type indexEnd struct {
		index int
		endAt time.Time
	}
	var candidates []indexEnd
	for _, idx := range onlineIndices {
		if _, ok := watching[idx]; ok {
			continue
		}
		s := streamers[idx]
		s.Mu.RLock()
		campaigns := s.Stream.Campaigns
		s.Mu.RUnlock()
		if len(campaigns) == 0 {
			continue
		}
		var earliest time.Time
		for _, c := range campaigns {
			if !c.EndAt.IsZero() && (earliest.IsZero() || c.EndAt.Before(earliest)) {
				earliest = c.EndAt
			}
		}
		if !earliest.IsZero() {
			candidates = append(candidates, indexEnd{idx, earliest})
		}
	}
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].endAt.Before(candidates[j].endAt)
	})
	for _, c := range candidates {
		if add(c.index) {
			remaining--
			if remaining <= 0 {
				break
			}
		}
	}
}

func applyPriorityLowAvailability(streamers []*model.Streamer, onlineIndices []int, watching map[int]struct{}, remaining int, add func(int) bool) {
	type indexDrops struct {
		index     int
		dropCount int
	}
	var candidates []indexDrops
	for _, idx := range onlineIndices {
		if _, ok := watching[idx]; ok {
			continue
		}
		s := streamers[idx]
		s.Mu.RLock()
		campaigns := s.Stream.Campaigns
		s.Mu.RUnlock()
		if len(campaigns) == 0 {
			continue
		}
		totalDrops := 0
		for _, c := range campaigns {
			totalDrops += len(c.Drops)
		}
		candidates = append(candidates, indexDrops{idx, totalDrops})
	}
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].dropCount < candidates[j].dropCount
	})
	for _, c := range candidates {
		if add(c.index) {
			remaining--
			if remaining <= 0 {
				break
			}
		}
	}
}
