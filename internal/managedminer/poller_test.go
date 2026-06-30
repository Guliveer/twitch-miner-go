package managedminer

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/Guliveer/twitch-miner-go/internal/config"
	"github.com/Guliveer/twitch-miner-go/internal/logger"
	"github.com/Guliveer/twitch-miner-go/internal/store"
)

// ── fakes ─────────────────────────────────────────────────────────────────────

type fakePollerStore struct {
	rows []store.AccountRow
	err  error
}

func (f *fakePollerStore) ListAccounts() ([]store.AccountRow, error) { return f.rows, f.err }
func (f *fakePollerStore) GetAccount(string) (*store.AccountRow, error) {
	return nil, nil
}
func (f *fakePollerStore) UpsertAccount(store.AccountRow) error { return nil }
func (f *fakePollerStore) DeleteAccount(string) error           { return nil }
func (f *fakePollerStore) TouchLastStartedAt(string) error      { return nil }
func (f *fakePollerStore) Changes() <-chan struct{}              { return nil }
func (f *fakePollerStore) Close() error                         { return nil }

type fakeMgr struct {
	started   []string
	stopped   []string
	restarted []string
}

func (m *fakeMgr) Start(cfg *config.AccountConfig)          { m.started = append(m.started, cfg.Username) }
func (m *fakeMgr) Stop(username string)                     { m.stopped = append(m.stopped, username) }
func (m *fakeMgr) RestartChanged(cfg *config.AccountConfig) { m.restarted = append(m.restarted, cfg.Username) }

func minimalConfigJSON() string {
	return `{"streamers":[{"username":"s1"}]}`
}

func discardLogger() *logger.Logger {
	cfg := logger.DefaultConfig()
	cfg.Colored = false
	cfg.Level = slog.Level(100)
	l, _ := logger.Setup(cfg)
	return l
}

func newTestPoller(st store.Store, mgr minerManager) *Poller {
	return &Poller{
		store:     st,
		manager:   mgr,
		interval:  time.Minute,
		log:       discardLogger(),
		watermark: make(map[string]int64),
	}
}

// ── tests ─────────────────────────────────────────────────────────────────────

func TestPoller_NewAccountStartsMiner(t *testing.T) {
	st := &fakePollerStore{rows: []store.AccountRow{
		{Username: "alice", ConfigJSON: minimalConfigJSON(), Enabled: true, UpdatedAt: time.Now()},
	}}
	mgr := &fakeMgr{}
	p := newTestPoller(st, mgr)

	p.sync(context.Background())

	if len(mgr.started) != 1 || mgr.started[0] != "alice" {
		t.Errorf("expected alice to be started, got %v", mgr.started)
	}
	if len(mgr.stopped) != 0 {
		t.Errorf("expected no stops, got %v", mgr.stopped)
	}
}

func TestPoller_UpdatedTimestampRestartsMiner(t *testing.T) {
	now := time.Now()
	st := &fakePollerStore{rows: []store.AccountRow{
		{Username: "alice", ConfigJSON: minimalConfigJSON(), Enabled: true, UpdatedAt: now},
	}}
	mgr := &fakeMgr{}
	p := newTestPoller(st, mgr)

	p.sync(context.Background())
	mgr.started = nil

	st.rows[0].UpdatedAt = now.Add(time.Second)
	p.sync(context.Background())

	if len(mgr.restarted) != 1 || mgr.restarted[0] != "alice" {
		t.Errorf("expected alice to be restarted, got %v", mgr.restarted)
	}
}

func TestPoller_SameTimestampNoRestart(t *testing.T) {
	now := time.Now()
	st := &fakePollerStore{rows: []store.AccountRow{
		{Username: "alice", ConfigJSON: minimalConfigJSON(), Enabled: true, UpdatedAt: now},
	}}
	mgr := &fakeMgr{}
	p := newTestPoller(st, mgr)

	p.sync(context.Background())
	p.sync(context.Background())

	if len(mgr.restarted) != 0 {
		t.Errorf("expected no restarts on unchanged timestamp, got %v", mgr.restarted)
	}
}

func TestPoller_DisabledAccountStopsMiner(t *testing.T) {
	now := time.Now()
	st := &fakePollerStore{rows: []store.AccountRow{
		{Username: "alice", ConfigJSON: minimalConfigJSON(), Enabled: true, UpdatedAt: now},
	}}
	mgr := &fakeMgr{}
	p := newTestPoller(st, mgr)

	p.sync(context.Background())

	st.rows[0].Enabled = false
	st.rows[0].UpdatedAt = now.Add(time.Second)
	p.sync(context.Background())

	if len(mgr.stopped) != 1 || mgr.stopped[0] != "alice" {
		t.Errorf("expected alice to be stopped, got %v", mgr.stopped)
	}
}

func TestPoller_RemovedAccountStopsMiner(t *testing.T) {
	now := time.Now()
	st := &fakePollerStore{rows: []store.AccountRow{
		{Username: "alice", ConfigJSON: minimalConfigJSON(), Enabled: true, UpdatedAt: now},
	}}
	mgr := &fakeMgr{}
	p := newTestPoller(st, mgr)

	p.sync(context.Background())

	st.rows = nil
	p.sync(context.Background())

	if len(mgr.stopped) != 1 || mgr.stopped[0] != "alice" {
		t.Errorf("expected alice to be stopped after removal, got %v", mgr.stopped)
	}
}

func TestPoller_InvalidConfigSkipped(t *testing.T) {
	st := &fakePollerStore{rows: []store.AccountRow{
		{Username: "bad", ConfigJSON: `{not json`, Enabled: true, UpdatedAt: time.Now()},
	}}
	mgr := &fakeMgr{}
	p := newTestPoller(st, mgr)

	p.sync(context.Background())

	if len(mgr.started) != 0 {
		t.Errorf("expected no starts for invalid config, got %v", mgr.started)
	}
}

func TestPoller_StoreErrorSkipsSync(t *testing.T) {
	st := &fakePollerStore{err: errors.New("fake db error")}
	mgr := &fakeMgr{}
	p := newTestPoller(st, mgr)

	p.sync(context.Background())

	if len(mgr.started)+len(mgr.stopped)+len(mgr.restarted) != 0 {
		t.Error("expected no manager calls on store error")
	}
}

func TestPoller_MultipleAccounts(t *testing.T) {
	now := time.Now()
	st := &fakePollerStore{rows: []store.AccountRow{
		{Username: "alice", ConfigJSON: minimalConfigJSON(), Enabled: true, UpdatedAt: now},
		{Username: "bob", ConfigJSON: minimalConfigJSON(), Enabled: true, UpdatedAt: now},
	}}
	mgr := &fakeMgr{}
	p := newTestPoller(st, mgr)

	p.sync(context.Background())

	if len(mgr.started) != 2 {
		t.Errorf("expected 2 starts, got %v", mgr.started)
	}
}
