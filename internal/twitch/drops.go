package twitch

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/Guliveer/twitch-miner-go/internal/constants"
	"github.com/Guliveer/twitch-miner-go/internal/model"
)

// SyncCampaigns synchronizes drop campaigns with the user's inventory
// and updates streamer campaign data. This is called periodically by the miner.
func (c *Client) SyncCampaigns(ctx context.Context, streamers []*model.Streamer) error {
	if err := c.ClaimAllDropsFromInventory(ctx); err != nil {
		c.Log.Warn("Failed to claim drops from inventory", "error", err)
	}

	dashboardCampaigns, err := c.GQL.GetDropsDashboard(ctx, "")
	if err != nil {
		return fmt.Errorf("getting drops dashboard: %w", err)
	}

	if len(dashboardCampaigns) == 0 {
		return nil
	}

	campaignIDs := make([]string, 0, len(dashboardCampaigns))
	for _, raw := range dashboardCampaigns {
		var campaignID struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(raw, &campaignID); err == nil && campaignID.ID != "" {
			campaignIDs = append(campaignIDs, campaignID.ID)
		}
	}

	campaignDetails, err := c.GQL.GetDropCampaignDetailsBatch(ctx, campaignIDs, c.Auth.UserID())
	if err != nil {
		return fmt.Errorf("getting campaign details: %w", err)
	}

	var allCampaigns []*model.Campaign
	activeCampaigns := make([]*model.Campaign, 0, len(campaignDetails))
	for _, raw := range campaignDetails {
		if raw == nil {
			continue
		}
		campaign, err := parseCampaign(raw)
		if err != nil {
			c.Log.Debug("Failed to parse campaign", "error", err)
			continue
		}
		allCampaigns = append(allCampaigns, campaign)
		if campaign.IsWithinTimeWindow {
			campaign.ClearDrops()
			if len(campaign.Drops) > 0 {
				activeCampaigns = append(activeCampaigns, campaign)
			}
		}
	}

	activeCampaigns, err = c.syncCampaignsWithInventory(ctx, activeCampaigns)
	if err != nil {
		c.Log.Warn("Failed to sync campaigns with inventory", "error", err)
	}

	// Check campaign reminders (replaces the old NEW_CAMPAIGN detection).
	// On the first sync, sends catch-up notifications for upcoming campaigns
	// and seeds knownCampaigns. On subsequent syncs, fires on_detection for
	// newly seen campaigns and time-based reminders in their window.
	isFirstSync := !c.campaignsInitialized.Load()
	c.checkCampaignReminders(ctx, allCampaigns, streamers, isFirstSync)
	if isFirstSync {
		c.campaignsInitialized.Store(true)
	}

	for _, streamer := range streamers {
		streamer.Mu.Lock()
		if streamer.DropsCondition() {
			var matchingCampaigns []model.Campaign
			for _, campaign := range activeCampaigns {
				if len(campaign.Drops) > 0 && campaignMatchesStreamer(campaign, streamer) {
					matchingCampaigns = append(matchingCampaigns, *campaign)
				}
			}
			streamer.Stream.Campaigns = matchingCampaigns
		}
		streamer.Mu.Unlock()
	}

	return nil
}

