package streakstore

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRecordSurvivesReopen(t *testing.T) {
	dir := t.TempDir()

	s, err := Open(dir, "3jakec")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := s.Record("Krissi", "broadcast-1", time.Now()); err != nil {
		t.Fatalf("record: %v", err)
	}

	reopened, err := Open(dir, "3jakec")
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	if !reopened.Earned("krissi", "broadcast-1") {
		t.Fatal("streak recorded before the restart was not restored")
	}
}

func TestEarnedIsScopedToTheBroadcast(t *testing.T) {
	s, err := Open(t.TempDir(), "acct")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := s.Record("krissi", "broadcast-1", time.Now()); err != nil {
		t.Fatalf("record: %v", err)
	}

	if !s.Earned("krissi", "broadcast-1") {
		t.Error("same broadcast should count as earned")
	}
	// A new broadcast earns a fresh streak, so the old record must not suppress it.
	if s.Earned("krissi", "broadcast-2") {
		t.Error("a different broadcast must not be treated as earned")
	}
	if s.Earned("someone-else", "broadcast-1") {
		t.Error("a different channel must not be treated as earned")
	}
}

// Without a broadcast ID there is no way to tell this stream from the last one,
// and wrongly skipping the chase costs a streak.
func TestUnknownBroadcastIsNeverEarned(t *testing.T) {
	s, err := Open(t.TempDir(), "acct")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := s.Record("krissi", "", time.Now()); err != nil {
		t.Fatalf("record: %v", err)
	}

	if s.Len() != 0 {
		t.Error("a record without a broadcast ID should not be stored")
	}
	if s.Earned("krissi", "") {
		t.Error("an empty broadcast ID must never report earned")
	}
}

func TestExpiredRecordsAreDroppedOnReopen(t *testing.T) {
	dir := t.TempDir()

	s, err := Open(dir, "acct")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := s.Record("stale", "b-old", time.Now().Add(-DefaultRetention-time.Hour)); err != nil {
		t.Fatalf("record: %v", err)
	}
	if err := s.Record("fresh", "b-new", time.Now()); err != nil {
		t.Fatalf("record: %v", err)
	}

	reopened, err := Open(dir, "acct")
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	if reopened.Earned("stale", "b-old") {
		t.Error("a record past the retention window should be dropped")
	}
	if !reopened.Earned("fresh", "b-new") {
		t.Error("a recent record should survive")
	}
}

func TestMissingFileOpensEmpty(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "nested"), "acct")
	if err != nil {
		t.Fatalf("expected a missing file to be fine, got %v", err)
	}
	if s.Len() != 0 {
		t.Errorf("expected an empty store, got %d records", s.Len())
	}
}

// A truncated or hand-edited file must not stop the miner from starting; the
// cost of an empty store is one harvest cycle, the cost of refusing is the run.
func TestCorruptFileYieldsUsableEmptyStore(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "acct.json"), []byte("{not json"), 0o600); err != nil {
		t.Fatalf("seed: %v", err)
	}

	s, err := Open(dir, "acct")
	if err == nil {
		t.Error("expected the parse failure to be reported")
	}
	if s == nil {
		t.Fatal("expected a usable store despite the parse failure")
	}
	if err := s.Record("krissi", "b-1", time.Now()); err != nil {
		t.Fatalf("store should still be writable: %v", err)
	}
	if !s.Earned("krissi", "b-1") {
		t.Error("store should work normally after recovering from a corrupt file")
	}
}

func TestUnknownVersionIsRejectedButUsable(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "acct.json"),
		[]byte(`{"version":99,"streaks":{"krissi":{"broadcast_id":"b","earned_at":"2026-01-01T00:00:00Z"}}}`), 0o600); err != nil {
		t.Fatalf("seed: %v", err)
	}

	s, err := Open(dir, "acct")
	if err == nil {
		t.Error("expected an unsupported version to be reported")
	}
	if s.Earned("krissi", "b") {
		t.Error("records from an unsupported version must not be trusted")
	}
}

func TestAccountNameIsCaseInsensitive(t *testing.T) {
	dir := t.TempDir()

	s, err := Open(dir, "3JakeC")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := s.Record("krissi", "b-1", time.Now()); err != nil {
		t.Fatalf("record: %v", err)
	}

	reopened, err := Open(dir, "3jakec")
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	if !reopened.Earned("krissi", "b-1") {
		t.Error("the same account in different casing should share one file")
	}
}

func TestRecordLeavesNoTempFileBehind(t *testing.T) {
	dir := t.TempDir()

	s, err := Open(dir, "acct")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := s.Record("krissi", "b-1", time.Now()); err != nil {
		t.Fatalf("record: %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".tmp" {
			t.Errorf("atomic write left %s behind", e.Name())
		}
	}
}
