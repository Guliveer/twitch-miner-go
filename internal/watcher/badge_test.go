package watcher

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/Guliveer/twitch-miner-go/internal/config"
	"github.com/Guliveer/twitch-miner-go/internal/gql"
	"github.com/Guliveer/twitch-miner-go/internal/logger"
	"github.com/Guliveer/twitch-miner-go/internal/model"
)

type stubBadgeGQL struct {
	owned                 map[string]struct{}
	streams               map[string][]gql.TopStream
	taggedStreams         map[string][]gql.TopStream
	streamInfo            map[string]*gql.StreamInfoResponse
	channelIDs            map[string]string
	streamErrors          map[string]error
	idErrors              map[string]error
	streamHits            map[string]int
	idHits                map[string]int
	categoryHits          []string
	categoryLimit         int
	categoryDropsOnly     bool
	categoryDropsHistory  []bool
	availableCampaignsErr error
	availableCampaigns    map[string][]string
	availableErrors       map[string]error
	availableHits         map[string]int
	raw                   json.RawMessage
}

func (s *stubBadgeGQL) GetAvailableBadgeNames(context.Context) (map[string]struct{}, error) {
	return s.owned, nil
}

func (s *stubBadgeGQL) GetDropsInventory(context.Context) (json.RawMessage, error) {
	if s.raw == nil {
		return json.RawMessage(`{}`), nil
	}
	return s.raw, nil
}

func (s *stubBadgeGQL) GetTopStreamsByCategory(_ context.Context, slug string, limit int, dropsOnly bool) ([]gql.TopStream, error) {
	s.categoryHits = append(s.categoryHits, slug)
	s.categoryLimit = limit
	s.categoryDropsOnly = dropsOnly
	s.categoryDropsHistory = append(s.categoryDropsHistory, dropsOnly)
	if dropsOnly && s.taggedStreams != nil {
		return s.taggedStreams[slug], nil
	}
	return s.streams[slug], nil
}

func (s *stubBadgeGQL) GetStreamInfo(_ context.Context, login string) (*gql.StreamInfoResponse, error) {
	if s.streamHits == nil {
		s.streamHits = make(map[string]int)
	}
	s.streamHits[login]++
	if err := s.streamErrors[login]; err != nil {
		return nil, err
	}
	return s.streamInfo[login], nil
}

func (s *stubBadgeGQL) GetUserID(_ context.Context, login string) (string, error) {
	if s.idHits == nil {
		s.idHits = make(map[string]int)
	}
	s.idHits[login]++
	if err := s.idErrors[login]; err != nil {
		return "", err
	}
	if id := s.channelIDs[login]; id != "" {
		return id, nil
	}
	return "id-" + login, nil
}

func (s *stubBadgeGQL) GetAvailableCampaigns(_ context.Context, channelID string) ([]string, error) {
	if s.availableHits == nil {
		s.availableHits = make(map[string]int)
	}
	s.availableHits[channelID]++
	if err := s.availableErrors[channelID]; err != nil {
		return nil, err
	}
	if s.availableCampaignsErr != nil {
		return nil, s.availableCampaignsErr
	}
	if ids, ok := s.availableCampaigns[channelID]; ok {
		return ids, nil
	}
	return []string{"one", "two", "three", "drop-campaign"}, nil
}

