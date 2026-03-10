package model

import (
	"testing"
)

// --- helpers ---

func twoOutcomes() []Outcome {
	return []Outcome{
		{ID: "o1", Title: "Yes", TotalUsers: 100, TotalPoints: 5000, TopPoints: 500},
		{ID: "o2", Title: "No", TotalUsers: 50, TotalPoints: 10000, TopPoints: 1000},
	}
}

func threeOutcomes() []Outcome {
	return []Outcome{
		{ID: "o1", Title: "A", TotalUsers: 80, TotalPoints: 4000, TopPoints: 300},
		{ID: "o2", Title: "B", TotalUsers: 60, TotalPoints: 6000, TopPoints: 800},
		{ID: "o3", Title: "C", TotalUsers: 10, TotalPoints: 2000, TopPoints: 200},
	}
}

func makeBet(outcomes []Outcome, settings *BetSettings) *Bet {
	b := NewBet(outcomes, settings)
	b.UpdateOutcomes(outcomes) // recompute stats
	return b
}

// --- Bet.Calculate strategy tests ---

func TestCalculate_MostVoted(t *testing.T) {
	s := DefaultBetSettings()
	s.Strategy = StrategyMostVoted
	b := makeBet(twoOutcomes(), s)

	d := b.Calculate(10000)
	if d.Choice != 0 {
		t.Errorf("expected choice 0 (most users), got %d", d.Choice)
	}
	if d.OutcomeID != "o1" {
		t.Errorf("expected outcome o1, got %s", d.OutcomeID)
	}
}

func TestCalculate_HighOdds(t *testing.T) {
	s := DefaultBetSettings()
	s.Strategy = StrategyHighOdds
	b := makeBet(twoOutcomes(), s)

	d := b.Calculate(10000)
	// outcome 0 has higher odds (total_points / outcome_points = 15000/5000=3.0 vs 15000/10000=1.5)
	if d.Choice != 0 {
		t.Errorf("expected choice 0 (high odds), got %d", d.Choice)
	}
}

func TestCalculate_Percentage(t *testing.T) {
	s := DefaultBetSettings()
	s.Strategy = StrategyPercentage
	b := makeBet(twoOutcomes(), s)

	d := b.Calculate(10000)
	// OddsPercentage = 100/odds. o1 odds=3.0 → 33.33, o2 odds=1.5 → 66.67
	// Percentage strategy picks highest odds_percentage → o2
	if d.Choice != 1 {
		t.Errorf("expected choice 1 (highest odds_percentage), got %d", d.Choice)
	}
}

func TestCalculate_SmartMoney(t *testing.T) {
	s := DefaultBetSettings()
	s.Strategy = StrategySmartMoney
	b := makeBet(twoOutcomes(), s)

	d := b.Calculate(10000)
	// o2 has TopPoints=1000 > o1 TopPoints=500
	if d.Choice != 1 {
		t.Errorf("expected choice 1 (highest top_points), got %d", d.Choice)
	}
}

func TestCalculate_Smart_HighOddsBranch(t *testing.T) {
	// When percentage gap is small (< PercentageGap), SMART uses HIGH_ODDS
	s := DefaultBetSettings()
	s.Strategy = StrategySmart
	s.PercentageGap = 90 // very large gap threshold; difference will be below

	outcomes := []Outcome{
		{ID: "o1", Title: "Yes", TotalUsers: 50, TotalPoints: 5000, TopPoints: 500},
		{ID: "o2", Title: "No", TotalUsers: 50, TotalPoints: 10000, TopPoints: 1000},
	}
	b := makeBet(outcomes, s)

	d := b.Calculate(10000)
	// Both have 50% users so diff=0 < 90 → uses HIGH_ODDS
	// o1 odds = 15000/5000=3.0, o2 odds = 15000/10000=1.5 → pick o1
	if d.Choice != 0 {
		t.Errorf("expected choice 0 (high odds, close race), got %d", d.Choice)
	}
}

