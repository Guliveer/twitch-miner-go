package model

import "testing"

func TestNewBetComputesDerivedOutcomeValues(t *testing.T) {
	t.Parallel()

	bet := NewBet([]Outcome{
		{ID: "a", TotalUsers: 40, TotalPoints: 4000, TopPoints: 1000},
		{ID: "b", TotalUsers: 60, TotalPoints: 6000, TopPoints: 2000},
	}, DefaultBetSettings())

	if bet.TotalUsers != 100 {
		t.Fatalf("expected total users to be computed, got %d", bet.TotalUsers)
	}
	if bet.TotalPoints != 10000 {
		t.Fatalf("expected total points to be computed, got %d", bet.TotalPoints)
	}
	if bet.Outcomes[0].PercentageUsers == 0 {
		t.Fatal("expected percentage users to be computed")
	}
	if bet.Outcomes[0].Odds == 0 {
		t.Fatal("expected odds to be computed")
	}
}