func (c *Client) syncCampaignsWithInventory(ctx context.Context, campaigns []*model.Campaign) ([]*model.Campaign, error) {
	inventoryData, err := c.GQL.GetDropsInventory(ctx)
	if err != nil {
		return campaigns, fmt.Errorf("getting inventory: %w", err)
	}

	if inventoryData == nil {
		return campaigns, nil
	}

	var inventory struct {
		DropCampaignsInProgress []struct {
			ID             string `json:"id"`
			TimeBasedDrops []struct {
				ID              string `json:"id"`
				Name            string `json:"name"`
				RequiredMinutes int    `json:"requiredMinutesWatched"`
				Self            *struct {
					HasPreconditionsMet   bool   `json:"hasPreconditionsMet"`
					CurrentMinutesWatched int    `json:"currentMinutesWatched"`
					DropInstanceID        string `json:"dropInstanceID"`
					IsClaimed             bool   `json:"isClaimed"`
				} `json:"self"`
			} `json:"timeBasedDrops"`
		} `json:"dropCampaignsInProgress"`
	}

	if err := json.Unmarshal(inventoryData, &inventory); err != nil {
		return campaigns, fmt.Errorf("parsing inventory: %w", err)
	}

	if inventory.DropCampaignsInProgress == nil {
		return campaigns, nil
	}

	for i, campaign := range campaigns {
		campaign.ClearDrops()
		for _, progress := range inventory.DropCampaignsInProgress {
			if progress.ID == campaign.ID {
				campaigns[i].InInventory = true
				for _, timeDrop := range progress.TimeBasedDrops {
					for _, drop := range campaigns[i].Drops {
						if drop.ID == timeDrop.ID && timeDrop.Self != nil {
							drop.Update(
								timeDrop.Self.HasPreconditionsMet,
								timeDrop.Self.CurrentMinutesWatched,
								timeDrop.Self.DropInstanceID,
								timeDrop.Self.IsClaimed,
							)
						}
					}
				}
				campaigns[i].ClearDrops()
				break
			}
		}
	}

	return campaigns, nil
}

// ClaimDrop claims a single drop reward.
func (c *Client) ClaimDrop(ctx context.Context, dropInstanceID string) error {
	c.Log.Info("Claiming drop", "drop_instance_id", dropInstanceID)
	claimed, err := c.GQL.ClaimDropRewards(ctx, dropInstanceID)
	if err != nil {
		return fmt.Errorf("claiming drop %s: %w", dropInstanceID, err)
	}
	if !claimed {
		return fmt.Errorf("drop %s was not claimed", dropInstanceID)
	}
	return nil
}

// ClaimAllDropsFromInventory claims all pending drops from the user's inventory.
func (c *Client) ClaimAllDropsFromInventory(ctx context.Context) error {
	inventoryData, err := c.GQL.GetDropsInventory(ctx)
	if err != nil {
		return fmt.Errorf("getting inventory: %w", err)
	}

	if inventoryData == nil {
		return nil
	}

	var inventory struct {
		DropCampaignsInProgress []struct {
			Game *struct {
				Name string `json:"name"`
				Slug string `json:"slug"`
			} `json:"game"`
			TimeBasedDrops []struct {
				ID           string `json:"id"`
				Name         string `json:"name"`
				BenefitEdges []struct {
					Benefit struct {
						Name string `json:"name"`
					} `json:"benefit"`
				} `json:"benefitEdges"`
				Self *struct {
					DropInstanceID string `json:"dropInstanceID"`
					IsClaimed      bool   `json:"isClaimed"`
				} `json:"self"`
			} `json:"timeBasedDrops"`
		} `json:"dropCampaignsInProgress"`
	}

	if err := json.Unmarshal(inventoryData, &inventory); err != nil {
		return fmt.Errorf("parsing inventory: %w", err)
	}

	if inventory.DropCampaignsInProgress == nil {
		return nil
	}

	for _, campaign := range inventory.DropCampaignsInProgress {
		for _, drop := range campaign.TimeBasedDrops {
			if drop.Self == nil {
				continue
			}
			if drop.Self.IsClaimed || drop.Self.DropInstanceID == "" {
				continue
			}

			// Dedup by drop definition ID — the same drop can appear across
			// multiple campaigns with different instance IDs. Claiming one
			// instance is enough; Twitch marks the rest as claimed server-side.
			if _, alreadyAttempted := c.claimedDrops.Load(drop.ID); alreadyAttempted {
				continue
			}

			categoryName := ""
			if campaign.Game != nil {
				categoryName = campaign.Game.Slug
				if categoryName == "" {
					categoryName = campaign.Game.Name
				}
			}

			// Use the benefit/reward name if available, fall back to the time-based drop name.
			dropName := drop.Name
			if len(drop.BenefitEdges) > 0 && drop.BenefitEdges[0].Benefit.Name != "" {
				dropName = drop.BenefitEdges[0].Benefit.Name
			}

			c.Log.Event(ctx, model.EventDropClaim, "Claiming drop from inventory",
				"drop", dropName,
				"category", categoryName)

			// Mark as attempted before calling the API to prevent duplicates.
			c.claimedDrops.Store(drop.ID, true)

			claimed, err := c.GQL.ClaimDropRewards(ctx, drop.Self.DropInstanceID)
			if err != nil {
				c.Log.Warn("Failed to claim drop from inventory",
					"drop", dropName, "error", err)
			} else if !claimed {
				c.Log.Warn("Drop claim was not accepted",
					"drop", dropName)
			}

			sleepDuration := time.Duration(5+rand.IntN(5)) * time.Second
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(sleepDuration):
			}
		}
	}

	return nil
}

