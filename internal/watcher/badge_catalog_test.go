package watcher

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestBadgeRewardMatching(t *testing.T) {
	for _, tc := range []struct {
		reward, game, badge string
		want                bool
	}{
		{"Two Point Pickle Chat Badge", "Two Point Museum", "Two Point Pickle", true},
		{"Pickle", "Two Point Museum", "Two Point Pickle", true},
		{"Solasta 2 Multiplayer", "Solasta II", "Solasta II Multiplayer", true},
		{"Pet", "Path of Exile 2", "Unrelated Pet", false},
	} {
		if got := isBadgeReward(tc.reward, tc.game, tc.badge); got != tc.want {
			t.Errorf("match(%q,%q,%q)=%v", tc.reward, tc.game, tc.badge, got)
		}
	}
}

func TestSlugifyPreservesRomanNumerals(t *testing.T) {
	if got := slugify("Solasta II"); got != "solasta-ii" {
		t.Fatalf("slugify(Solasta II)=%q", got)
	}
}

func TestLoadBadgeCampaignsAcceptsSingleChannel(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/drops", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"games":[{"game":"WARDOGS","campaigns":[{"id":"c1","name":"Launch","ends_at":"2027-01-01T00:00:00Z","all_channels":false,"channels":"wardogs","drops":[{"name":"WARDOG","requirement":"Watch 30m"}]}]}]}`))
	})
	mux.HandleFunc("/badges", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"sets":[{"versions":[{"title":"WARDOG"}]}]}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	got, err := loadBadgeCampaigns(context.Background(), srv.Client(), srv.URL+"/drops", srv.URL+"/badges", time.Date(2026, 9, 11, 0, 0, 0, 0, time.UTC), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || len(got[0].Channels) != 1 || got[0].Channels[0] != "wardogs" {
		t.Fatalf("unexpected campaigns: %#v", got)
	}
}

func TestLoadBadgeCampaignsUsesCanonicalSourceSlug(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/drops", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"games":[{"source":"https://twitchdrops.app/game/canonical-slug","game":"Different Display Name","campaigns":[{"id":"c1","name":"Launch","ends_at":"2027-01-01T00:00:00Z","all_channels":true,"drops":[{"name":"Launch Badge","requirement":"Watch 1h"}]}]}]}`))
	})
	mux.HandleFunc("/badges", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"sets":[{"versions":[{"title":"Launch Badge"}]}]}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	got, err := loadBadgeCampaigns(context.Background(), srv.Client(), srv.URL+"/drops", srv.URL+"/badges", time.Date(2026, 9, 18, 0, 0, 0, 0, time.UTC), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].GameSlug != "canonical-slug" {
		t.Fatalf("unexpected campaigns: %#v", got)
	}
}

func TestLoadBadgeCampaignsSkipsAmbiguousDescriptionFallback(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/drops", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"games":[{"source":"https://twitchdrops.app/game/special-events","game":"Special Events","campaigns":[{"id":"later","name":"First Partners Collection","ends_at":"2027-02-01T00:00:00Z","all_channels":true,"drops":[{"name":"Great Ball","requirement":"Watch 20m"}]},{"id":"sooner","name":"First Partners Collection","ends_at":"2027-01-01T00:00:00Z","all_channels":true,"drops":[{"name":"Great Ball","requirement":"Watch 20m"}]}]}]}`))
	})
	mux.HandleFunc("/badges", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"sets":[{"versions":[{"title":"Pichu","description":"This badge was earned during the Pokémon First Partners Collection campaign."}]},{"versions":[{"title":"Bulbasaur","description":"This badge was earned during the Pokémon First Partners Collection campaign."}]}]}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	var loggedCampaign, loggedReward, loggedReason string
	got, err := loadBadgeCampaigns(context.Background(), srv.Client(), srv.URL+"/drops", srv.URL+"/badges", time.Date(2026, 9, 18, 0, 0, 0, 0, time.UTC), func(campaign, reward, reason string, _ []string) {
		loggedCampaign, loggedReward, loggedReason = campaign, reward, reason
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 || loggedCampaign != "First Partners Collection" || loggedReward != "Great Ball" || loggedReason != "ambiguous badge matches" {
		t.Fatalf("ambiguous campaign was not skipped and logged: campaigns=%#v log=(%q,%q,%q)", got, loggedCampaign, loggedReward, loggedReason)
	}
}

func TestLoadBadgeCampaignsKeepsCleanMatchWhenAnotherDropIsAmbiguous(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/drops", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"games":[{"game":"Special Events","campaigns":[{"id":"mixed","name":"Community Celebration","ends_at":"2027-01-01T00:00:00Z","all_channels":true,"drops":[{"name":"Clean Badge","requirement":"Watch 10m"},{"name":"Mystery Reward","requirement":"Watch 20m"}]}]}]}`))
	})
	mux.HandleFunc("/badges", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"sets":[{"versions":[{"title":"Clean Badge","description":"Watch reward."},{"title":"Ambiguous One","description":"Community Celebration reward."},{"title":"Ambiguous Two","description":"Community Celebration reward."}]}]}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	logs := 0
	got, err := loadBadgeCampaigns(context.Background(), srv.Client(), srv.URL+"/drops", srv.URL+"/badges", time.Date(2026, 9, 18, 0, 0, 0, 0, time.UTC), func(_, _, _ string, _ []string) { logs++ })
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || len(got[0].BadgeNames) != 1 || got[0].BadgeNames[0] != "Clean Badge" || logs != 1 {
		t.Fatalf("clean drop was not preserved: campaigns=%#v logs=%d", got, logs)
	}
}

