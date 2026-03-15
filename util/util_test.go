package util

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestMilliseconds(t *testing.T) {
	if got := Milliseconds(250); got != 250*time.Millisecond {
		t.Fatalf("unexpected duration: %v", got)
	}
}

func TestSleepContextReturnsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := SleepContext(ctx, 10*time.Millisecond); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected cancellation error, got %v", err)
	}
}

func TestSleepContextSleepsUntilDone(t *testing.T) {
	ctx := context.Background()
	started := time.Now()
	if err := SleepContext(ctx, 5*time.Millisecond); err != nil {
		t.Fatalf("unexpected sleep error: %v", err)
	}
	if time.Since(started) < 5*time.Millisecond {
		t.Fatal("expected SleepContext to wait for the requested duration")
	}
}

func TestWithOptionalTimeout(t *testing.T) {
	ctx, cancel := WithOptionalTimeout(context.Background(), 5*time.Millisecond)
	defer cancel()

	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatal("expected context deadline")
	}
	if time.Until(deadline) <= 0 {
		t.Fatal("expected deadline in the future")
	}
}

func TestPtr(t *testing.T) {
	value := Ptr("gut")
	if value == nil || *value != "gut" {
		t.Fatalf("unexpected pointer result: %v", value)
	}
}
