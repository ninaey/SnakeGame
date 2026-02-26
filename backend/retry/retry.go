package retry

import (
	"context"
	"errors"
	"math"
	"time"
)

// ErrMaxAttemptsExceeded is returned when all retries are exhausted.
var ErrMaxAttemptsExceeded = errors.New("max retry attempts exceeded")

// NonRetryableError marks an error that should not be retried (e.g. validation, bad request).
// If fn returns an error wrapping this or implementing this behavior, retry stops immediately.
type NonRetryableError struct{ Err error }

func (e *NonRetryableError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return "non-retryable error"
}

func (e *NonRetryableError) Unwrap() error { return e.Err }

// IsNonRetryable returns true if err is non-retryable (e.g. validation failure).
func IsNonRetryable(err error) bool {
	var nr *NonRetryableError
	return errors.As(err, &nr)
}

// Config holds retry behavior. Stop conditions: success, non-retryable error,
// max attempts reached, or context cancelled.
type Config struct {
	MaxAttempts  int           // max attempts (including first); default 5
	InitialDelay time.Duration // first backoff; default 100ms
	MaxDelay     time.Duration // cap on backoff; default 30s
	Multiplier   float64       // exponential multiplier; default 2.0
}

// DefaultConfig returns a config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		MaxAttempts:  5,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
	}
}

// Do runs fn. On error, retries with exponential backoff until:
// - fn returns nil (success),
// - fn returns a non-retryable error (returned as-is),
// - max attempts are reached (returns ErrMaxAttemptsExceeded wrapping last error),
// - ctx is cancelled (returns ctx.Err() or last error).
func Do(ctx context.Context, cfg Config, fn func() error) error {
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = 5
	}
	if cfg.InitialDelay <= 0 {
		cfg.InitialDelay = 100 * time.Millisecond
	}
	if cfg.MaxDelay <= 0 {
		cfg.MaxDelay = 30 * time.Second
	}
	if cfg.Multiplier <= 0 {
		cfg.Multiplier = 2.0
	}

	var lastErr error
	delay := cfg.InitialDelay

	for attempt := 0; attempt < cfg.MaxAttempts; attempt++ {
		lastErr = fn()
		if lastErr == nil {
			return nil
		}
		if IsNonRetryable(lastErr) {
			return lastErr
		}
		if attempt == cfg.MaxAttempts-1 {
			break
		}
		select {
		case <-ctx.Done():
			return errors.Join(ctx.Err(), lastErr)
		case <-time.After(delay):
			delay = time.Duration(float64(delay) * cfg.Multiplier)
			if delay > cfg.MaxDelay {
				delay = cfg.MaxDelay
			}
		}
	}

	return errors.Join(ErrMaxAttemptsExceeded, lastErr)
}

// ExponentialBackoffDelay returns the delay for the given attempt (0-based).
// Used for testing or custom loops.
func ExponentialBackoffDelay(attempt int, initial, max time.Duration, multiplier float64) time.Duration {
	if attempt <= 0 {
		return initial
	}
	d := float64(initial) * math.Pow(multiplier, float64(attempt))
	if d > float64(max) {
		return max
	}
	return time.Duration(d)
}
