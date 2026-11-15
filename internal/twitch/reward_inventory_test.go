package twitch

import (
	"context"
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/Guliveer/twitch-miner-go/internal/constants"
	"github.com/Guliveer/twitch-miner-go/internal/model"
)

func TestUrgentRewardUsesInventoryWhenDashboardIsEmpty(t *testing.T) {
	t.Parallel()

	now := time.Now()
	start := now.Add(-time.Hour).Format(time.RFC3339)
	end := now.Add(time.Hour).Format(time.RFC3339)
	tr := newMockTransport()
	tr.responses["ViewerDropsDashboard"] = `{"data":{"currentUser":{"dropCampaigns":[]}}}`
	tr.responses["Inventory"] = fmt.Sprintf(`{"data":{"currentUser":{"inventory":{"dropCampaignsInProgress":[{"id":"campaign","startAt":%q,"endAt":%q,"allow":{"channels":[{"id":"allowed"}]},"timeBasedDrops":[{"id":"reward","startAt":%q,"endAt":%q,"requiredMinutesWatched":90,"self":{"currentMinutesWatched":45,"hasPreconditionsMet":true}}]}]}}}}`, start, end, start, end)

	client := newTestClient(t, tr)
	wrong := makeOnlineStreamer("wrong")
	wrong.Settings.ClaimDrops = true
	wrong.Stream.CampaignIDs = []string{"different"}
	blocked := makeOnlineStreamer("blocked")
	blocked.ChannelID = "blocked"
	blocked.Settings.ClaimDrops = true
	blocked.Stream.CampaignIDs = []string{"campaign"}
	right := makeOnlineStreamer("right")
	right.ChannelID = "allowed"
	right.Settings.ClaimDrops = true
	right.Stream.CampaignIDs = []string{"campaign"}
	streamers := []*model.Streamer{wrong, blocked, right}
	if err := client.SyncCampaigns(context.Background(), streamers); err != nil {
		t.Fatal(err)
	}
	set := selectWatchSetAt(streamers, WatchOptions{MaxWatch: 1, Priorities: []model.Priority{model.PriorityUrgentRewards, model.PriorityOrder}}, now)
	if got := usernames(set); !slices.Equal(got, []string{"right"}) {
		t.Fatalf("selection = %v; want the allow-listed channel in the inventory campaign", got)
	}
	tr.responses["Inventory"] = `{"data":{"currentUser":{"inventory":{"dropCampaignsInProgress":[]}}}}`
	if err := client.SyncCampaigns(context.Background(), streamers); err != nil {
		t.Fatal(err)
	}
	set = selectWatchSetAt(streamers, WatchOptions{MaxWatch: 1, Priorities: []model.Priority{model.PriorityUrgentRewards, model.PriorityOrder}}, now)
	if got := usernames(set); !slices.Equal(got, []string{"wrong"}) {
		t.Fatalf("selection = %v; stale inventory must not retain urgency", got)
	}
}

func TestEmptyDashboardAndInventorySkipVanishDetectionAndPrefetch(t *testing.T) {
	t.Parallel()

	tr := newMockTransport()
	tr.responses["ViewerDropsDashboard"] = `{"data":{"currentUser":{"dropCampaigns":[]}}}`
	tr.responses["Inventory"] = `{"data":{"currentUser":{"inventory":{"dropCampaignsInProgress":[]}}}}`
	client := newTestClient(t, tr)
	client.dropSeenCount.Store("ended", 2)
	streamer := makeOnlineStreamer("idle")
	streamer.Settings.ClaimDrops = true
	streamer.Stream.Campaigns = []model.Campaign{{ID: "ended"}}

	if err := client.SyncCampaigns(context.Background(), []*model.Streamer{streamer}); err != nil {
		t.Fatal(err)
	}
	if calls := tr.callCount(constants.GQLDropsHighlightServiceAvailableDrops.OperationName); calls != 0 {
		t.Fatalf("available campaigns calls = %d; want 0", calls)
	}
	if _, vanished := client.dropVanished.Load("ended"); vanished {
		t.Fatal("empty dashboard and inventory marked an ended drop as vanished")
	}
	if len(streamer.Stream.Campaigns) != 0 {
		t.Fatal("ended campaign was not removed from the streamer")
	}
}
