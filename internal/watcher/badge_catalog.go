package watcher

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	defaultDropsCatalogURL  = "https://gist.githubusercontent.com/zarmstrong/72433778ae596815f4c6ff5e1d278cd2/raw/twitch-drops.json"
	defaultBadgesCatalogURL = "https://gist.githubusercontent.com/zarmstrong/d4fc5f87e2a5258a28421f7fdb8037d6/raw/twitch-badges.json"
	maxCatalogBytes         = 4 << 20
)

type badgeCampaign struct {
	ID, Name, GameName, GameSlug string
	StartsAt, EndsAt             time.Time
	AllChannels                  bool
	HasAmbiguousDrops            bool
	Channels                     []string
	BadgeNames                   []string
}

type flexibleStrings []string

func (s *flexibleStrings) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*s = nil
		return nil
	}
	var many []string
	if err := json.Unmarshal(data, &many); err == nil {
		*s = many
		return nil
	}
	var one string
	if err := json.Unmarshal(data, &one); err != nil {
		return err
	}
	if strings.TrimSpace(one) == "" {
		*s = nil
	} else {
		*s = []string{one}
	}
	return nil
}

type dropsCatalog struct {
	Games []struct {
		Game      string `json:"game"`
		Source    string `json:"source"`
		Campaigns []struct {
			ID          string          `json:"id"`
			Name        string          `json:"name"`
			StartsAt    *time.Time      `json:"starts_at"`
			EndsAt      *time.Time      `json:"ends_at"`
			AllChannels bool            `json:"all_channels"`
			Channels    flexibleStrings `json:"channels"`
			Drops       []struct {
				Name        string `json:"name"`
				Requirement string `json:"requirement"`
			} `json:"drops"`
		} `json:"campaigns"`
	} `json:"games"`
}

type badgesCatalog struct {
	Sets []struct {
		Versions []struct {
			Title       string `json:"title"`
			Description string `json:"description"`
		} `json:"versions"`
	} `json:"sets"`
}

type badgeDefinition struct {
	Title       string
	Description string
}

var badgeWordsRE = regexp.MustCompile(`[a-z0-9]+`)

var romanBadgeNumbers = map[string]string{
	"i": "1", "ii": "2", "iii": "3", "iv": "4", "v": "5",
	"vi": "6", "vii": "7", "viii": "8", "ix": "9", "x": "10",
}

func words(s string) []string { return badgeWordsRE.FindAllString(strings.ToLower(s), -1) }

