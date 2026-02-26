package retry

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDo_SuccessFirstTry(t *testing.T) {
	ctx := context.Background()
	calls := 0
	err := Do(ctx, DefaultConfig(), func() error {
		calls++
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 1 {
		t.Errorf("expected 1 call, got %d", calls)
	}
}

func TestDo_SuccessAfterRetry(t *testing.T) {
	ctx := context.Background()
	calls := 0
	err := Do(ctx, Config{MaxAttempts: 4, InitialDelay: 1 * time.Millisecond, MaxDelay: 5 * time.Millisecond}, func() error {
		calls++
		if calls < 3 {
			return errors.New("temp")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 3 {
		t.Errorf("expected 3 calls, got %d", calls)
	}
}

func TestDo_StopOnNonRetryable(t *testing.T) {
	ctx := context.Background()
	calls := 0
	validationErr := &NonRetryableError{Err: errors.New("invalid request")}
	err := Do(ctx, Config{MaxAttempts: 5, InitialDelay: 1 * time.Millisecond}, func() error {
		calls++
		return validationErr
	})
	if err != validationErr {
		t.Fatalf("expected validationErr, got %v", err)
	}
	if calls != 1 {
		t.Errorf("expected 1 call (no retries), got %d", calls)
	}
}

func TestDo_MaxAttemptsExceeded(t *testing.T) {
	ctx := context.Background()
	calls := 0
	err := Do(ctx, Config{MaxAttempts: 3, InitialDelay: 1 * time.Millisecond, MaxDelay: 2 * time.Millisecond}, func() error {
		calls++
		return errors.New("gateway timeout")
	})
	if err == nil {
		t.Fatal("expected error after max attempts")
	}
	if !errors.Is(err, ErrMaxAttemptsExceeded) {
		t.Errorf("expected ErrMaxAttemptsExceeded, got %v", err)
	}
	if calls != 3 {
		t.Errorf("expected 3 calls, got %d", calls)
	}
}

func TestDo_ContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	calls := 0
	err := Do(ctx, Config{MaxAttempts: 5, InitialDelay: 1 * time.Millisecond}, func() error {
		calls++
		return errors.New("timeout")
	})
	if err == nil {
		t.Fatal("expected error when context canceled")
	}
	if calls != 1 {
		t.Errorf("expected 1 call then stop, got %d", calls)
	}
}

func TestIsNonRetryable(t *testing.T) {
	if IsNonRetryable(nil) {
		t.Error("nil should not be non-retryable")
	}
	if IsNonRetryable(errors.New("plain")) {
		t.Error("plain error should not be non-retryable")
	}
	nr := &NonRetryableError{Err: errors.New("bad")}
	if !IsNonRetryable(nr) {
		t.Error("NonRetryableError should be non-retryable")
	}
}

func TestExponentialBackoffDelay(t *testing.T) {
	initial := 100 * time.Millisecond
	max := 5 * time.Second
	mult := 2.0

	if d := ExponentialBackoffDelay(0, initial, max, mult); d != initial {
		t.Errorf("attempt 0: want %v, got %v", initial, d)
	}
	d1 := ExponentialBackoffDelay(1, initial, max, mult)
	if d1 != 200*time.Millisecond {
		t.Errorf("attempt 1: want 200ms, got %v", d1)
	}
	d3 := ExponentialBackoffDelay(3, initial, max, mult)
	if d3 != 800*time.Millisecond {
		t.Errorf("attempt 3: want 800ms, got %v", d3)
	}
	// Should cap at max
	d10 := ExponentialBackoffDelay(10, initial, max, mult)
	if d10 != max {
		t.Errorf("attempt 10: want max %v, got %v", max, d10)
	}
}