func TestCalculate_Smart_MostVotedBranch(t *testing.T) {
	// When percentage gap is large (>= PercentageGap), SMART uses MOST_VOTED
	s := DefaultBetSettings()
	s.Strategy = StrategySmart
	s.PercentageGap = 5 // very small gap threshold; difference will be above

	outcomes := []Outcome{
		{ID: "o1", Title: "Yes", TotalUsers: 90, TotalPoints: 5000, TopPoints: 500},
		{ID: "o2", Title: "No", TotalUsers: 10, TotalPoints: 10000, TopPoints: 1000},
	}
	b := makeBet(outcomes, s)

	d := b.Calculate(10000)
	// Percentage users: 90% vs 10%, diff=80 >= 5 → uses MOST_VOTED → o1
	if d.Choice != 0 {
		t.Errorf("expected choice 0 (most voted, wide gap), got %d", d.Choice)
	}
}

func TestCalculate_NumberStrategies(t *testing.T) {
	outcomes := threeOutcomes()

	tests := []struct {
		strategy Strategy
		expected int
	}{
		{StrategyNumber1, 0},
		{StrategyNumber2, 1},
		{StrategyNumber3, 2},
		// NUMBER_4 through NUMBER_8 fall back to 0 since only 3 outcomes
		{StrategyNumber4, 0},
		{StrategyNumber5, 0},
		{StrategyNumber6, 0},
		{StrategyNumber7, 0},
		{StrategyNumber8, 0},
	}

	for _, tc := range tests {
		s := DefaultBetSettings()
		s.Strategy = tc.strategy
		b := makeBet(outcomes, s)
		d := b.Calculate(10000)
		if d.Choice != tc.expected {
			t.Errorf("strategy %s: expected choice %d, got %d", tc.strategy, tc.expected, d.Choice)
		}
	}
}

// --- Amount calculation ---

func TestCalculate_AmountFormula(t *testing.T) {
	s := DefaultBetSettings()
	s.Strategy = StrategyNumber1
	s.Percentage = 10
	s.MaxPoints = 50000
	s.StealthMode = false
	b := makeBet(twoOutcomes(), s)

	d := b.Calculate(20000)
	// 20000 * 10 / 100 = 2000
	if d.Amount != 2000 {
		t.Errorf("expected amount 2000, got %d", d.Amount)
	}
}

func TestCalculate_AmountCappedByMaxPoints(t *testing.T) {
	s := DefaultBetSettings()
	s.Strategy = StrategyNumber1
	s.Percentage = 50
	s.MaxPoints = 1000
	s.StealthMode = false
	b := makeBet(twoOutcomes(), s)

	d := b.Calculate(100000)
	// 100000 * 50 / 100 = 50000, capped at 1000
	if d.Amount != 1000 {
		t.Errorf("expected amount 1000 (capped), got %d", d.Amount)
	}
}

func TestCalculate_StealthModeReducesAmount(t *testing.T) {
	s := DefaultBetSettings()
	s.Strategy = StrategyNumber1
	s.Percentage = 50
	s.MaxPoints = 50000
	s.StealthMode = true
	b := makeBet(twoOutcomes(), s)

	// balance * percentage / 100 = 10000 * 50 / 100 = 5000
	// o1 TopPoints = 500, amount (5000) >= TopPoints (500) → stealth applies
	d := b.Calculate(10000)
	// Amount should be less than TopPoints (500)
	if d.Amount >= 500 {
		t.Errorf("stealth mode: expected amount < 500 (topPoints), got %d", d.Amount)
	}
}

func TestCalculate_StealthModeNoEffectWhenBelowTop(t *testing.T) {
	s := DefaultBetSettings()
	s.Strategy = StrategyNumber1
	s.Percentage = 1
	s.MaxPoints = 50000
	s.StealthMode = true
	b := makeBet(twoOutcomes(), s)

	// balance * 1 / 100 = 100, which is < TopPoints (500)
	d := b.Calculate(10000)
	if d.Amount != 100 {
		t.Errorf("stealth mode should not apply when amount < topPoints, expected 100, got %d", d.Amount)
	}
}

func TestCalculate_StealthModeNoEffectWhenTopPointsZero(t *testing.T) {
	s := DefaultBetSettings()
	s.Strategy = StrategyNumber1
	s.Percentage = 50
	s.MaxPoints = 50000
	s.StealthMode = true

	outcomes := []Outcome{
		{ID: "o1", Title: "Yes", TotalUsers: 10, TotalPoints: 1000, TopPoints: 0},
		{ID: "o2", Title: "No", TotalUsers: 5, TotalPoints: 500, TopPoints: 0},
	}
	b := makeBet(outcomes, s)

	d := b.Calculate(10000)
	// TopPoints is 0, so stealth condition (amount >= topPoints && topPoints > 0) is false
	if d.Amount != 5000 {
		t.Errorf("expected 5000 when topPoints=0, got %d", d.Amount)
	}
}

