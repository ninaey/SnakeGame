# ADR-001: Idempotency and Payment Retries

## Summary

Checkout is made safe for retries by (1) **idempotency keys** so duplicate requests return the same result without re-charging, and (2) **retrying payment gateway calls** with exponential backoff when the gateway fails (e.g. timeout).

## Idempotency Approach

- **Idempotency-Key header**: Clients send an opaque key (e.g. UUID) on `POST /api/user/orders` (checkout). The server caches the **first response** (status + body) per key for **24 hours**.
- **Repeated requests**: If the same key is sent again within TTL, the server returns the cached response without running checkout again. No double charge, no double balance deduction.
- **Scope**: One key maps to one logical checkout. Keys are not tied to cart contents; the client is responsible for using one key per intended purchase.
- **Storage**: In-memory map (per process). Keys expire after 24h to bound memory.

## Retry Approach

- **When**: The **payment gateway** step is retried on failure (e.g. timeout, temporary error). Validation (empty cart, insufficient balance) is **not** retried; those return 400 immediately.
- **How**: Exponential backoff: 100ms → 200ms → 400ms → … up to 5s, max 5 attempts. Implemented in `retry.Do()`.
- **Stop conditions** (no infinite retries):
  1. **Success**: gateway returns nil → apply balance/cart and return 200.
  2. **Non-retryable error**: caller can wrap errors in `retry.NonRetryableError` to stop immediately (e.g. future validation in gateway).
  3. **Max attempts**: after 5 attempts, return 503 and do not apply any charge.
  4. **Context cancelled**: request timeout or client disconnect stops retries and returns error.

## Risks

- **In-memory idempotency**: Not shared across instances; duplicate keys on different servers can both run checkout. Mitigation: single instance or external store (Redis/DB) for keys in production.
- **Key reuse**: If a client reuses a key after 24h, the key may have expired and a new checkout will run; acceptable if key TTL is understood.
- **No idempotency on gateway**: The stub gateway does not dedupe by key. A real gateway should accept the same idempotency key and return the same result for duplicate calls.

## Future Work

- Persist idempotency keys in Redis or DB for multi-instance and restarts.
- Pass Idempotency-Key to the real payment gateway and honor gateway-level idempotency.
- Add metrics for retry attempts and 503 rate; alert on high payment failure rate.
- Consider circuit breaker around the gateway after repeated failures.
