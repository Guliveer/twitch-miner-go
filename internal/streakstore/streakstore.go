// Package streakstore persists which watch streaks have already been earned,
// so a restart does not spend slot time chasing streaks Twitch has already paid
// out.
//
// A streak belongs to a single broadcast: Twitch grants it once per stream, and
// a channel that goes offline and comes back starts a new one. Records are
// therefore keyed by channel and qualified by broadcast ID — a stored record
// only suppresses the chase while the same broadcast is still running.
package streakstore

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// DefaultRetention is how long a record is kept. Broadcasts do not outlive it,
// so anything older is only taking up space.
const DefaultRetention = 14 * 24 * time.Hour

// Record is one earned watch streak.
type Record struct {
	BroadcastID string    `json:"broadcast_id"`
	EarnedAt    time.Time `json:"earned_at"`
}

// file is the on-disk shape, versioned so the format can change later without
// a silent misread.
type file struct {
	Version int               `json:"version"`
	Streaks map[string]Record `json:"streaks"`
}

const currentVersion = 1

// Store is a concurrency-safe set of earned watch streaks backed by a JSON file.
type Store struct {
	mu        sync.Mutex
	path      string
	records   map[string]Record
	retention time.Duration
}

// Open loads the store for an account, creating the directory if needed.
// A missing or unreadable file yields an empty store rather than an error:
// losing the record costs one harvest cycle, refusing to start costs the run.
func Open(dir, account string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("creating streak directory %s: %w", dir, err)
	}

	s := &Store{
		path:      filepath.Join(dir, strings.ToLower(account)+".json"),
		records:   make(map[string]Record),
		retention: DefaultRetention,
	}

	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return s, nil
		}
		return s, fmt.Errorf("reading %s: %w", s.path, err)
	}

	var parsed file
	if err := json.Unmarshal(data, &parsed); err != nil {
		return s, fmt.Errorf("parsing %s: %w", s.path, err)
	}
	if parsed.Version != currentVersion {
		return s, fmt.Errorf("unsupported streak store version %d in %s", parsed.Version, s.path)
	}

	now := time.Now()
	for channel, rec := range parsed.Streaks {
		if now.Sub(rec.EarnedAt) <= s.retention {
			s.records[strings.ToLower(channel)] = rec
		}
	}

	return s, nil
}

// Path returns the backing file path.
func (s *Store) Path() string { return s.path }

// Len returns how many records are held.
func (s *Store) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.records)
}

// Earned reports whether the streak for this channel's current broadcast has
// already been collected. An empty broadcast ID always reports false: without
// it there is no way to tell this broadcast from the previous one, and wrongly
// skipping a chase costs a streak.
func (s *Store) Earned(channel, broadcastID string) bool {
	if broadcastID == "" {
		return false
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	rec, ok := s.records[strings.ToLower(channel)]
	if !ok || rec.BroadcastID != broadcastID {
		return false
	}
	return time.Since(rec.EarnedAt) <= s.retention
}

// Record stores an earned streak and flushes to disk immediately — streaks are
// rare enough that batching would only risk losing them to a hard restart.
func (s *Store) Record(channel, broadcastID string, at time.Time) error {
	if broadcastID == "" {
		return nil
	}

	s.mu.Lock()
	s.records[strings.ToLower(channel)] = Record{BroadcastID: broadcastID, EarnedAt: at}
	data, err := s.encodeLocked(at)
	s.mu.Unlock()

	if err != nil {
		return err
	}
	return s.write(data)
}

// encodeLocked serialises the records, dropping expired ones. Caller holds mu.
func (s *Store) encodeLocked(now time.Time) ([]byte, error) {
	out := file{Version: currentVersion, Streaks: make(map[string]Record, len(s.records))}
	for channel, rec := range s.records {
		if now.Sub(rec.EarnedAt) > s.retention {
			delete(s.records, channel)
			continue
		}
		out.Streaks[channel] = rec
	}

	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encoding streak store: %w", err)
	}
	return data, nil
}

// write replaces the file atomically so a crash mid-write cannot leave a
// truncated file that would fail to parse on the next start.
func (s *Store) write(data []byte) error {
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return fmt.Errorf("writing %s: %w", tmp, err)
	}
	if err := os.Rename(tmp, s.path); err != nil {
		return fmt.Errorf("replacing %s: %w", s.path, err)
	}
	return nil
}
