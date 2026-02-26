package payment

import (
	"context"
	"testing"
)

func TestStubGateway_Success(t *testing.T) {
	g := &StubGateway{}
	ctx := context.Background()
	err := g.Charge(ctx, 100, "key-1")
	if err != nil {
		t.Fatalf("expected nil when SimulateTimeout=false, got %v", err)
	}
}

func TestStubGateway_SimulateTimeout(t *testing.T) {
	g := &StubGateway{SimulateTimeout: true}
	ctx := context.Background()
	err := g.Charge(ctx, 100, "key-1")
	if err != ErrTimeout {
		t.Fatalf("expected ErrTimeout when SimulateTimeout=true, got %v", err)
	}
}