func normalizedBadgeWords(s string) []string {
	result := words(s)
	for i, word := range result {
		if number, ok := romanBadgeNumbers[word]; ok {
			result[i] = number
		}
	}
	return result
}
func comparableWords(s string) []string {
	w := normalizedBadgeWords(s)
	if len(w) >= 2 && w[len(w)-2] == "chat" && w[len(w)-1] == "badge" {
		w = w[:len(w)-2]
	} else if len(w) > 0 && w[len(w)-1] == "badge" {
		w = w[:len(w)-1]
	}
	return w
}
func equalWords(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func isBadgeReward(reward, game, badge string) bool {
	rw := comparableWords(reward)
	bw := comparableWords(badge)
	if len(rw) == 0 || len(bw) == 0 {
		return false
	}
	if equalWords(rw, bw) {
		return true
	}
	gw := make(map[string]bool)
	for _, w := range normalizedBadgeWords(game) {
		gw[w] = true
	}
	if len(bw) > len(rw) && equalWords(bw[len(bw)-len(rw):], rw) {
		for _, p := range bw[:len(bw)-len(rw)] {
			if gw[p] {
				return true
			}
		}
	}
	if len(rw) > len(bw) && equalWords(rw[len(rw)-len(bw):], bw) {
		for _, p := range rw[:len(rw)-len(bw)] {
			if gw[p] {
				return true
			}
		}
	}
	return false
}

func slugify(s string) string { return strings.Join(words(s), "-") }

func canonicalGameSlug(source, game string) string {
	// twitchdrops.app currently exposes canonical game pages as /game/<slug>.
	// If that contract changes, fall back to the normalized display name.
	parsed, err := url.Parse(source)
	if err == nil {
		parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
		for i := 0; i+1 < len(parts); i++ {
			if parts[i] != "game" {
				continue
			}
			slug := strings.ToLower(strings.TrimSpace(parts[i+1]))
			if slug != "" && slugify(slug) == slug {
				return slug
			}
		}
	}
	return slugify(game)
}

// Keep payment detection narrow and Twitch-specific. Word boundaries avoid
// prefix false positives, while roots cover regular inflections without an
// open-ended dictionary of everyday words.
var paidBadgeWordRE = regexp.MustCompile(`^(?:subs?|subscrib(?:e[ds]?|ing|ers?)|subscriptions?|gift(?:s|ed|ing)?|bits|donat(?:e[ds]?|ing|ions?)|purchas(?:e[ds]?|ing))$`)

// Everyday payment-adjacent words are only considered paid when they appear
// in an unambiguous entitlement phrase. This keeps "watch a bit" and "pay
// attention" eligible while rejecting descriptions such as "buying a tier".
var paidBadgePhraseRE = regexp.MustCompile(`(?:\b(?:buy|buys|buying|bought)\b(?:\s+[a-z0-9]+){0,3}\s+\b(?:tiers?|game|edition|bundle|subs?|subscriptions?)\b|\bprime\s+(?:gaming|subs?|subscriptions?|rewards?)\b|\bcheer(?:s|ed|ing)?\s+(?:[0-9]+|to\s+unlock|with\s+bits)\b|\b(?:pay|pays|paying|paid)\s+to\s+unlock\b|\b(?:paid|payments?)\s+(?:badges?|rewards?|tiers?|subs?|subscriptions?)\b)`)

func badgeRequiresPayment(description string) bool {
	descriptionWords := words(description)
	for _, word := range descriptionWords {
		if paidBadgeWordRE.MatchString(word) {
			return true
		}
	}
	return paidBadgePhraseRE.MatchString(strings.Join(descriptionWords, " "))
}

func badgeTitlesForWatchReward(reward, game, campaign string, badges []badgeDefinition) ([]string, bool) {
	for _, badge := range badges {
		if !badgeRequiresPayment(badge.Description) && isBadgeReward(reward, game, badge.Title) {
			return []string{badge.Title}, false
		}
	}

	campaignKey := strings.Join(words(campaign), " ")
	if len(words(campaign)) < 2 {
		return nil, false
	}
	var matches []string
	for _, badge := range badges {
		if badgeRequiresPayment(badge.Description) {
			continue
		}
		description := strings.Join(words(badge.Description), " ")
		if strings.Contains(description, campaignKey) {
			matches = append(matches, badge.Title)
		}
	}
	if len(matches) > 1 {
		return matches, true
	}
	return matches, false
}

func appendUnique(values []string, additions ...string) []string {
	for _, addition := range additions {
		found := false
		for _, value := range values {
			if strings.EqualFold(value, addition) {
				found = true
				break
			}
		}
		if !found {
			values = append(values, addition)
		}
	}
	return values
}

func fetchJSON(ctx context.Context, client *http.Client, url string, dst any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("catalog returned HTTP %d", resp.StatusCode)
	}
	limited := io.LimitReader(resp.Body, maxCatalogBytes+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return err
	}
	if len(data) > maxCatalogBytes {
		return fmt.Errorf("catalog exceeds %d bytes", maxCatalogBytes)
	}
	if err := json.Unmarshal(data, dst); err != nil {
		return fmt.Errorf("invalid catalog JSON: %w", err)
	}
	return nil
}

type badgeCatalogSkipLogger func(campaign, reward, reason string, matches []string)

func loadBadgeCampaigns(ctx context.Context, client *http.Client, dropsURL, badgesURL string, now time.Time, logSkip badgeCatalogSkipLogger) ([]badgeCampaign, error) {
	if dropsURL == "" {
		dropsURL = defaultDropsCatalogURL
	}
	if badgesURL == "" {
		badgesURL = defaultBadgesCatalogURL
	}
	var dc dropsCatalog
	if err := fetchJSON(ctx, client, dropsURL, &dc); err != nil {
		return nil, fmt.Errorf("drops catalog: %w", err)
	}
	var bc badgesCatalog
	if err := fetchJSON(ctx, client, badgesURL, &bc); err != nil {
		return nil, fmt.Errorf("badges catalog: %w", err)
	}
	var badges []badgeDefinition
	for _, set := range bc.Sets {
		for _, v := range set.Versions {
			if strings.TrimSpace(v.Title) != "" {
				badges = append(badges, badgeDefinition{Title: v.Title, Description: v.Description})
			}
		}
	}
	var result []badgeCampaign
	for _, game := range dc.Games {
		gameSlug := canonicalGameSlug(game.Source, game.Game)
		for _, campaign := range game.Campaigns {
			if campaign.StartsAt != nil && campaign.StartsAt.After(now) {
				continue
			}
			if campaign.EndsAt == nil || !campaign.EndsAt.After(now) {
				continue
			}
			if strings.TrimSpace(campaign.ID) == "" {
				if logSkip != nil {
					logSkip(campaign.Name, "", "missing campaign ID", nil)
				}
				continue
			}
			var matches []string
			watchDrops := 0
			hadAmbiguousDrop := false
			lastUnmatchedDrop := ""
			for _, drop := range campaign.Drops {
				if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(drop.Requirement)), "watch ") {
					continue
				}
				watchDrops++
				rewardMatches, rewardAmbiguous := badgeTitlesForWatchReward(drop.Name, game.Game, campaign.Name, badges)
				if rewardAmbiguous {
					hadAmbiguousDrop = true
					if logSkip != nil {
						logSkip(campaign.Name, drop.Name, "ambiguous badge matches", rewardMatches)
					}
					// Keep clean matches from this campaign; ambiguous drops remain
					// visible in logs but are not treated as owned-badge evidence.
					continue
				}
				if len(rewardMatches) == 0 {
					lastUnmatchedDrop = drop.Name
				}
				matches = appendUnique(matches, rewardMatches...)
			}
			if len(matches) == 0 {
				if watchDrops > 0 && !hadAmbiguousDrop && logSkip != nil {
					logSkip(campaign.Name, lastUnmatchedDrop, "no badge match", nil)
				}
				continue
			}
			sort.Strings(matches)
			entry := badgeCampaign{ID: campaign.ID, Name: campaign.Name, GameName: game.Game, GameSlug: gameSlug, EndsAt: *campaign.EndsAt, AllChannels: campaign.AllChannels, HasAmbiguousDrops: hadAmbiguousDrop, Channels: campaign.Channels, BadgeNames: matches}
			if campaign.StartsAt != nil {
				entry.StartsAt = *campaign.StartsAt
			}
			duplicate := -1
			for i := range result {
				if result[i].ID == entry.ID {
					duplicate = i
					break
				}
			}
			if duplicate < 0 {
				result = append(result, entry)
			} else {
				result[duplicate] = mergeBadgeCampaigns(result[duplicate], entry)
			}
		}
	}
	return result, nil
}