func parseCampaign(raw json.RawMessage) (*model.Campaign, error) {
	var data struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Status  string `json:"status"`
		StartAt string `json:"startAt"`
		EndAt   string `json:"endAt"`
		Game    *struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			DisplayName string `json:"displayName"`
			Slug        string `json:"slug"`
		} `json:"game"`
		Allow *struct {
			Channels []struct {
				ID string `json:"id"`
			} `json:"channels"`
		} `json:"allow"`
		TimeBasedDrops []struct {
			ID              string `json:"id"`
			Name            string `json:"name"`
			RequiredMinutes int    `json:"requiredMinutesWatched"`
			StartAt         string `json:"startAt"`
			EndAt           string `json:"endAt"`
			BenefitEdges    []struct {
				Benefit struct {
					Name string `json:"name"`
				} `json:"benefit"`
			} `json:"benefitEdges"`
		} `json:"timeBasedDrops"`
	}

	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, fmt.Errorf("parsing campaign JSON: %w", err)
	}

	var gameInfo *model.GameInfo
	if data.Game != nil {
		gameInfo = &model.GameInfo{
			ID:          data.Game.ID,
			Name:        data.Game.Name,
			DisplayName: data.Game.DisplayName,
			Slug:        data.Game.Slug,
		}
	}

	startAt, _ := time.Parse(time.RFC3339, data.StartAt)
	endAt, _ := time.Parse(time.RFC3339, data.EndAt)

	var channels []string
	if data.Allow != nil {
		for _, channel := range data.Allow.Channels {
			channels = append(channels, channel.ID)
		}
	}

	campaign := model.NewCampaign(data.ID, data.Name, data.Status, gameInfo, startAt, endAt, channels)

	for _, timeDrop := range data.TimeBasedDrops {
		dropStart, _ := time.Parse(time.RFC3339, timeDrop.StartAt)
		dropEnd, _ := time.Parse(time.RFC3339, timeDrop.EndAt)

		var benefits []string
		for _, benefitEdge := range timeDrop.BenefitEdges {
			benefits = append(benefits, benefitEdge.Benefit.Name)
		}

		drop := model.NewDrop(timeDrop.ID, timeDrop.Name, benefits, timeDrop.RequiredMinutes, dropStart, dropEnd)
		campaign.Drops = append(campaign.Drops, drop)
	}

	return campaign, nil
}

func campaignMatchesStreamer(campaign *model.Campaign, streamer *model.Streamer) bool {
	if campaign.Game != nil && streamer.Stream.Game != nil {
		if campaign.Game.Name != streamer.Stream.Game.Name {
			return false
		}
	}

	for _, id := range streamer.Stream.CampaignIDs {
		if id == campaign.ID {
			return true
		}
	}

	return false
}

