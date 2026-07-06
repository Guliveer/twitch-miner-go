package model

import (
	"testing"
	"time"
)

func TestDropUpdate_IsPrintable_FirstMinute(t *testing.T) {
	t.Parallel()

	drop := NewDrop("d1", "Test Drop", nil, 60, refTime(t, "00:00"), refTime(t, "23:59"))
	drop.CurrentMinutesWatched = 0
	drop.PercentageProgress = 0

	drop.Update(true, 1, "", false)

	if !drop.IsPrintable {
		t.Errorf("expected IsPrintable=true for first minute (0→1), got false")
	}
	if drop.CurrentMinutesWatched != 1 {
		t.Errorf("expected CurrentMinutesWatched=1, got %d", drop.CurrentMinutesWatched)
	}
}

func TestDropUpdate_IsPrintable_FirstMinuteAt25(t *testing.T) {
	t.Parallel()

	drop := NewDrop("d1", "Test Drop", nil, 4, refTime(t, "00:00"), refTime(t, "23:59"))
	drop.CurrentMinutesWatched = 0
	drop.PercentageProgress = 0

	drop.Update(true, 1, "", false)

	// 0→1 min, 4 required → 25%. First minute condition should still fire.
	if !drop.IsPrintable {
		t.Errorf("expected IsPrintable=true for first minute even when it hits 25%%, got false")
	}
}

func TestDropUpdate_IsPrintable_QuarterBoundary(t *testing.T) {
	t.Parallel()

	drop := NewDrop("d1", "Test Drop", nil, 60, refTime(t, "00:00"), refTime(t, "23:59"))
	drop.CurrentMinutesWatched = 14
	drop.PercentageProgress = 23

	drop.Update(true, 15, "", false)

	// 14→15 min, 60 required → 23%→25%. Quarter boundary at 25%.
	if !drop.IsPrintable {
		t.Errorf("expected IsPrintable=true for quarter boundary (14→15, 25%%), got false")
	}
}

func TestDropUpdate_IsPrintable_QuarterBoundary50(t *testing.T) {
	t.Parallel()

	drop := NewDrop("d1", "Test Drop", nil, 60, refTime(t, "00:00"), refTime(t, "23:59"))
	drop.CurrentMinutesWatched = 29
	drop.PercentageProgress = 48

	drop.Update(true, 30, "", false)

	// 29→30 min, 60 required → 48%→50%. Quarter boundary at 50%.
	if !drop.IsPrintable {
		t.Errorf("expected IsPrintable=true for quarter boundary (29→30, 50%%), got false")
	}
}

func TestDropUpdate_IsPrintable_QuarterBoundary75(t *testing.T) {
	t.Parallel()

	drop := NewDrop("d1", "Test Drop", nil, 60, refTime(t, "00:00"), refTime(t, "23:59"))
	drop.CurrentMinutesWatched = 44
	drop.PercentageProgress = 73

	drop.Update(true, 45, "", false)

	// 44→45 min → 73%→75%. Quarter boundary.
	if !drop.IsPrintable {
		t.Errorf("expected IsPrintable=true for quarter boundary (44→45, 75%%), got false")
	}
}

func TestDropUpdate_IsPrintable_QuarterBoundary100(t *testing.T) {
	t.Parallel()

	drop := NewDrop("d1", "Test Drop", nil, 60, refTime(t, "00:00"), refTime(t, "23:59"))
	drop.CurrentMinutesWatched = 59
	drop.PercentageProgress = 98

	drop.Update(true, 60, "", false)

	// 59→60 min → 98%→100%. Quarter boundary.
	if !drop.IsPrintable {
		t.Errorf("expected IsPrintable=true for quarter boundary (59→60, 100%%), got false")
	}
}

func TestDropUpdate_IsPrintable_SameQuarter(t *testing.T) {
	t.Parallel()

	drop := NewDrop("d1", "Test Drop", nil, 60, refTime(t, "00:00"), refTime(t, "23:59"))
	drop.CurrentMinutesWatched = 5
	drop.PercentageProgress = 8

	drop.Update(true, 10, "", false)

	// 5→10 min, 60 required → 8%→16%. No quarter boundary.
	if drop.IsPrintable {
		t.Errorf("expected IsPrintable=false within same quarter (5→10, 8%%→16%%), got true")
	}
}

func TestDropUpdate_IsPrintable_NoProgress(t *testing.T) {
	t.Parallel()

	drop := NewDrop("d1", "Test Drop", nil, 60, refTime(t, "00:00"), refTime(t, "23:59"))
	drop.CurrentMinutesWatched = 15
	drop.PercentageProgress = 25

	drop.Update(true, 15, "", false)

	// Same value → no progress.
	if drop.IsPrintable {
		t.Errorf("expected IsPrintable=false when no progress (15→15), got true")
	}
}

func TestDropUpdate_IsPrintable_Regress(t *testing.T) {
	t.Parallel()

	drop := NewDrop("d1", "Test Drop", nil, 60, refTime(t, "00:00"), refTime(t, "23:59"))
	drop.CurrentMinutesWatched = 15
	drop.PercentageProgress = 25

	drop.Update(true, 10, "", false)

	// Going backwards → not printable (currentMinutesWatched > old check fails).
	if drop.IsPrintable {
		t.Errorf("expected IsPrintable=false on regress (15→10), got true")
	}
}