func normalizedBadgeName(name string) string {
	return strings.Join(comparableWords(name), " ")
}

func sameBadgeNames(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	left := make([]string, len(a))
	right := make([]string, len(b))
	for i := range a {
		left[i] = normalizedBadgeName(a[i])
	}
	for i := range b {
		right[i] = normalizedBadgeName(b[i])
	}
	sort.Strings(left)
	sort.Strings(right)
	return equalWords(left, right)
}

func mergeBadgeCampaigns(a, b badgeCampaign) badgeCampaign {
	merged := a
	if b.EndsAt.Before(a.EndsAt) {
		merged = b
	}
	merged.AllChannels = a.AllChannels || b.AllChannels
	merged.HasAmbiguousDrops = a.HasAmbiguousDrops || b.HasAmbiguousDrops
	merged.Channels = appendUnique(append([]string(nil), a.Channels...), b.Channels...)
	merged.BadgeNames = appendUniqueBadgeNames(append([]string(nil), a.BadgeNames...), b.BadgeNames...)
	sort.Strings(merged.BadgeNames)
	return merged
}

func appendUniqueBadgeNames(values []string, additions ...string) []string {
	for _, addition := range additions {
		key := normalizedBadgeName(addition)
		found := false
		for _, value := range values {
			if normalizedBadgeName(value) == key {
				found = true
				break
			}
		}
		if !found {
			values = append(values, addition)
		}
	}
	return values
}

func ownsCampaignBadge(c badgeCampaign, owned map[string]struct{}) bool {
	if len(c.BadgeNames) == 0 || c.HasAmbiguousDrops {
		return false
	}
	for _, reward := range c.BadgeNames {
		ownedReward := false
		for title := range owned {
			if isBadgeReward(reward, c.GameName, title) {
				ownedReward = true
				break
			}
		}
		if !ownedReward {
			return false
		}
	}
	return true
}

type completedCampaignSignature struct {
	ID, GameSlug, Name string
	EndsAt             time.Time
}

func completedCampaigns(raw json.RawMessage) ([]completedCampaignSignature, error) {
	var inv struct {
		Completed []struct {
			Campaign *struct {
				ID     string    `json:"id"`
				Name   string    `json:"name"`
				EndAt  time.Time `json:"endAt"`
				EndsAt time.Time `json:"endsAt"`
				Game   struct {
					Name        string `json:"name"`
					DisplayName string `json:"displayName"`
				} `json:"game"`
			} `json:"campaign"`
		} `json:"completedRewardCampaigns"`
	}
	if err := json.Unmarshal(raw, &inv); err != nil {
		return nil, fmt.Errorf("parse completed reward campaigns: %w", err)
	}
	var result []completedCampaignSignature
	for _, item := range inv.Completed {
		if item.Campaign == nil {
			continue
		}
		end := item.Campaign.EndAt
		if end.IsZero() {
			end = item.Campaign.EndsAt
		}
		game := item.Campaign.Game.DisplayName
		if game == "" {
			game = item.Campaign.Game.Name
		}
		if game != "" && item.Campaign.Name != "" && !end.IsZero() {
			result = append(result, completedCampaignSignature{item.Campaign.ID, slugify(game), strings.ToLower(strings.TrimSpace(item.Campaign.Name)), end})
		}
	}
	return result, nil
}

func campaignCompleted(c badgeCampaign, completed []completedCampaignSignature) bool {
	for _, s := range completed {
		if c.ID != "" && s.ID != "" && c.ID == s.ID {
			return true
		}
		const endTimeTolerance = 5 * time.Second
		if c.GameSlug == s.GameSlug && strings.EqualFold(c.Name, s.Name) && c.EndsAt.Sub(s.EndsAt) < endTimeTolerance && s.EndsAt.Sub(c.EndsAt) < endTimeTolerance {
			return true
		}
	}
	return false
}