func TestParseResultUsesClampedDecisionAmount(t *testing.T) {
	s := DefaultBetSettings()
	s.Strategy = StrategyNumber1
	s.Percentage = 50
	s.MaxPoints = 50000

	b := makeBet(twoOutcomes(), s)
	decision := b.Calculate(1000)
	decision.Amount = 400
	b.Decision = decision

	ep := &EventPrediction{Bet: b}
	points := ep.ParseResult("LOSE", 0)

	if points["placed"] != 400 {
		t.Fatalf("expected placed amount 400, got %d", points["placed"])
	}
	if points["gained"] != -400 {
		t.Fatalf("expected gained amount -400, got %d", points["gained"])
	}
}

// --- Edge cases ---

func TestCalculate_AllOutcomesZeroUsers(t *testing.T) {
	s := DefaultBetSettings()
	s.Strategy = StrategyMostVoted

	outcomes := []Outcome{
		{ID: "o1", Title: "Yes", TotalUsers: 0, TotalPoints: 0, TopPoints: 0},
		{ID: "o2", Title: "No", TotalUsers: 0, TotalPoints: 0, TopPoints: 0},
	}
	b := makeBet(outcomes, s)

	d := b.Calculate(10000)
	// All equal at 0 → returnChoice picks index 0
	if d.Choice != 0 {
		t.Errorf("expected choice 0 when all zero, got %d", d.Choice)
	}
}

func TestCalculate_EqualOdds(t *testing.T) {
	s := DefaultBetSettings()
	s.Strategy = StrategyHighOdds

	outcomes := []Outcome{
		{ID: "o1", Title: "Yes", TotalUsers: 50, TotalPoints: 5000, TopPoints: 100},
		{ID: "o2", Title: "No", TotalUsers: 50, TotalPoints: 5000, TopPoints: 100},
	}
	b := makeBet(outcomes, s)

	d := b.Calculate(10000)
	// Both identical odds, returnChoice picks 0 (first remains largest)
	if d.Choice != 0 {
		t.Errorf("expected choice 0 when equal odds, got %d", d.Choice)
	}
}

func TestCalculate_SmartWithSingleOutcome(t *testing.T) {
	s := DefaultBetSettings()
	s.Strategy = StrategySmart

	outcomes := []Outcome{
		{ID: "o1", Title: "Only", TotalUsers: 100, TotalPoints: 5000, TopPoints: 500},
	}
	b := makeBet(outcomes, s)

	d := b.Calculate(10000)
	// len(outcomes) < 2, so SMART does nothing, choice remains -1
	if d.Choice != -1 {
		t.Errorf("expected choice -1 for SMART with 1 outcome, got %d", d.Choice)
	}
	if d.Amount != 0 {
		t.Errorf("expected amount 0 for invalid choice, got %d", d.Amount)
	}
}

func TestCalculate_ZeroBalance(t *testing.T) {
	s := DefaultBetSettings()
	s.Strategy = StrategyNumber1
	s.StealthMode = false
	b := makeBet(twoOutcomes(), s)

	d := b.Calculate(0)
	if d.Amount != 0 {
		t.Errorf("expected 0 amount for 0 balance, got %d", d.Amount)
	}
}

// --- UpdateOutcomes ---

func TestUpdateOutcomes_ComputesStats(t *testing.T) {
	b := NewBet(twoOutcomes(), DefaultBetSettings())
	b.UpdateOutcomes(twoOutcomes())

	if b.TotalUsers != 150 {
		t.Errorf("expected total_users 150, got %d", b.TotalUsers)
	}
	if b.TotalPoints != 15000 {
		t.Errorf("expected total_points 15000, got %d", b.TotalPoints)
	}
	// o1: percentage_users = 100/150*100 = 66.67
	if b.Outcomes[0].PercentageUsers < 66.0 || b.Outcomes[0].PercentageUsers > 67.0 {
		t.Errorf("expected o1 percentage_users ~66.67, got %.2f", b.Outcomes[0].PercentageUsers)
	}
	// o1: odds = 15000/5000 = 3.0
	if b.Outcomes[0].Odds != 3.0 {
		t.Errorf("expected o1 odds 3.0, got %.2f", b.Outcomes[0].Odds)
	}
	// o1: odds_percentage = 100/3.0 = 33.33
	if b.Outcomes[0].OddsPercentage < 33.0 || b.Outcomes[0].OddsPercentage > 34.0 {
		t.Errorf("expected o1 odds_percentage ~33.33, got %.2f", b.Outcomes[0].OddsPercentage)
	}
}

