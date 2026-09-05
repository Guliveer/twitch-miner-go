package logger

import (
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"
)

func TestLogFileNameEncodesFullStartupTimestamp(t *testing.T) {
	dir := t.TempDir()

	log, err := Setup(Config{Level: slog.LevelInfo, LogDir: dir})
	if err != nil {
		t.Fatalf("Setup: %v", err)
	}
	defer log.Close()

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("reading log dir: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected exactly 1 log file, got %d", len(entries))
	}

	name := strings.TrimSuffix(entries[0].Name(), ".log")
	// A broken layout silently degrades to a near-constant name, so every start
	// would append to the same file forever.
	if _, err := time.Parse("2006-01-02_15-04-05", name); err != nil {
		t.Fatalf("log filename %q is not a parseable timestamp: %v", name, err)
	}
}