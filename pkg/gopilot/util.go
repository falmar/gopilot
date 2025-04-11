package gopilot

import (
	"context"
	"time"
)

func sleepWithCtx(ctx context.Context, d time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.NewTimer(d).C:
		return nil
	}
}