func TestUpdateOutcomes_ZeroPointsOutcome(t *testing.T) {
	outcomes := []Outcome{
		{ID: "o1", Title: "Yes", TotalUsers: 10, TotalPoints: 1000, TopPoints: 100},
		{ID: "o2", Title: "No", TotalUsers: 5, TotalPoints: 0, TopPoints: 0},
	}
	b := NewBet(outcomes, DefaultBetSettings())
	b.UpdateOutcomes(outcomes)

	// o2 has 0 total_points → odds=0, odds_percentage=0
	if b.Outcomes[1].Odds != 0 {
		t.Errorf("expected o2 odds 0 for zero points, got %.2f", b.Outcomes[1].Odds)
	}
	if b.Outcomes[1].OddsPercentage != 0 {
		t.Errorf("expected o2 odds_percentage 0 for zero points, got %.2f", b.Outcomes[1].OddsPercentage)
	}
}

// --- FilterCondition / Skip ---

func TestSkip_NilFilterCondition(t *testing.T) {
	s := DefaultBetSettings()
	s.FilterCondition = nil
	b := makeBet(twoOutcomes(), s)
	b.Calculate(10000)

	skip, _ := b.Skip()
	if skip {
		t.Error("expected no skip when filter condition is nil")
	}
}

func TestSkip_GT_Pass(t *testing.T) {
	s := DefaultBetSettings()
	s.Strategy = StrategyNumber1
	s.FilterCondition = &FilterCondition{
		By:    OutcomeKeyTotalUsers,
		Where: ConditionGT,
		Value: 100, // total users = 150, 150 > 100 → pass
	}
	b := makeBet(twoOutcomes(), s)
	b.Calculate(10000)

	skip, val := b.Skip()
	if skip {
		t.Errorf("expected no skip (150 > 100), got skip with val=%.2f", val)
	}
}

func TestSkip_GT_Skip(t *testing.T) {
	s := DefaultBetSettings()
	s.Strategy = StrategyNumber1
	s.FilterCondition = &FilterCondition{
		By:    OutcomeKeyTotalUsers,
		Where: ConditionGT,
		Value: 200, // total users = 150, 150 !> 200 → skip
	}
	b := makeBet(twoOutcomes(), s)
	b.Calculate(10000)

	skip, _ := b.Skip()
	if !skip {
		t.Error("expected skip (150 is not > 200)")
	}
}

func TestSkip_LT_Pass(t *testing.T) {
	s := DefaultBetSettings()
	s.Strategy = StrategyNumber1
	s.FilterCondition = &FilterCondition{
		By:    OutcomeKeyTotalUsers,
		Where: ConditionLT,
		Value: 200, // 150 < 200 → pass
	}
	b := makeBet(twoOutcomes(), s)
	b.Calculate(10000)

	skip, _ := b.Skip()
	if skip {
		t.Error("expected no skip (150 < 200)")
	}
}

func TestSkip_LT_Skip(t *testing.T) {
	s := DefaultBetSettings()
	s.Strategy = StrategyNumber1
	s.FilterCondition = &FilterCondition{
		By:    OutcomeKeyTotalUsers,
		Where: ConditionLT,
		Value: 100, // 150 !< 100 → skip
	}
	b := makeBet(twoOutcomes(), s)
	b.Calculate(10000)

	skip, _ := b.Skip()
	if !skip {
		t.Error("expected skip (150 is not < 100)")
	}
}

func TestSkip_GTE_Boundary(t *testing.T) {
	s := DefaultBetSettings()
	s.Strategy = StrategyNumber1
	s.FilterCondition = &FilterCondition{
		By:    OutcomeKeyTotalUsers,
		Where: ConditionGTE,
		Value: 150, // 150 >= 150 → pass
	}
	b := makeBet(twoOutcomes(), s)
	b.Calculate(10000)

	skip, _ := b.Skip()
	if skip {
		t.Error("expected no skip (150 >= 150)")
	}
}

