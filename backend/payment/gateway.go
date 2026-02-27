package payment

import (
	"context"
	"errors"
	"time"
)

// ErrTimeout is returned when the payment gateway times out.
var ErrTimeout = errors.New("payment gateway timeout")

// Gateway performs the actual charge. In production this would call an external API.
type Gateway interface {
	Charge(ctx context.Context, amountCents int, idempotencyKey string) error
}

// StubGateway is an in-memory stub. When SimulateTimeout is true, Charge returns ErrTimeout
// (to test retry logic). Otherwise it succeeds immediately.
type StubGateway struct {
	SimulateTimeout bool
	// SimulateDelay optionally sleeps before returning (e.g. to simulate slow response).
	SimulateDelay time.Duration
}

// Charge implements Gateway. When SimulateTimeout is true, returns ErrTimeout.
func (g *StubGateway) Charge(ctx context.Context, amountCents int, idempotencyKey string) error {
	if g.SimulateDelay > 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(g.SimulateDelay):
		}
	}
	if g.SimulateTimeout {
		return ErrTimeout
	}
	return nil
}
