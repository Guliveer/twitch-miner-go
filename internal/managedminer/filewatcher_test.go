package managedminer

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Guliveer/twitch-miner-go/internal/config"
)

const minimalYAML = `max_watch_streams: 1
streamers:
  - username: some_streamer
`

func writeYAML(t *testing.T, dir, username, content string) string {
	t.Helper()
	path := filepath.Join(dir, username+".yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func newTestFileWatcher(dir string, mgr fileManager) *FileWatcher {
	return &FileWatcher{
		dir:       dir,
		manager:   mgr,
		interval:  time.Minute,
		log:       discardLogger(),
		watermark: make(map[string]int64),
	}
}

// ── tests ─────────────────────────────────────────────────────────────────────

func TestFileWatcher_NewFileStartsMiner(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, dir, "alice", minimalYAML)

	mgr := &fakeMgr{}
	w := newTestFileWatcher(dir, mgr)
	w.sync()

	if len(mgr.started) != 1 || mgr.started[0] != "alice" {
		t.Errorf("expected alice to be started, got %v", mgr.started)
	}
}

func TestFileWatcher_ChangedMtimeRestartsMiner(t *testing.T) {
	dir := t.TempDir()
	path := writeYAML(t, dir, "alice", minimalYAML)

	mgr := &fakeMgr{}
	w := newTestFileWatcher(dir, mgr)
	w.sync()
	mgr.started = nil

	// bump mtime
	past := time.Now().Add(-time.Second)
	future := time.Now().Add(time.Second)
	_ = past
	if err := os.Chtimes(path, future, future); err != nil {
		t.Fatal(err)
	}
	w.sync()

	if len(mgr.restarted) != 1 || mgr.restarted[0] != "alice" {
		t.Errorf("expected alice to be restarted, got %v", mgr.restarted)
	}
}

func TestFileWatcher_SameMtimeNoRestart(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, dir, "alice", minimalYAML)

	mgr := &fakeMgr{}
	w := newTestFileWatcher(dir, mgr)
	w.sync()
	w.sync()

	if len(mgr.restarted) != 0 {
		t.Errorf("expected no restarts, got %v", mgr.restarted)
	}
}

func TestFileWatcher_RemovedFileStopsMiner(t *testing.T) {
	dir := t.TempDir()
	path := writeYAML(t, dir, "alice", minimalYAML)

	mgr := &fakeMgr{}
	w := newTestFileWatcher(dir, mgr)
	w.sync()

	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	w.sync()

	if len(mgr.stopped) != 1 || mgr.stopped[0] != "alice" {
		t.Errorf("expected alice to be stopped, got %v", mgr.stopped)
	}
}

func TestFileWatcher_InvalidYAMLSkipped(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, dir, "bad", "max_watch_streams: -1\n") // fails Validate (< 0)

	mgr := &fakeMgr{}
	w := newTestFileWatcher(dir, mgr)
	w.sync()

	if len(mgr.started) != 0 {
		t.Errorf("expected no starts for invalid config, got %v", mgr.started)
	}
}

func TestFileWatcher_MultipleAccounts(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, dir, "alice", minimalYAML)
	writeYAML(t, dir, "bob", minimalYAML)

	mgr := &fakeMgr{}
	w := newTestFileWatcher(dir, mgr)
	w.sync()

	if len(mgr.started) != 2 {
		t.Errorf("expected 2 starts, got %v", mgr.started)
	}
}

func TestFileWatcher_EmptyDirNoStarts(t *testing.T) {
	dir := t.TempDir()

	mgr := &fakeMgr{}
	w := newTestFileWatcher(dir, mgr)
	w.sync()

	if len(mgr.started) != 0 {
		t.Errorf("expected no starts for empty dir, got %v", mgr.started)
	}
}

// fakeMgr already implements fileManager (Start, Stop, RestartChanged).
var _ fileManager = (*fakeMgr)(nil)

// compile-time check that AccountConfig loads correctly from minimalYAML
func TestFileWatcher_MinimalYAMLValid(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, dir, "testuser", minimalYAML)
	cfgs, err := config.LoadAllAccountConfigs(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfgs) != 1 {
		t.Fatalf("expected 1 config, got %d", len(cfgs))
	}
	if err := config.Validate(cfgs[0]); err != nil {
		t.Fatalf("minimalYAML should be valid: %v", err)
	}
}
