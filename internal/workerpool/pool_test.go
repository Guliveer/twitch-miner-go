package workerpool

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestRunEmpty(t *testing.T) {
	err := Run(context.Background(), []int{}, 4, func(_ context.Context, _ int) error {
		t.Error("fn must not be called for empty input")
		return nil
	})
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestRunAllItems(t *testing.T) {
	items := []int{1, 2, 3, 4, 5}
	var count atomic.Int64
	err := Run(context.Background(), items, 2, func(_ context.Context, _ int) error {
		count.Add(1)
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if count.Load() != int64(len(items)) {
		t.Errorf("expected %d calls, got %d", len(items), count.Load())
	}
}

func TestRunReturnsFirstError(t *testing.T) {
	sentinel := errors.New("sentinel error")
	items := []int{1, 2, 3}
	err := Run(context.Background(), items, 2, func(_ context.Context, i int) error {
		if i == 2 {
			return sentinel
		}
		return nil
	})
	if !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}

func TestRunWorkersBound(t *testing.T) {
	// Verify that at most `workers` goroutines run concurrently.
	const workers = 3
	const items = 20

	var active atomic.Int64
	var maxSeen atomic.Int64

	err := Run(context.Background(), make([]int, items), workers, func(_ context.Context, _ int) error {
		cur := active.Add(1)
		for {
			old := maxSeen.Load()
			if cur <= old || maxSeen.CompareAndSwap(old, cur) {
				break
			}
		}
		time.Sleep(5 * time.Millisecond)
		active.Add(-1)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if maxSeen.Load() > workers {
		t.Errorf("max concurrent goroutines = %d, want <= %d", maxSeen.Load(), workers)
	}
}

func TestRunZeroWorkersDefaultsToOne(t *testing.T) {
	var count atomic.Int64
	err := Run(context.Background(), []int{1, 2, 3}, 0, func(_ context.Context, _ int) error {
		count.Add(1)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if count.Load() != 3 {
		t.Errorf("expected 3 calls, got %d", count.Load())
	}
}

func TestRunContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel before start

	var count atomic.Int64
	// With context already cancelled, most items should be skipped.
	// At least the loop checks ctx.Err() before each dispatch.
	Run(ctx, make([]int, 100), 4, func(_ context.Context, _ int) error {
		count.Add(1)
		return nil
	})
	// Should process far fewer than 100 items (likely 0 or a handful in-flight)
	if count.Load() >= 100 {
		t.Errorf("expected context cancellation to skip items, got %d/100 processed", count.Load())
	}
}
