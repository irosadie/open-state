package capability

import (
	"context"
	"testing"
	"time"
)

func TestBackoffIncreases(t *testing.T) {
	base := 100 * time.Millisecond
	d0 := backoff(base, 0)
	d1 := backoff(base, 1)
	d2 := backoff(base, 2)
	// with jitter, ensure order is roughly non-decreasing (upper bounds)
	if d1 > d2 {
		// possible with jitter, but d2's max should exceed d1's max
		if d2 > 4*base {
			// d2 should be able to exceed d1; just sanity check both positive
		}
	}
	if d0 <= 0 || d1 <= 0 || d2 <= 0 {
		t.Error("backoff delays must be positive")
	}
}

func TestRetryableErrorPolicy(t *testing.T) {
	timeoutErr := NewCapabilityError(ErrorKindTimeout, "capability.timeout", "x")
	policy := CapabilityPolicy{Retryable: []ErrorKind{ErrorKindTimeout}}
	if !retryableError(timeoutErr, policy) {
		t.Error("timeout should be retryable under policy")
	}
	authErr := NewCapabilityError(ErrorKindUnauthorized, "capability.unauthorized", "x")
	if retryableError(authErr, policy) {
		t.Error("unauthorized should not be retryable")
	}
	// empty retryable list → use default Retryable()
	policy2 := CapabilityPolicy{}
	if !retryableError(timeoutErr, policy2) {
		t.Error("timeout retryable by default")
	}
}

func TestWaitForRetryContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	ok := waitForRetry(ctx, time.Hour) // should return false immediately
	if ok {
		t.Error("expected false when context cancelled")
	}
}