func testBadgeWatcher(t *testing.T, client *stubBadgeGQL, limit int, campaigns ...badgeCampaign) *BadgeWatcher {
	t.Helper()
	log, err := logger.Setup(logger.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	bw := NewBadgeWatcher(config.BadgeWatcherConfig{StreamerLimit: limit}, client, nil, log, nil, &model.StreamerSettings{})
	bw.campaigns = campaigns
	bw.catalogLoadedAt = time.Now()
	return bw
}

func activeBadgeCampaign(id, name, slug, badge string) badgeCampaign {
	return badgeCampaign{ID: id, Name: name, GameName: name, GameSlug: slug, EndsAt: time.Now().Add(time.Hour), AllChannels: true, BadgeNames: []string{badge}}
}

func TestBadgeWatcherEvaluateAddsAndLimitsStreams(t *testing.T) {
	campaigns := []badgeCampaign{
		activeBadgeCampaign("one", "Game One", "game-one", "Badge One"),
		activeBadgeCampaign("two", "Game Two", "game-two", "Badge Two"),
		activeBadgeCampaign("three", "Game Three", "game-three", "Badge Three"),
	}
	client := &stubBadgeGQL{streams: map[string][]gql.TopStream{
		"game-one":   {{Username: "one", ChannelID: "1", GameName: "Game One"}},
		"game-two":   {{Username: "two", ChannelID: "2", GameName: "Game Two"}},
		"game-three": {{Username: "three", ChannelID: "3", GameName: "Game Three"}},
	}}
	bw := testBadgeWatcher(t, client, 99, campaigns...)
	var added []*model.Streamer
	bw.evaluate(context.Background(), func(_ context.Context, streamer *model.Streamer) {
		added = append(added, streamer)
	}, func(string, string) {}, func() []*model.Streamer { return added })

	if len(added) != 2 {
		t.Fatalf("added %d streamers, want runtime cap of 2", len(added))
	}
	for _, streamer := range added {
		if !streamer.IsBadgeWatched || !streamer.Settings.ClaimDrops || !streamer.Settings.DropsOnly || streamer.Settings.Chat != model.ChatNever {
			t.Fatalf("unexpected badge streamer settings: %#v", streamer)
		}
	}
}

func TestBadgeWatcherAllChannelsSearchDoesNotRequireDropsTag(t *testing.T) {
	campaign := activeBadgeCampaign("one", "Game One", "game-one", "Badge One")
	client := &stubBadgeGQL{streams: map[string][]gql.TopStream{"game-one": {{Username: "candidate"}}}}
	bw := testBadgeWatcher(t, client, 1, campaign)
	streams, err := bw.findCampaignStreams(context.Background(), campaign)
	if err != nil {
		t.Fatal(err)
	}
	if len(streams) != 1 || client.categoryLimit != badgeCategoryStreamLimit || client.categoryDropsOnly {
		t.Fatalf("unexpected category query: streams=%#v limit=%d dropsOnly=%v", streams, client.categoryLimit, client.categoryDropsOnly)
	}
}

func TestBadgeWatcherAllChannelsFallsBackAndSelectsEligibleCampaign(t *testing.T) {
	campaign := activeBadgeCampaign("twitchdrops-app-17cdc638f243b467", "Game", "game", "Badge")
	client := &stubBadgeGQL{
		taggedStreams: map[string][]gql.TopStream{"game": {{Username: "largest", ChannelID: "1"}}},
		streams: map[string][]gql.TopStream{"game": {
			{Username: "largest", ChannelID: "1"},
			{Username: "eligible", ChannelID: "2"},
		}},
		availableCampaigns: map[string][]string{"1": {}, "2": {"174fd699-3230-4076-b542-c15a8dcb4ab1"}},
	}
	bw := testBadgeWatcher(t, client, 1, campaign)
	candidate, err := bw.findCampaignCandidate(context.Background(), campaign, map[string]bool{}, make(map[string]badgeCampaignLookup))
	if err != nil {
		t.Fatal(err)
	}
	if candidate == nil || candidate.stream.Username != "eligible" || len(candidate.campaignIDs) != 1 || candidate.campaignIDs[0] != "174fd699-3230-4076-b542-c15a8dcb4ab1" {
		t.Fatalf("unexpected candidate: %#v", candidate)
	}
	if client.availableHits["1"] != 1 || client.availableHits["2"] != 1 {
		t.Fatalf("campaign lookups were not cached: %v", client.availableHits)
	}
	if len(client.categoryDropsHistory) != 2 || !client.categoryDropsHistory[0] || client.categoryDropsHistory[1] {
		t.Fatalf("unexpected category query order: %v", client.categoryDropsHistory)
	}
}

func TestBadgeWatcherAllChannelsStopsAtFirstEligibleTaggedStream(t *testing.T) {
	campaign := activeBadgeCampaign("twitchdrops-app-17cdc638f243b467", "Game", "game", "Badge")
	client := &stubBadgeGQL{
		taggedStreams:      map[string][]gql.TopStream{"game": {{Username: "eligible", ChannelID: "1"}, {Username: "unused", ChannelID: "2"}}},
		streams:            map[string][]gql.TopStream{"game": {{Username: "fallback", ChannelID: "3"}}},
		availableCampaigns: map[string][]string{"1": {"174fd699-3230-4076-b542-c15a8dcb4ab1"}},
	}
	bw := testBadgeWatcher(t, client, 1, campaign)
	candidate, err := bw.findCampaignCandidate(context.Background(), campaign, map[string]bool{}, make(map[string]badgeCampaignLookup))
	if err != nil {
		t.Fatal(err)
	}
	if candidate == nil || candidate.stream.Username != "eligible" || len(client.categoryDropsHistory) != 1 || !client.categoryDropsHistory[0] {
		t.Fatalf("tagged candidate was not selected directly: candidate=%#v queries=%v", candidate, client.categoryDropsHistory)
	}
	if client.availableHits["2"] != 0 || client.availableHits["3"] != 0 {
		t.Fatalf("search did not stop after the first eligible stream: %v", client.availableHits)
	}
}

func TestBadgeWatcherAllChannelsSkipsLookupErrorsAndIneligibleStreams(t *testing.T) {
	campaign := activeBadgeCampaign("twitchdrops-app-17cdc638f243b467", "Game", "game", "Badge")
	client := &stubBadgeGQL{
		taggedStreams: map[string][]gql.TopStream{"game": {
			{Username: "error", ChannelID: "1"},
			{Username: "wrong", ChannelID: "2"},
		}},
		streams:            map[string][]gql.TopStream{"game": {{Username: "wrong", ChannelID: "2"}}},
		availableCampaigns: map[string][]string{"2": {}},
		availableErrors:    map[string]error{"1": errors.New("campaign lookup failed")},
	}
	bw := testBadgeWatcher(t, client, 1, campaign)
	candidate, err := bw.findCampaignCandidate(context.Background(), campaign, map[string]bool{}, make(map[string]badgeCampaignLookup))
	if err != nil {
		t.Fatal(err)
	}
	if candidate != nil || client.availableHits["1"] != 1 || client.availableHits["2"] != 1 {
		t.Fatalf("unexpected ineligible result: candidate=%#v hits=%v", candidate, client.availableHits)
	}
}

func TestBadgeWatcherCampaignLookupCacheIsReusableWithinPoll(t *testing.T) {
	first := activeBadgeCampaign("twitchdrops-app-first", "Game", "game", "First Badge")
	second := activeBadgeCampaign("twitchdrops-app-second", "Game", "game", "Second Badge")
	client := &stubBadgeGQL{
		taggedStreams:      map[string][]gql.TopStream{"game": {{Username: "eligible", ChannelID: "1"}}},
		availableCampaigns: map[string][]string{"1": {"174fd699-3230-4076-b542-c15a8dcb4ab1"}},
	}
	bw := testBadgeWatcher(t, client, 2, first, second)
	cache := make(map[string]badgeCampaignLookup)
	if candidate, err := bw.findCampaignCandidate(context.Background(), first, map[string]bool{}, cache); err != nil || candidate == nil {
		t.Fatalf("first lookup failed: candidate=%#v err=%v", candidate, err)
	}
	if candidate, err := bw.findCampaignCandidate(context.Background(), second, map[string]bool{}, cache); err != nil || candidate == nil {
		t.Fatalf("second lookup failed: candidate=%#v err=%v", candidate, err)
	}
	if client.availableHits["1"] != 1 {
		t.Fatalf("campaign lookup cache was not reused: %v", client.availableHits)
	}
}

func TestBadgeStreamerUsesPreloadedCampaignIDs(t *testing.T) {
	campaign := activeBadgeCampaign("twitchdrops-app-17cdc638f243b467", "Game", "game", "Badge")
	client := &stubBadgeGQL{}
	bw := testBadgeWatcher(t, client, 1, campaign)
	streamer := bw.badgeStreamer(context.Background(), &gql.TopStream{Username: "eligible", ChannelID: "1"}, campaign, []string{"174fd699-3230-4076-b542-c15a8dcb4ab1"})
	if client.availableHits["1"] != 0 || len(streamer.Stream.CampaignIDs) != 1 || streamer.Stream.CampaignIDs[0] != "174fd699-3230-4076-b542-c15a8dcb4ab1" {
		t.Fatalf("preloaded campaign IDs were not reused: hits=%v streamer=%#v", client.availableHits, streamer)
	}
}

func TestBadgeStreamerValidRequiresActiveDropsForAllChannels(t *testing.T) {
	campaign := activeBadgeCampaign("twitchdrops-app-17cdc638f243b467", "Game", "game", "Badge")
	streamer := model.NewStreamer("candidate")
	streamer.IsOnline = true
	streamer.Stream.Game = &model.GameInfo{Slug: "game"}
	streamer.Stream.CampaignIDs = nil
	if badgeStreamerValid(streamer, campaign) {
		t.Fatal("streamer without an active Twitch campaign was considered valid")
	}
	streamer.Stream.CampaignIDs = []string{"174fd699-3230-4076-b542-c15a8dcb4ab1"}
	if !badgeStreamerValid(streamer, campaign) {
		t.Fatal("streamer with an active Twitch campaign was considered invalid")
	}
}

func TestBadgeWatcherEvaluateMarksExistingStreamer(t *testing.T) {
	campaign := activeBadgeCampaign("twitchdrops-app-17cdc638f243b467", "Game One", "game-one", "Badge One")
	client := &stubBadgeGQL{streams: map[string][]gql.TopStream{
		"game-one": {{Username: "existing", ChannelID: "1", GameName: "Game One"}},
	}}
	existing := model.NewStreamer("existing")
	existing.IsOnline = true
	existing.Stream.Game = &model.GameInfo{Slug: "game-one", Name: "Game One"}
	bw := testBadgeWatcher(t, client, 1, campaign)
	added := false
	bw.evaluate(context.Background(), func(context.Context, *model.Streamer) { added = true }, func(string, string) {}, func() []*model.Streamer { return []*model.Streamer{existing} })

	if added || !existing.IsBadgeWatched || existing.BadgeCampaign != campaign.Name || len(existing.Stream.CampaignIDs) == 0 {
		t.Fatalf("existing streamer was not marked correctly: added=%v streamer=%#v", added, existing)
	}
}

func TestBadgeWatcherEligibleAllChannelsStreamerDoesNotFlap(t *testing.T) {
	campaign := activeBadgeCampaign("twitchdrops-app-17cdc638f243b467", "Game", "game", "Badge")
	client := &stubBadgeGQL{
		taggedStreams:      map[string][]gql.TopStream{"game": {{Username: "eligible", ChannelID: "1", GameSlug: "game"}}},
		availableCampaigns: map[string][]string{"1": {"174fd699-3230-4076-b542-c15a8dcb4ab1"}},
	}
	bw := testBadgeWatcher(t, client, 1, campaign)
	var streamers []*model.Streamer
	get := func() []*model.Streamer { return streamers }
	bw.evaluate(context.Background(), func(_ context.Context, streamer *model.Streamer) { streamers = append(streamers, streamer) }, func(string, string) {}, get)
	if len(streamers) != 1 {
		t.Fatalf("eligible streamer was not added: %#v", streamers)
	}
	bw.evaluate(context.Background(), func(context.Context, *model.Streamer) {}, func(_, reason string) {
		t.Fatalf("eligible streamer flapped on the next poll: %s", reason)
	}, get)
}

func TestBadgeWatcherFindsRestrictedSpecialEventChannelOutsideCampaignCategory(t *testing.T) {
	campaign := badgeCampaign{
		ID:         "mouseathon",
		Name:       "Ironmouse Subathon 2026",
		GameName:   "Special Events",
		GameSlug:   "special-events",
		EndsAt:     time.Now().Add(time.Hour),
		Channels:   []string{"ironmouse"},
		BadgeNames: []string{"Mouseathon"},
	}
	client := &stubBadgeGQL{
		streams: map[string][]gql.TopStream{},
		streamInfo: map[string]*gql.StreamInfoResponse{
			"ironmouse": {Game: &model.GameInfo{ID: "509658", Slug: "just-chatting", Name: "Just Chatting"}, ViewersCount: 12000},
		},
		channelIDs: map[string]string{"ironmouse": "123456"},
	}
	bw := testBadgeWatcher(t, client, 1, campaign)
	var added []*model.Streamer
	get := func() []*model.Streamer { return added }
	bw.evaluate(context.Background(), func(_ context.Context, streamer *model.Streamer) {
		added = append(added, streamer)
	}, func(string, string) {}, get)

	if len(added) != 1 || added[0].Username != "ironmouse" {
		t.Fatalf("restricted special-event streamer not added: %#v", added)
	}
	if added[0].Settings.DropsOnly {
		t.Fatal("restricted badge streamer should trust its allow-list when campaign IDs are unavailable")
	}
	if got := added[0].Stream.Game.Slug; got != "just-chatting" {
		t.Fatalf("stream category=%q, want actual category just-chatting", got)
	}
	if len(client.categoryHits) != 0 {
		t.Fatalf("restricted campaign unexpectedly searched categories: %v", client.categoryHits)
	}

	bw.evaluate(context.Background(), func(context.Context, *model.Streamer) {}, func(_, reason string) {
		t.Fatalf("restricted special-event streamer retired on next poll: %s", reason)
	}, get)
}

func TestBadgeWatcherRestrictedCampaignRequiresMatchingCategoryOutsideSpecialEvents(t *testing.T) {
	campaign := badgeCampaign{
		ID:         "restricted",
		Name:       "Restricted Game Campaign",
		GameName:   "Expected Game",
		GameSlug:   "expected-game",
		EndsAt:     time.Now().Add(time.Hour),
		Channels:   []string{"allowed"},
		BadgeNames: []string{"Expected Badge"},
	}
	client := &stubBadgeGQL{
		streams: map[string][]gql.TopStream{},
		streamInfo: map[string]*gql.StreamInfoResponse{
			"allowed": {Game: &model.GameInfo{Slug: "other-game", Name: "Other Game"}},
		},
	}
	bw := testBadgeWatcher(t, client, 1, campaign)
	var added []*model.Streamer
	bw.evaluate(context.Background(), func(_ context.Context, streamer *model.Streamer) {
		added = append(added, streamer)
	}, func(string, string) {}, func() []*model.Streamer { return added })

	if len(added) != 0 {
		t.Fatalf("restricted streamer in wrong category was added: %#v", added)
	}
}

func TestBadgeWatcherRestrictedChannelsHandleOfflineErrorsBlacklistAndIDCache(t *testing.T) {
	campaign := badgeCampaign{
		ID:         "restricted",
		Name:       "Restricted Campaign",
		GameSlug:   crossCategoryBadgeCampaignSlug,
		Channels:   []string{"MixedCase", "offline", "stream-error", "id-error", "blocked"},
		BadgeNames: []string{"Badge"},
	}
	client := &stubBadgeGQL{
		streamInfo: map[string]*gql.StreamInfoResponse{
			"mixedcase": {Game: &model.GameInfo{Slug: "just-chatting"}},
			"offline":   nil,
			"id-error":  {Game: &model.GameInfo{Slug: "just-chatting"}},
		},
		streamErrors: map[string]error{"stream-error": errors.New("stream lookup failed")},
		idErrors:     map[string]error{"id-error": errors.New("id lookup failed")},
		channelIDs:   map[string]string{"mixedcase": "123"},
	}
	bw := testBadgeWatcher(t, client, 1, campaign)
	bw.blacklist["blocked"] = true

	first, err := bw.findCampaignStreams(context.Background(), campaign)
	if err != nil {
		t.Fatal(err)
	}
	second, err := bw.findCampaignStreams(context.Background(), campaign)
	if err != nil {
		t.Fatal(err)
	}
	if len(first) != 1 || len(second) != 1 || first[0].DisplayName != "MixedCase" {
		t.Fatalf("unexpected restricted streams: first=%#v second=%#v", first, second)
	}
	if client.idHits["mixedcase"] != 1 {
		t.Fatalf("channel ID fetched %d times, want once", client.idHits["mixedcase"])
	}
	if client.idHits["offline"] != 0 || client.streamHits["blocked"] != 0 {
		t.Fatalf("offline or blacklisted channel caused extra lookup: idHits=%v streamHits=%v", client.idHits, client.streamHits)
	}
}

func TestBadgeWatcherSpecialEventWithoutGameDoesNotInventCategory(t *testing.T) {
	campaign := badgeCampaign{ID: "special", Name: "Special", GameSlug: crossCategoryBadgeCampaignSlug, EndsAt: time.Now().Add(time.Hour), Channels: []string{"ChannelCase"}, BadgeNames: []string{"Badge"}}
	client := &stubBadgeGQL{streamInfo: map[string]*gql.StreamInfoResponse{"channelcase": {ViewersCount: 1}}, channelIDs: map[string]string{"channelcase": "123"}}
	bw := testBadgeWatcher(t, client, 1, campaign)
	var added []*model.Streamer
	bw.evaluate(context.Background(), func(_ context.Context, streamer *model.Streamer) { added = append(added, streamer) }, func(string, string) {}, func() []*model.Streamer { return added })
	if len(added) != 1 || added[0].DisplayName != "ChannelCase" || added[0].Stream.Game != nil {
		t.Fatalf("unexpected special-event streamer: %#v", added)
	}
}

func TestBadgeWatcherPickCandidateHonorsReservationAllowlistAndBlacklist(t *testing.T) {
	bw := &BadgeWatcher{blacklist: map[string]bool{"blocked": true}}
	streams := []gql.TopStream{{Username: "blocked"}, {Username: "reserved"}, {Username: "not-allowed"}, {Username: "chosen"}}
	got := bw.pickCandidate(streams, map[string]bool{"reserved": true}, map[string]bool{"chosen": true}, false)
	if got == nil || got.Username != "chosen" {
		t.Fatalf("picked %#v, want chosen", got)
	}
}

func TestBadgeWatcherRetiresOwnedCampaignAndCleansUp(t *testing.T) {
	campaign := activeBadgeCampaign("one", "Game One", "game-one", "Badge One")
	client := &stubBadgeGQL{owned: map[string]struct{}{"badge one": {}}, streams: map[string][]gql.TopStream{}}
	streamer := model.NewStreamer("added")
	streamer.IsOnline = true
	streamer.Stream.Game = &model.GameInfo{Slug: "game-one"}
	bw := testBadgeWatcher(t, client, 1, campaign)
	bw.tracked[campaign.ID] = badgeTracked{username: streamer.Username, added: true}
	var removals []string
	remove := func(_ string, reason string) { removals = append(removals, reason) }
	bw.evaluate(context.Background(), func(context.Context, *model.Streamer) {}, remove, func() []*model.Streamer { return []*model.Streamer{streamer} })
	if len(removals) != 1 || removals[0] != "badge_campaign_completed_or_expired" {
		t.Fatalf("unexpected retirement: %v", removals)
	}

	bw.tracked["cleanup"] = badgeTracked{username: streamer.Username, added: true}
	bw.cleanup(remove, func() []*model.Streamer { return []*model.Streamer{streamer} })
	if len(removals) != 2 || removals[1] != "badge_watcher_shutdown" || len(bw.tracked) != 0 {
		t.Fatalf("unexpected cleanup: removals=%v tracked=%v", removals, bw.tracked)
	}
}

func TestBadgeWatcherRetiresInvalidTrackedStreamer(t *testing.T) {
	for _, tc := range []struct {
		name     string
		online   bool
		category string
	}{
		{name: "offline", online: false, category: "game-one"},
		{name: "changed category", online: true, category: "other-game"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			campaign := activeBadgeCampaign("one", "Game One", "game-one", "Badge One")
			client := &stubBadgeGQL{streams: map[string][]gql.TopStream{}}
			streamer := model.NewStreamer("tracked")
			streamer.IsOnline = tc.online
			streamer.Stream.Game = &model.GameInfo{Slug: tc.category}
			bw := testBadgeWatcher(t, client, 1, campaign)
			bw.tracked[campaign.ID] = badgeTracked{username: streamer.Username, added: true}
			var reason string
			bw.evaluate(context.Background(), func(context.Context, *model.Streamer) {}, func(_ string, got string) { reason = got }, func() []*model.Streamer { return []*model.Streamer{streamer} })
			if reason != "badge_stream_no_longer_eligible" {
				t.Fatalf("removal reason=%q", reason)
			}
		})
	}
}
