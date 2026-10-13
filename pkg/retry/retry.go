package retry

import (
	"context"
	"math/rand"
	"time"
)

// Config specifies the retry backoff parameters.
type Config struct {
	MaxRetries int
	BaseDelay  time.Duration
	MaxDelay   time.Duration
}

// DefaultConfig provides standard exponential backoff defaults.
var DefaultConfig = Config{
	MaxRetries: 3,
	BaseDelay:  500 * time.Millisecond,
	MaxDelay:   5 * time.Second,
}

// Do executes an operation with exponential backoff and jitter.
func Do(ctx context.Context, cfg Config, op func(ctx context.Context) error) error {
	var err error
	delay := cfg.BaseDelay

	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		err = op(ctx)
		if err == nil {
			return nil
		}

		if attempt == cfg.MaxRetries {
			break
		}

		// Calculate exponential backoff with jitter (+/- 20%)
		jitter := time.Duration(float64(delay) * (0.8 + 0.4*rand.Float64()))
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(jitter):
		}

		delay *= 2
		if delay > cfg.MaxDelay {
			delay = cfg.MaxDelay
		}
	}

	return err
}
