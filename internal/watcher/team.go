package watcher

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/Guliveer/twitch-miner-go/internal/config"
	"github.com/Guliveer/twitch-miner-go/internal/constants"
	"github.com/Guliveer/twitch-miner-go/internal/gql"
	"github.com/Guliveer/twitch-miner-go/internal/logger"
	"github.com/Guliveer/twitch-miner-go/internal/model"
)

// TeamWatcher polls Twitch GQL for live members of configured teams
// and adds/removes streamers dynamically. Similar to CategoryWatcher,
// it tracks one live streamer per team.
type TeamWatcher struct {
	mu sync.Mutex

	gqlClient        *gql.Client
	log              *logger.Logger
	teams            []string
	pollInterval     time.Duration
	blacklist        map[string]bool
	streamerDefaults *model.StreamerSettings

	// teamStreamers maps team name → currently tracked username (empty if none).
	teamStreamers map[string]string
}

// NewTeamWatcher creates a new TeamWatcher from configuration.
func NewTeamWatcher(
	cfg config.TeamWatcherConfig,
	gqlClient *gql.Client,
	log *logger.Logger,
	blacklist []string,
	streamerDefaults *model.StreamerSettings,
) *TeamWatcher {
	teams := make([]string, 0, len(cfg.Teams))
	for _, t := range cfg.Teams {
		teams = append(teams, t.Name)
	}

	blacklistMap := make(map[string]bool, len(blacklist))
	for _, name := range blacklist {
		blacklistMap[strings.ToLower(name)] = true
	}

	interval := cfg.PollInterval
	if interval <= 0 {
		interval = constants.DefaultCategoryWatcherInterval
	}

	teamStreamers := make(map[string]string, len(teams))
	for _, t := range teams {
		teamStreamers[t] = ""
	}

	return &TeamWatcher{
		gqlClient:        gqlClient,
		log:              log,
		teams:            teams,
		pollInterval:     interval,
		blacklist:        blacklistMap,
		streamerDefaults: streamerDefaults,
		teamStreamers:    teamStreamers,
	}
}

// Run starts the team watcher loop. It blocks until the context is cancelled.
func (tw *TeamWatcher) Run(
	ctx context.Context,
	addStreamer func(context.Context, *model.Streamer),
	removeStreamer func(string, string),
	getTrackedStreamers func() []*model.Streamer,
) error {
	tw.log.Info("👥 TeamWatcher started",
		"teams", len(tw.teams),
		"poll_interval", tw.pollInterval,
	)

	tw.evaluate(ctx, addStreamer, removeStreamer, getTrackedStreamers)

	ticker := time.NewTicker(tw.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			tw.log.Info("👥 TeamWatcher stopping")
			tw.mu.Lock()
			for team, username := range tw.teamStreamers {
				if username != "" {
					removeStreamer(username, "team_watcher_shutdown")
					tw.teamStreamers[team] = ""
				}
			}
			tw.mu.Unlock()
			return ctx.Err()
		case <-ticker.C:
			tw.evaluate(ctx, addStreamer, removeStreamer, getTrackedStreamers)
		}
	}
}

func (tw *TeamWatcher) evaluate(
	ctx context.Context,
	addStreamer func(context.Context, *model.Streamer),
	removeStreamer func(string, string),
	getTrackedStreamers func() []*model.Streamer,
) {
	for _, teamName := range tw.teams {
		if ctx.Err() != nil {
			return
		}

		trackedStreamers := getTrackedStreamers()

		tw.mu.Lock()
		currentUsername := tw.teamStreamers[teamName]
		tw.mu.Unlock()

		// Check if the current streamer is still valid (online and still a team member).
		if currentUsername != "" {
			stillOnline := false
			for _, s := range trackedStreamers {
				s.Mu.RLock()
				if strings.EqualFold(s.Username, currentUsername) && s.IsOnline {
					stillOnline = true
				}
				s.Mu.RUnlock()
				if stillOnline {
					break
				}
			}
			if stillOnline {
				continue // Still online — keep them.
			}
			removeStreamer(currentUsername, "team_member_went_offline")
			tw.mu.Lock()
			tw.teamStreamers[teamName] = ""
			tw.mu.Unlock()
		}

		// Check if any manually-tracked streamer is already a live team member.
		members, err := tw.gqlClient.GetTeamMembers(ctx, teamName)
		if err != nil {
			tw.log.Warn("Failed to fetch team members",
				"team", teamName,
				"error", err,
			)
			continue
		}

		if len(members) == 0 {
			tw.log.Info("No members found for team", "team", teamName)
			continue
		}

		// Build set of already-tracked channel IDs.
		existingIDs := make(map[string]bool, len(trackedStreamers))
		existingLogins := make(map[string]bool, len(trackedStreamers))
		for _, s := range trackedStreamers {
			s.Mu.RLock()
			existingIDs[s.ChannelID] = true
			existingLogins[strings.ToLower(s.Username)] = true
			s.Mu.RUnlock()
		}

		// Check if team is already covered by a manually-listed, live streamer.
		covered := false
		for _, m := range members {
			if m.IsLive && existingLogins[strings.ToLower(m.Login)] {
				covered = true
				break
			}
		}
		if covered {
			continue
		}

		// Find the best live candidate (highest viewers, not already tracked, not blacklisted).
		var best *gql.TeamMember
		for i := range members {
			m := &members[i]
			if !m.IsLive {
				continue
			}
			if existingIDs[m.UserID] || existingLogins[strings.ToLower(m.Login)] {
				continue
			}
			if tw.blacklist[strings.ToLower(m.Login)] {
				continue
			}
			if best == nil || m.ViewersCount > best.ViewersCount {
				best = m
			}
		}

		if best == nil {
			liveCount := 0
			for _, m := range members {
				if m.IsLive {
					liveCount++
				}
			}
			if liveCount > 0 {
				tw.log.Info("All live team members are already tracked",
					"team", teamName,
					"live_members", liveCount,
				)
			} else {
				tw.log.Info("No live members in team", "team", teamName)
			}
			continue
		}

		streamer := model.NewStreamer(best.Login)
		streamer.ChannelID = best.UserID
		streamer.DisplayName = best.DisplayName
		streamer.IsTeamWatched = true
		streamer.TeamName = teamName

		streamer.IsOnline = true
		streamer.OnlineAt = time.Now()
		stream := model.NewStream()
		if best.GameID != "" {
			stream.Game = &model.GameInfo{
				ID:   best.GameID,
				Slug: best.GameSlug,
				Name: best.GameName,
			}
		}
		stream.ViewersCount = best.ViewersCount
		streamer.Stream = stream

		defaults := *tw.streamerDefaults
		if defaults.Bet != nil {
			betCopy := *defaults.Bet
			if betCopy.FilterCondition != nil {
				fcCopy := *betCopy.FilterCondition
				betCopy.FilterCondition = &fcCopy
			}
			defaults.Bet = &betCopy
		}
		defaults.FollowRaid = false
		streamer.Settings = &defaults

		tw.mu.Lock()
		tw.teamStreamers[teamName] = best.Login
		tw.mu.Unlock()

		addStreamer(ctx, streamer)

		tw.log.Info("🔍 Discovered via team",
			"streamer", best.Login,
			"team", teamName,
			"viewers", best.ViewersCount,
		)
	}
}
