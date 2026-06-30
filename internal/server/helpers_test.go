package server

import (
	"log/slog"
	"testing"

	"github.com/Guliveer/twitch-miner-go/internal/logger"
)

func newTestLogger(t *testing.T) *logger.Logger {
	t.Helper()
	cfg := logger.DefaultConfig()
	cfg.Colored = false
	cfg.Level = slog.Level(100) // suppress all output during tests
	l, err := logger.Setup(cfg)
	if err != nil {
		t.Fatal(err)
	}
	return l
}
