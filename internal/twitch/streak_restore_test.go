package twitch

import (
	"testing"

	"github.com/Guliveer/twitch-miner-go/internal/logger"
	"github.com/Guliveer/twitch-miner-go/internal/model"
)

// fakeStreaks is a StreakLookup that reports one known (channel, broadcast).
type fakeStreaks struct {
	channel     string
	broadcastID string
	queries     int
}

func (f *fakeStreaks) Earned(channel, broadcastID string) bool {
	f.queries++
	return channel == f.channel && broadcastID == f.broadcastID
}

func clientWithStreaks(t *testing.T, l StreakLookup) *Client {
	t.Helper()
	log, err := logger.Setup(logger.Config{Level: 100}) // suppress output
	if err != nil {
		t.Fatalf("logger: %v", err)
	}
	c := &Client{Log: log}
	c.SetStreakLookup(l)
	return c
}

// streamerWithBroadcast builds a streamer freshly marked online, which is the
// state restoreWatchStreak has to correct.
func streamerWithBroadcast(username, broadcastID string) *model.Streamer {
	s := model.NewStreamer(username)
	s.Settings = model.DefaultStreamerSettings()
	s.IsOnline = true
	s.Stream.BroadcastID = broadcastID
	s.Stream.IsWatchStreakMissing = true
	return s
}

func TestRestoreClearsFlagForKnownBroadcast(t *testing.T) {
	c := clientWithStreaks(t, &fakeStreaks{channel: "krissi", broadcastID: "b-1"})
	s := streamerWithBroadcast("krissi", "b-1")

	c.restoreWatchStreak(s)

	if s.Stream.IsWatchStreakMissing {
		t.Fatal("a streak already collected for this broadcast should not be chased again")
	}
}

func TestRestoreKeepsChasingNewBroadcast(t *testing.T) {
	c := clientWithStreaks(t, &fakeStreaks{channel: "krissi", broadcastID: "b-1"})
	s := streamerWithBroadcast("krissi", "b-2")

	c.restoreWatchStreak(s)

	if !s.Stream.IsWatchStreakMissing {
		t.Fatal("a new broadcast earns a fresh streak and must stay in the rotation")
	}
}

func TestRestoreKeepsChasingUnknownChannel(t *testing.T) {
	c := clientWithStreaks(t, &fakeStreaks{channel: "krissi", broadcastID: "b-1"})
	s := streamerWithBroadcast("someone-else", "b-1")

	c.restoreWatchStreak(s)

	if !s.Stream.IsWatchStreakMissing {
		t.Fatal("an unrelated channel must stay in the rotation")
	}
}

func TestRestoreDoesNotQueryWhenStreakAlreadySettled(t *testing.T) {
	fake := &fakeStreaks{channel: "krissi", broadcastID: "b-1"}
	c := clientWithStreaks(t, fake)

	s := streamerWithBroadcast("krissi", "b-1")
	s.Stream.IsWatchStreakMissing = false

	c.restoreWatchStreak(s)

	if fake.queries != 0 {
		t.Errorf("no lookup is needed when nothing is pending, got %d", fake.queries)
	}
}

func TestRestoreIsNoOpWithoutAStore(t *testing.T) {
	log, err := logger.Setup(logger.Config{Level: 100})
	if err != nil {
		t.Fatalf("logger: %v", err)
	}
	c := &Client{Log: log}

	s := streamerWithBroadcast("krissi", "b-1")
	c.restoreWatchStreak(s)

	if !s.Stream.IsWatchStreakMissing {
		t.Fatal("without a store the streak state must be left untouched")
	}
}
