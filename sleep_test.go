package gut

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestSleepMsRespectsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := SleepMs(ctx, 10)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected canceled error, got %v", err)
	}
}

func TestSleepReturnsNilForZeroDuration(t *testing.T) {
	if err := Sleep(context.Background(), 0); err != nil {
		t.Fatalf("unexpected Sleep error: %v", err)
	}
}

func TestBusyWaitRespectsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := BusyWait(ctx, 10*time.Millisecond)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected canceled error, got %v", err)
	}
}

func TestBusyWaitCompletesForPositiveDuration(t *testing.T) {
	if err := BusyWait(context.Background(), time.Millisecond); err != nil {
		t.Fatalf("unexpected BusyWait error: %v", err)
	}
}