func TestLoadBadgeCampaignsMatchesCampaignDescription(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/drops", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"games":[{"source":"https://twitchdrops.app/game/special-events","game":"Special Events","campaigns":[{"id":"c1","name":"Community Celebration","ends_at":"2027-01-01T00:00:00Z","all_channels":true,"drops":[{"name":"Mystery Reward","requirement":"Watch 15m"}]}]}]}`))
	})
	mux.HandleFunc("/badges", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"sets":[{"versions":[{"title":"Celebration Badge","description":"Earned during the Community Celebration campaign."}]}]}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	got, err := loadBadgeCampaigns(context.Background(), srv.Client(), srv.URL+"/drops", srv.URL+"/badges", time.Date(2026, 9, 18, 0, 0, 0, 0, time.UTC), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || len(got[0].BadgeNames) != 1 || got[0].BadgeNames[0] != "Celebration Badge" {
		t.Fatalf("unexpected campaigns: %#v", got)
	}
}

func TestLoadBadgeCampaignsDoesNotConfusePaidBadgeWithWatchCampaign(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/drops", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"games":[{"source":"https://twitchdrops.app/game/grand-theft-auto-v","game":"Grand Theft Auto V","campaigns":[{"id":"watch","name":"nopixel V","ends_at":"2027-01-01T00:00:00Z","all_channels":true,"drops":[{"name":"GTA$250K","requirement":"Watch 1h"}]}]}]}`))
	})
	mux.HandleFunc("/badges", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"sets":[{"versions":[{"title":"nopixel V Launch","description":"This badge was earned by subscribing during the nopixel V Launch."}]}]}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	var skippedReward, skippedReason string
	got, err := loadBadgeCampaigns(context.Background(), srv.Client(), srv.URL+"/drops", srv.URL+"/badges", time.Date(2026, 9, 18, 0, 0, 0, 0, time.UTC), func(_, reward, reason string, _ []string) {
		skippedReward, skippedReason = reward, reason
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("paid badge matched watch campaign: %#v", got)
	}
	if skippedReward != "GTA$250K" || skippedReason != "no badge match" {
		t.Fatalf("missing catalog skip context: reward=%q reason=%q", skippedReward, skippedReason)
	}
}

func TestBadgeRequiresPaymentUsesWholeWords(t *testing.T) {
	for _, tc := range []struct {
		description string
		want        bool
	}{
		{"Earned with 100 bits", true},
		{"Available to Prime subscribers", true},
		{"Gift a subscription", true},
		{"Earned through donations and purchases", true},
		{"Available after buying a tier", true},
		{"Buy the Deluxe edition to earn this badge", true},
		{"Cheer 500 to unlock", true},
		{"Prime Gaming reward", true},
		{"Paid reward for members", true},
		{"Watch a bit and pay attention to the prime target", false},
		{"Cheers to everyone reaching tier three", false},
		{"Reach tier 3 of the watch marathon", false},
		{"Explore the orbiting station", false},
		{"Watch the cheerfully hosted stream", false},
	} {
		if got := badgeRequiresPayment(tc.description); got != tc.want {
			t.Errorf("badgeRequiresPayment(%q)=%v, want %v", tc.description, got, tc.want)
		}
	}
}

func TestOwnsCampaignBadgeRequiresEveryKnownBadge(t *testing.T) {
	campaign := badgeCampaign{GameName: "Game", BadgeNames: []string{"First Badge", "Second Badge"}}
	if ownsCampaignBadge(campaign, map[string]struct{}{"First Badge": {}}) {
		t.Fatal("partially owned campaign was treated as complete")
	}
	if !ownsCampaignBadge(campaign, map[string]struct{}{"First Badge": {}, "Second Badge": {}}) {
		t.Fatal("fully owned campaign was not treated as complete")
	}
	if ownsCampaignBadge(badgeCampaign{}, map[string]struct{}{}) {
		t.Fatal("campaign without known badges was treated as owned")
	}
	if ownsCampaignBadge(badgeCampaign{BadgeNames: []string{"Known"}, HasAmbiguousDrops: true}, map[string]struct{}{"Known": {}}) {
		t.Fatal("campaign with an ambiguous unresolved drop was treated as complete")
	}
}