func TestSkip_LTE_Boundary(t *testing.T) {
	s := DefaultBetSettings()
	s.Strategy = StrategyNumber1
	s.FilterCondition = &FilterCondition{
		By:    OutcomeKeyTotalUsers,
		Where: ConditionLTE,
		Value: 150, // 150 <= 150 → pass
	}
	b := makeBet(twoOutcomes(), s)
	b.Calculate(10000)

	skip, _ := b.Skip()
	if skip {
		t.Error("expected no skip (150 <= 150)")
	}
}

func TestSkip_DecisionUsers_UsesChosenOutcome(t *testing.T) {
	s := DefaultBetSettings()
	s.Strategy = StrategyNumber2 // picks outcome index 1
	s.FilterCondition = &FilterCondition{
		By:    OutcomeKeyDecisionUsers,
		Where: ConditionGT,
		Value: 40, // o2 has 50 users → 50 > 40 → pass
	}
	b := makeBet(twoOutcomes(), s)
	b.Calculate(10000)

	skip, val := b.Skip()
	if skip {
		t.Errorf("expected no skip (o2 users=50 > 40), got skip with val=%.2f", val)
	}
	if val != 50 {
		t.Errorf("expected compared value 50, got %.2f", val)
	}
}

func TestSkip_DecisionPoints_UsesChosenOutcome(t *testing.T) {
	s := DefaultBetSettings()
	s.Strategy = StrategyNumber1 // picks outcome index 0
	s.FilterCondition = &FilterCondition{
		By:    OutcomeKeyDecisionPoints,
		Where: ConditionGT,
		Value: 6000, // o1 has 5000 points → 5000 !> 6000 → skip
	}
	b := makeBet(twoOutcomes(), s)
	b.Calculate(10000)

	skip, _ := b.Skip()
	if !skip {
		t.Error("expected skip (o1 points=5000 is not > 6000)")
	}
}

func TestSkip_TotalPoints_SumsAllOutcomes(t *testing.T) {
	s := DefaultBetSettings()
	s.Strategy = StrategyNumber1
	s.FilterCondition = &FilterCondition{
		By:    OutcomeKeyTotalPoints,
		Where: ConditionGT,
		Value: 14000, // total = 15000 > 14000 → pass
	}
	b := makeBet(twoOutcomes(), s)
	b.Calculate(10000)

	skip, val := b.Skip()
	if skip {
		t.Errorf("expected no skip (total_points=15000 > 14000), got skip with val=%.2f", val)
	}
	if val != 15000 {
		t.Errorf("expected compared value 15000, got %.2f", val)
	}
}

func TestSkip_OddsPercentage_UsesChosenOutcome(t *testing.T) {
	s := DefaultBetSettings()
	s.Strategy = StrategyNumber1 // outcome 0
	s.FilterCondition = &FilterCondition{
		By:    OutcomeKeyOddsPercentage,
		Where: ConditionGT,
		Value: 30, // o1 odds_percentage ~33.33 > 30 → pass
	}
	b := makeBet(twoOutcomes(), s)
	b.Calculate(10000)

	skip, _ := b.Skip()
	if skip {
		t.Error("expected no skip (o1 odds_percentage ~33.33 > 30)")
	}
}

// --- ParseStrategy ---

func TestParseStrategy(t *testing.T) {
	tests := []struct {
		input    string
		expected Strategy
	}{
		{"MOST_VOTED", StrategyMostVoted},
		{"HIGH_ODDS", StrategyHighOdds},
		{"PERCENTAGE", StrategyPercentage},
		{"SMART_MONEY", StrategySmartMoney},
		{"SMART", StrategySmart},
		{"NUMBER_1", StrategyNumber1},
		{"NUMBER_8", StrategyNumber8},
		{"UNKNOWN", StrategySmart}, // default
		{"", StrategySmart},        // default
	}
	for _, tc := range tests {
		got := ParseStrategy(tc.input)
		if got != tc.expected {
			t.Errorf("ParseStrategy(%q) = %v, want %v", tc.input, got, tc.expected)
		}
	}
}