func TestDropUpdate_IsPrintable_HalfMinuteRequirement(t *testing.T) {
	t.Parallel()

	// Drop requiring only 2 minutes: 0→1 is first minute (printable),
	// 1→2 is 50%→100% which is a quarter boundary (printable).
	// Test the second transition specifically.
	drop := NewDrop("d1", "Test Drop", nil, 2, refTime(t, "00:00"), refTime(t, "23:59"))
	drop.CurrentMinutesWatched = 1
	drop.PercentageProgress = 50

	drop.Update(true, 2, "", false)

	// 1→2 min → 50%→100%. Quarter boundary at 100%.
	if !drop.IsPrintable {
		t.Errorf("expected IsPrintable=true for 1→2 with 2-min requirement (50%%→100%%), got false")
	}
}

func TestDropUpdate_IsClaimable_WithInstanceID(t *testing.T) {
	t.Parallel()

	drop := NewDrop("d1", "Test Drop", nil, 60, refTime(t, "00:00"), refTime(t, "23:59"))
	drop.Update(true, 60, "inst-123", false)

	if !drop.IsClaimable {
		t.Errorf("expected IsClaimable=true with instance ID and not claimed, got false")
	}
}

func TestDropUpdate_IsClaimable_EmptyInstanceID(t *testing.T) {
	t.Parallel()

	drop := NewDrop("d1", "Test Drop", nil, 60, refTime(t, "00:00"), refTime(t, "23:59"))
	drop.Update(true, 60, "", false)

	if drop.IsClaimable {
		t.Errorf("expected IsClaimable=false with empty instance ID, got true")
	}
}

func TestDropUpdate_IsClaimable_Claimed(t *testing.T) {
	t.Parallel()

	drop := NewDrop("d1", "Test Drop", nil, 60, refTime(t, "00:00"), refTime(t, "23:59"))
	drop.Update(true, 60, "inst-123", true)

	if drop.IsClaimable {
		t.Errorf("expected IsClaimable=false when claimed, got true")
	}
}

func TestDropUpdate_PercentageProgress(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		watched  int
		required int
		want     int
	}{
		{"zero progress", 0, 60, 0},
		{"one minute", 1, 60, 1},
		{"quarter", 15, 60, 25},
		{"half", 30, 60, 50},
		{"three quarters", 45, 60, 75},
		{"complete", 60, 60, 100},
		{"over achievement", 120, 60, 200}, // utils.Percentage doesn't cap at 100
		{"rounds down", 1, 60, 1},
		{"exact percentage", 10, 100, 10},
		{"small requirement", 1, 1, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			drop := NewDrop("d1", "Test", nil, tt.required, refTime(t, "00:00"), refTime(t, "23:59"))
			drop.Update(true, tt.watched, "", false)

			if drop.PercentageProgress != tt.want {
				t.Errorf("Percentage(%d/%d) = %d, want %d",
					tt.watched, tt.required, drop.PercentageProgress, tt.want)
			}
		})
	}
}

func TestDropUpdate_FieldUpdate_HasPreconditionsMet(t *testing.T) {
	t.Parallel()

	drop := NewDrop("d1", "Test", nil, 60, refTime(t, "00:00"), refTime(t, "23:59"))
	if drop.HasPreconditionsMet != nil {
		t.Fatal("expected HasPreconditionsMet to be nil initially")
	}

	drop.Update(true, 0, "", false)
	if drop.HasPreconditionsMet == nil || !*drop.HasPreconditionsMet {
		t.Errorf("expected HasPreconditionsMet=true, got %v", drop.HasPreconditionsMet)
	}

	drop.Update(false, 0, "", false)
	if drop.HasPreconditionsMet == nil || *drop.HasPreconditionsMet {
		t.Errorf("expected HasPreconditionsMet=false, got %v", *drop.HasPreconditionsMet)
	}
}

func TestDropUpdate_FieldUpdate_DropInstanceID(t *testing.T) {
	t.Parallel()

	drop := NewDrop("d1", "Test", nil, 60, refTime(t, "00:00"), refTime(t, "23:59"))
	drop.Update(true, 0, "", false)

	if drop.DropInstanceID != "" {
		t.Errorf("expected empty DropInstanceID, got %q", drop.DropInstanceID)
	}

	drop.Update(true, 0, "inst-456", false)
	if drop.DropInstanceID != "inst-456" {
		t.Errorf("expected DropInstanceID=inst-456, got %q", drop.DropInstanceID)
	}
}

func TestDropUpdate_FieldUpdate_IsClaimed(t *testing.T) {
	t.Parallel()

	drop := NewDrop("d1", "Test", nil, 60, refTime(t, "00:00"), refTime(t, "23:59"))
	drop.Update(true, 0, "", false)

	if drop.IsClaimed {
		t.Errorf("expected IsClaimed=false initially, got true")
	}

	drop.Update(true, 0, "", true)
	if !drop.IsClaimed {
		t.Errorf("expected IsClaimed=true after update, got false")
	}
}

// refTime is a test helper that returns a time.Time from an HH:MM string
// on a fixed date to avoid test pollution from time.Now().
func refTime(t *testing.T, hhmm string) time.Time {
	t.Helper()
	// Use a fixed date in the past so time windows work.
	v, err := time.Parse("2006-01-02 15:04", "2025-01-01 "+hhmm)
	if err != nil {
		t.Fatalf("bad time %q: %v", hhmm, err)
	}
	return v
}