func TestMergeBadgeCampaignsPreservesScopeAndEarliestEnd(t *testing.T) {
	late := badgeCampaign{ID: "late", EndsAt: time.Date(2027, 2, 1, 0, 0, 0, 0, time.UTC), Channels: []string{"MixedCase"}, BadgeNames: []string{"Launch Badge"}}
	early := badgeCampaign{ID: "early", EndsAt: time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC), AllChannels: true, Channels: []string{"other"}, BadgeNames: []string{"launch"}}
	got := mergeBadgeCampaigns(late, early)
	if got.ID != "early" || !got.AllChannels || len(got.Channels) != 2 || !sameBadgeNames(got.BadgeNames, late.BadgeNames) {
		t.Fatalf("unexpected merged campaign: %#v", got)
	}
}

func TestLoadBadgeCampaignsMergesDuplicateCampaignID(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/drops", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"games":[{"game":"Game","campaigns":[{"id":"same","name":"Launch","ends_at":"2027-02-01T00:00:00Z","channels":["One"],"drops":[{"name":"First Badge","requirement":"Watch 10m"}]},{"id":"same","name":"Launch","ends_at":"2027-01-01T00:00:00Z","all_channels":true,"channels":["Two"],"drops":[{"name":"Second Badge","requirement":"Watch 20m"}]}]}]}`))
	})
	mux.HandleFunc("/badges", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"sets":[{"versions":[{"title":"First Badge"},{"title":"Second Badge"}]}]}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	got, err := loadBadgeCampaigns(context.Background(), srv.Client(), srv.URL+"/drops", srv.URL+"/badges", time.Date(2026, 9, 18, 0, 0, 0, 0, time.UTC), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].ID != "same" || !got[0].AllChannels || len(got[0].Channels) != 2 || len(got[0].BadgeNames) != 2 || got[0].EndsAt.Month() != time.January {
		t.Fatalf("duplicate campaign ID was not merged: %#v", got)
	}
}

func TestLoadBadgeCampaignsSkipsMissingCampaignID(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/drops", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"games":[{"game":"Game","campaigns":[{"name":"Missing ID","ends_at":"2027-01-01T00:00:00Z","all_channels":true,"drops":[{"name":"Badge","requirement":"Watch 10m"}]}]}]}`))
	})
	mux.HandleFunc("/badges", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"sets":[{"versions":[{"title":"Badge"}]}]}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	var reason string
	got, err := loadBadgeCampaigns(context.Background(), srv.Client(), srv.URL+"/drops", srv.URL+"/badges", time.Date(2026, 9, 18, 0, 0, 0, 0, time.UTC), func(_, _, gotReason string, _ []string) { reason = gotReason })
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 || reason != "missing campaign ID" {
		t.Fatalf("missing-ID campaign was not skipped observably: campaigns=%#v reason=%q", got, reason)
	}
}

func TestCompletedCampaigns(t *testing.T) {
	raw := json.RawMessage(`{"completedRewardCampaigns":[{"campaign":{"name":"Launch Badge","endAt":"2026-09-12T00:00:00Z","game":{"displayName":"Some Game"}}}]}`)
	got, err := completedCampaigns(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].GameSlug != "some-game" {
		t.Fatalf("unexpected signatures: %#v", got)
	}
	c := badgeCampaign{Name: "Launch Badge", GameSlug: "some-game", EndsAt: time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)}
	if !campaignCompleted(c, got) {
		t.Fatal("expected campaign to be completed")
	}
}

func TestCompletedCampaignsRejectsMalformedInventory(t *testing.T) {
	if _, err := completedCampaigns(json.RawMessage(`{`)); err == nil {
		t.Fatal("expected malformed inventory error")
	}
}

func TestCampaignCompletedUsesIDOrEndTimeTolerance(t *testing.T) {
	endsAt := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
	campaign := badgeCampaign{ID: "campaign-id", Name: "Launch", GameSlug: "game", EndsAt: endsAt}
	if !campaignCompleted(campaign, []completedCampaignSignature{{ID: "campaign-id"}}) {
		t.Fatal("expected campaign ID match")
	}
	completed := completedCampaignSignature{GameSlug: "game", Name: "launch", EndsAt: endsAt.Add(4 * time.Second)}
	if !campaignCompleted(campaign, []completedCampaignSignature{completed}) {
		t.Fatal("expected completion within five-second tolerance")
	}
	completed.EndsAt = endsAt.Add(6 * time.Second)
	if campaignCompleted(campaign, []completedCampaignSignature{completed}) {
		t.Fatal("unexpected completion outside tolerance")
	}
}