// --- ParseCondition ---

func TestParseCondition(t *testing.T) {
	tests := []struct {
		input    string
		expected Condition
	}{
		{"GT", ConditionGT},
		{"LT", ConditionLT},
		{"GTE", ConditionGTE},
		{"LTE", ConditionLTE},
		{"UNKNOWN", ConditionGT},
	}
	for _, tc := range tests {
		got := ParseCondition(tc.input)
		if got != tc.expected {
			t.Errorf("ParseCondition(%q) = %v, want %v", tc.input, got, tc.expected)
		}
	}
}

// --- GetPredictionWindow ---

func TestGetPredictionWindow(t *testing.T) {
	tests := []struct {
		name     string
		mode     DelayMode
		delay    float64
		window   float64
		expected float64
	}{
		{"FromStart_Normal", DelayModeFromStart, 10, 30, 10},
		{"FromStart_Exceeds", DelayModeFromStart, 50, 30, 30},
		{"FromEnd_Normal", DelayModeFromEnd, 6, 30, 24},
		{"FromEnd_Large", DelayModeFromEnd, 50, 30, 0},
		{"Percentage", DelayModePercentage, 0.5, 30, 15},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := DefaultBetSettings()
			s.DelayMode = tc.mode
			s.Delay = tc.delay
			got := GetPredictionWindow(s, tc.window)
			if got != tc.expected {
				t.Errorf("GetPredictionWindow: expected %.2f, got %.2f", tc.expected, got)
			}
		})
	}
}

// --- Strategy.String ---

func TestStrategy_String(t *testing.T) {
	if StrategyMostVoted.String() != "MOST_VOTED" {
		t.Errorf("expected MOST_VOTED, got %s", StrategyMostVoted.String())
	}
	if StrategySmart.String() != "SMART" {
		t.Errorf("expected SMART, got %s", StrategySmart.String())
	}
	// Out of range defaults to SMART
	if Strategy(99).String() != "SMART" {
		t.Errorf("expected SMART for out-of-range, got %s", Strategy(99).String())
	}
}

// --- ParseResult ---

func TestParseResult_Win(t *testing.T) {
	s := DefaultBetSettings()
	s.Strategy = StrategyNumber1
	b := makeBet(twoOutcomes(), s)
	b.Calculate(10000)

	ep := &EventPrediction{Bet: b}
	points := ep.ParseResult("WIN", 2000)

	if points["placed"] != b.Decision.Amount {
		t.Errorf("expected placed=%d, got %d", b.Decision.Amount, points["placed"])
	}
	if points["won"] != 2000 {
		t.Errorf("expected won=2000, got %d", points["won"])
	}
	if ep.Result.Type != "WIN" {
		t.Errorf("expected result type WIN, got %s", ep.Result.Type)
	}
}

func TestParseResult_Lose(t *testing.T) {
	s := DefaultBetSettings()
	s.Strategy = StrategyNumber1
	b := makeBet(twoOutcomes(), s)
	b.Calculate(10000)

	ep := &EventPrediction{Bet: b}
	points := ep.ParseResult("LOSE", 0)

	if points["won"] != 0 {
		// pointsWon=0 and not REFUND, so "won" key should not be present
		// Actually re-reading the code: won is set only if pointsWon > 0 or REFUND
		if _, ok := points["won"]; ok {
			t.Errorf("expected no 'won' key for LOSE with 0 points, got %d", points["won"])
		}
	}
	if ep.Result.Type != "LOSE" {
		t.Errorf("expected result type LOSE, got %s", ep.Result.Type)
	}
}

func TestParseResult_Refund(t *testing.T) {
	s := DefaultBetSettings()
	s.Strategy = StrategyNumber1
	b := makeBet(twoOutcomes(), s)
	b.Calculate(10000)

	ep := &EventPrediction{Bet: b}
	points := ep.ParseResult("REFUND", 500)

	// REFUND: no "placed" key
	if _, ok := points["placed"]; ok {
		t.Error("expected no 'placed' key for REFUND")
	}
	if points["won"] != 500 {
		t.Errorf("expected won=500, got %d", points["won"])
	}
	if ep.Result.Type != "REFUND" {
		t.Errorf("expected result type REFUND, got %s", ep.Result.Type)
	}
}
