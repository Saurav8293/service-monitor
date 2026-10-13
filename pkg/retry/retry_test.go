package retry_test

import (
	"context"
	"errors"
	"service-monitor/pkg/retry"
	"testing"
	"time"
)

func TestRetrySuccess(t *testing.T) {
	attempts := 0
	cfg := retry.Config{
		MaxRetries: 3,
		BaseDelay:  5 * time.Millisecond,
		MaxDelay:   20 * time.Millisecond,
	}

	err := retry.Do(context.Background(), cfg, func(ctx context.Context) error {
		attempts++
		if attempts < 3 {
			return errors.New("temporary error")
		}
		return nil
	})

	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
}

func TestRetryExceedsMax(t *testing.T) {
	attempts := 0
	cfg := retry.Config{
		MaxRetries: 2,
		BaseDelay:  5 * time.Millisecond,
		MaxDelay:   20 * time.Millisecond,
	}

	err := retry.Do(context.Background(), cfg, func(ctx context.Context) error {
		attempts++
		return errors.New("persistent error")
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if attempts != 3 { // attempt 0, 1, 2 = 3 total attempts
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
}

func TestRetryContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	cfg := retry.Config{
		MaxRetries: 3,
		BaseDelay:  50 * time.Millisecond,
		MaxDelay:   100 * time.Millisecond,
	}

	err := retry.Do(ctx, cfg, func(ctx context.Context) error {
		return errors.New("failed")
	})

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}
