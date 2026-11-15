package model

import "time"

// RewardDeadlineSafetyMargin is the minimum slack allowed for urgent selection.
const RewardDeadlineSafetyMargin = 5 * time.Minute

// RewardDeadline describes the time still needed for an eligible reward.
type RewardDeadline struct {
	EndAt     time.Time
	Remaining time.Duration
}

// LatestStart is when uninterrupted credited viewing would need to begin.
func (r RewardDeadline) LatestStart() time.Time {
	return r.EndAt.Add(-r.Remaining)
}
