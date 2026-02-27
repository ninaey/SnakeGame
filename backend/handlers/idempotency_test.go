package handlers

import (
	"testing"
)

// Test the idempotency cache functions.
func TestGetIdempotency_EmptyKey(t *testing.T) {
	status, body, ok := getIdempotency("")
	if ok {
		t.Fatal("getIdempotency(empty) should return ok=false")
	}
	if status != 0 || body != nil {
		t.Errorf("expected status=0 body=nil; got status=%d body=%v", status, body)
	}
}

// Test idempotency roundtrip: set a key, get it, check it's the same.
func TestSetGetIdempotency_RoundTrip(t *testing.T) {
	key := "test-key-roundtrip-1"
	statusCode := 200
	payload := []byte(`{"Status":"Success"}`)

	setIdempotency(key, statusCode, payload)

	status, body, ok := getIdempotency(key)
	if !ok {
		t.Fatal("getIdempotency after set should return ok=true")
	}
	if status != statusCode {
		t.Errorf("status: want %d, got %d", statusCode, status)
	}
	if string(body) != string(payload) {
		t.Errorf("body: want %q, got %q", payload, body)
	}
	// Idempotent get: same key returns same result again
	status2, body2, ok2 := getIdempotency(key)
	if !ok2 || status2 != status || string(body2) != string(body) {
		t.Errorf("second get: want same result; ok=%v status=%d body=%s", ok2, status2, body2)
	}
}

// Test idempotency with unknown key: should return ok=false.
func TestGetIdempotency_UnknownKey(t *testing.T) {
	status, body, ok := getIdempotency("nonexistent-key-12345")
	if ok {
		t.Fatal("getIdempotency(unknown key) should return ok=false")
	}
	if status != 0 || body != nil {
		t.Errorf("expected status=0 body=nil; got status=%d body=%v", status, body)
	}
}

// Test idempotency overwrite: should overwrite the key with the new status and body.
func TestSetIdempotency_Overwrite(t *testing.T) {
	key := "test-key-overwrite"
	setIdempotency(key, 200, []byte(`"first"`))
	setIdempotency(key, 503, []byte(`"second"`))

	status, body, ok := getIdempotency(key)
	if !ok {
		t.Fatal("getIdempotency should return ok=true after overwrite")
	}
	if status != 503 {
		t.Errorf("status: want 503, got %d", status)
	}
	if string(body) != `"second"` {
		t.Errorf("body: want \"second\", got %q", body)
	}
}

// Test idempotency empty key: should ignore the key and return ok=false.
func TestSetIdempotency_EmptyKeyIgnored(t *testing.T) {
	setIdempotency("", 200, []byte("x"))
	// Should not panic; get with empty key still returns false
	_, _, ok := getIdempotency("")
	if ok {
		t.Fatal("empty key should not be stored")
	}
}