// checkCampaignReminders processes campaign reminder notifications.
// On the first sync it sends one catch-up notification per upcoming campaign
// and seeds knownCampaigns. On subsequent syncs it fires on_detection for
// newly seen campaigns and time-based reminders whose window matches now.
func (c *Client) checkCampaignReminders(ctx context.Context, campaigns []*model.Campaign, streamers []*model.Streamer, isFirstSync bool) {
	now := time.Now()
	syncInterval := constants.DefaultCampaignSyncInterval

	for _, campaign := range campaigns {
		gameName := campaignGameName(campaign)

		// Find first matching streamer with reminders configured.
		var reminders *model.CampaignReminderConfig
		for _, streamer := range streamers {
			if !streamer.CampaignReminders.HasReminders() {
				continue
			}
			if campaignMatchesCategory(campaign, streamer) {
				reminders = streamer.CampaignReminders
				break
			}
		}

		if reminders == nil {
			// Seed knownCampaigns even without reminders so future
			// on_detection doesn't fire for pre-existing campaigns.
			if isFirstSync {
				c.knownCampaigns.Store(campaign.ID, true)
			}
			continue
		}

		if isFirstSync {
			// Seed knownCampaigns so on_detection doesn't double-fire.
			c.knownCampaigns.Store(campaign.ID, true)

			// Catch-up: one notification per upcoming campaign.
			if campaign.StartAt.After(now) {
				timeUntil := time.Until(campaign.StartAt)
				c.Log.Event(ctx, model.EventCampaignReminder,
					fmt.Sprintf("Upcoming campaign: %s (starts in %s)", campaign.Name, formatDuration(timeUntil)),
					"campaign", campaign.Name,
					"category", gameName,
					"starts_at", campaign.StartAt.Format("02 Jan 15:04"),
					"ends_at", campaign.EndAt.Format("02 Jan 15:04"),
				)
			}
			continue
		}

		// on_detection: fire once per session for newly seen campaigns.
		if reminders.OnDetection {
			if _, seen := c.knownCampaigns.LoadOrStore(campaign.ID, true); !seen {
				c.Log.Event(ctx, model.EventCampaignReminder,
					fmt.Sprintf("New campaign: %s", campaign.Name),
					"campaign", campaign.Name,
					"category", gameName,
					"starts_at", campaign.StartAt.Format("02 Jan 15:04"),
					"ends_at", campaign.EndAt.Format("02 Jan 15:04"),
				)
			}
		}

		// Time-based reminders: only for campaigns that haven't started yet.
		if campaign.StartAt.After(now) {
			for _, dur := range reminders.Durations {
				reminderTime := campaign.StartAt.Add(-dur)
				diff := now.Sub(reminderTime)
				if diff >= 0 && diff < syncInterval {
					c.Log.Event(ctx, model.EventCampaignReminder,
						fmt.Sprintf("Campaign starts in %s: %s", formatDuration(dur), campaign.Name),
						"campaign", campaign.Name,
						"category", gameName,
						"starts_at", campaign.StartAt.Format("02 Jan 15:04"),
						"ends_at", campaign.EndAt.Format("02 Jan 15:04"),
					)
					break // One time-based reminder per campaign per sync
				}
			}
		}
	}
}

// campaignMatchesCategory checks if a campaign's game matches a streamer's category.
// Uses slug comparison for category-watched streamers, falls back to game name.
func campaignMatchesCategory(campaign *model.Campaign, streamer *model.Streamer) bool {
	if campaign.Game == nil {
		return false
	}
	// Match by category slug (primary, for category-watched streamers).
	if streamer.CategorySlug != "" && campaign.Game.Slug != "" {
		return strings.EqualFold(campaign.Game.Slug, streamer.CategorySlug)
	}
	// Fallback: match by game name.
	if streamer.Stream != nil && streamer.Stream.Game != nil {
		return campaign.Game.Name == streamer.Stream.Game.Name
	}
	return false
}

// campaignGameName returns the best available game identifier for a campaign.
func campaignGameName(campaign *model.Campaign) string {
	if campaign.Game == nil {
		return ""
	}
	if campaign.Game.Slug != "" {
		return campaign.Game.Slug
	}
	return campaign.Game.Name
}

// formatDuration formats a duration into a human-readable string (e.g., "3 days", "1 hour").
func formatDuration(d time.Duration) string {
	if d >= 24*time.Hour {
		days := int(d.Hours() / 24)
		if days == 1 {
			return "1 day"
		}
		return fmt.Sprintf("%d days", days)
	}
	if d >= time.Hour {
		hours := int(d.Hours())
		if hours == 1 {
			return "1 hour"
		}
		return fmt.Sprintf("%d hours", hours)
	}
	minutes := int(d.Minutes())
	if minutes <= 1 {
		return "1 minute"
	}
	return fmt.Sprintf("%d minutes", minutes)
}
