package gut

import (
	"context"
	"runtime"
	"time"

	"github.com/PandelisZ/gut/util"
)

func Sleep(ctx context.Context, duration time.Duration) error {
	return util.SleepContext(ctx, duration)
}

func SleepMs(ctx context.Context, ms int) error {
	return Sleep(ctx, time.Duration(ms)*time.Millisecond)
}

func BusyWait(ctx context.Context, duration time.Duration) error {
	if duration <= 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			return nil
		}
	}

	deadline := time.Now().Add(duration)
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			runtime.Gosched()
		}
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}
