package twitch

import (
	"testing"
	"time"

	"github.com/Guliveer/twitch-miner-go/internal/constants"
	"github.com/Guliveer/twitch-miner-go/internal/logger"
	"github.com/Guliveer/twitch-miner-go/internal/model"
)

func newDropCreditClient(t *testing.T) *Client {
	t.Helper()

	log, err := logger.Setup(logger.Config{Level: 100})
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	return &Client{Log: log}
}

// dropCreditStreamer builds an online streamer watching one campaign whose only
// drop has the given progress.
func dropCreditStreamer(username, campaignID, dropID string, watched int) *model.Streamer {
	s := makeOnlineStreamer(username)
	s.Settings.ClaimDrops = true

	start := time.Now().Add(-time.Hour)
	end := time.Now().Add(time.Hour)

	drop := model.NewDrop(dropID, "Reward "+dropID, nil, 60, start, end)
	drop.Update(true, watched, "", false)

	campaign := model.NewCampaign(campaignID, "Campaign "+campaignID, "ACTIVE", nil, start, end, nil)
	campaign.Drops = []*model.Drop{drop}
	s.Stream.Campaigns = []model.Campaign{*campaign}

	return s
}

// watching stamps a fresh minute-watched attempt and acceptance on the streamer,
// which is what continuous watching looks like by the time a sync runs.
func watching(s *model.Streamer, at time.Time) {
	s.Stream.LastMinuteAttemptedAt = at
	s.Stream.LastMinuteCreditedAt = at
}

func TestStalledDropProgressCoolsDownStreamer(t *testing.T) {
	t.Parallel()

	client := newDropCreditClient(t)
	streamer := dropCreditStreamer("stalled", "c1", "d1", 10)
	streamers := []*model.Streamer{streamer}

	base := time.Now()
	watching(streamer, base)
	client.detectStalledDropCredit(streamers, base)

	tenMinutes := base.Add(10 * time.Minute)
	watching(streamer, tenMinutes)
	client.detectStalledDropCredit(streamers, tenMinutes)
	if !streamer.Stream.StalledCooldownUntil.IsZero() {
		t.Fatal("one quiet poll at 10 minutes is below the threshold")
	}

	twentyMinutes := base.Add(20 * time.Minute)
	watching(streamer, twentyMinutes)
	client.detectStalledDropCredit(streamers, twentyMinutes)

	want := twentyMinutes.Add(constants.StalledCooldownDuration)
	if !streamer.Stream.StalledCooldownUntil.Equal(want) {
		t.Fatalf("expected cooldown until %s, got %s", want, streamer.Stream.StalledCooldownUntil)
	}
}

func TestAdvancingDropProgressKeepsStreamerWatching(t *testing.T) {
	t.Parallel()

	client := newDropCreditClient(t)
	streamer := dropCreditStreamer("advancing", "c1", "d1", 10)
	streamers := []*model.Streamer{streamer}

	base := time.Now()
	watching(streamer, base)
	client.detectStalledDropCredit(streamers, base)

	tenMinutes := base.Add(10 * time.Minute)
	watching(streamer, tenMinutes)
	streamer.Stream.Campaigns[0].Drops[0].CurrentMinutesWatched = 20
	client.detectStalledDropCredit(streamers, tenMinutes)

	twentyMinutes := base.Add(20 * time.Minute)
	watching(streamer, twentyMinutes)
	streamer.Stream.Campaigns[0].Drops[0].CurrentMinutesWatched = 30
	client.detectStalledDropCredit(streamers, twentyMinutes)

	if !streamer.Stream.StalledCooldownUntil.IsZero() {
		t.Fatal("a drop that keeps advancing proves credit is flowing")
	}
}

// Progress is account-wide, so a campaign shared by two watched streamers
// cannot tell which one failed to credit.
func TestSharedCampaignIsNotAttributed(t *testing.T) {
	t.Parallel()

	client := newDropCreditClient(t)
	first := dropCreditStreamer("first", "shared", "d1", 10)
	second := dropCreditStreamer("second", "shared", "d2", 10)
	streamers := []*model.Streamer{first, second}

	base := time.Now()
	for _, s := range streamers {
		watching(s, base)
	}
	client.detectStalledDropCredit(streamers, base)

	later := base.Add(25 * time.Minute)
	for _, s := range streamers {
		watching(s, later)
	}
	client.detectStalledDropCredit(streamers, later)

	for _, s := range streamers {
		if !s.Stream.StalledCooldownUntil.IsZero() {
			t.Fatalf("%s was cooled down for a campaign it shares", s.Username)
		}
	}
}

// A streamer that has the campaign attached but is not being watched earns no
// credit, so it cannot hide a frozen drop on the one channel that is watched.
func TestCampaignSharedWithIdleStreamerIsAttributed(t *testing.T) {
	t.Parallel()

	client := newDropCreditClient(t)
	watched := dropCreditStreamer("watched", "shared", "d1", 10)
	idle := dropCreditStreamer("idle", "shared", "d2", 10)
	streamers := []*model.Streamer{watched, idle}

	base := time.Now()
	watching(watched, base)
	client.detectStalledDropCredit(streamers, base)

	later := base.Add(25 * time.Minute)
	watching(watched, later)
	client.detectStalledDropCredit(streamers, later)

	if watched.Stream.StalledCooldownUntil.IsZero() {
		t.Fatal("a frozen drop on the only watched channel must cool it down")
	}
	if !idle.Stream.StalledCooldownUntil.IsZero() {
		t.Fatal("a channel that is not being watched must not be cooled down")
	}
}

func TestUnwatchedStreamerIsNotCooledDown(t *testing.T) {
	t.Parallel()

	client := newDropCreditClient(t)
	streamer := dropCreditStreamer("rotated-out", "c1", "d1", 10)
	streamers := []*model.Streamer{streamer}

	base := time.Now()
	watching(streamer, base)
	client.detectStalledDropCredit(streamers, base)

	// The channel left the watch rotation; its frozen drop is not evidence.
	later := base.Add(time.Hour)
	watching(streamer, base)
	client.detectStalledDropCredit(streamers, later)

	if !streamer.Stream.StalledCooldownUntil.IsZero() {
		t.Fatal("a channel that is not being watched must not be cooled down")
	}
}
