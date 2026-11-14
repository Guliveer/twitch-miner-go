package twitch

import (
	"context"
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/Guliveer/twitch-miner-go/internal/model"
)

func TestUrgentRewardUsesInventoryWhenDashboardIsEmpty(t *testing.T) {
	t.Parallel()

	now := time.Now()
	start := now.Add(-time.Hour).Format(time.RFC3339)
	end := now.Add(time.Hour).Format(time.RFC3339)
	tr := newMockTransport()
	tr.responses["ViewerDropsDashboard"] = `{"data":{"currentUser":{"dropCampaigns":[]}}}`
	tr.responses["Inventory"] = fmt.Sprintf(`{"data":{"currentUser":{"inventory":{"dropCampaignsInProgress":[{"id":"campaign","startAt":%q,"endAt":%q,"timeBasedDrops":[{"id":"reward","startAt":%q,"endAt":%q,"requiredMinutesWatched":90,"self":{"currentMinutesWatched":45,"hasPreconditionsMet":true}}]}]}}}}`, start, end, start, end)

	client := newTestClient(t, tr)
	wrong := makeOnlineStreamer("wrong")
	wrong.Settings.ClaimDrops = true
	wrong.Stream.CampaignIDs = []string{"different"}
	right := makeOnlineStreamer("right")
	right.Settings.ClaimDrops = true
	right.Stream.CampaignIDs = []string{"campaign"}
	streamers := []*model.Streamer{wrong, right}
	if err := client.SyncCampaigns(context.Background(), streamers); err != nil {
		t.Fatal(err)
	}
	set := selectWatchSetAt(streamers, WatchOptions{MaxWatch: 1, Priorities: []model.Priority{model.PriorityUrgentRewards, model.PriorityOrder}}, now)
	if got := usernames(set); !slices.Equal(got, []string{"right"}) {
		t.Fatalf("selection = %v; want channel with inventory campaign and enough time", got)
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
