package capability

import (
	"context"
	"math"
	"math/rand"
	"time"
)

// backoff calculates the exponential backoff delay with jitter for a retry
// attempt (0-based), given a base delay (PRD §88).
func backoff(base time.Duration, attempt int) time.Duration {
	if base <= 0 {
		base = 100 * time.Millisecond
	}
	exp := math.Pow(2, float64(attempt))
	delay := float64(base) * exp
	// full jitter: random in [0, delay]
	return time.Duration(rand.Float64() * delay)
}

// retryableError reports whether an error is retryable for the given policy.
func retryableError(err error, policy CapabilityPolicy) bool {
	ce, ok := err.(*CapabilityError)
	if !ok {
		return false
	}
	if len(policy.Retryable) > 0 {
		for _, k := range policy.Retryable {
			if ce.Kind == k {
				return true
			}
		}
		return false
	}
	return ce.Retryable()
}

// waitForRetry sleeps until the context is done or the backoff elapses.
func waitForRetry(ctx context.Context, delay time.Duration) bool {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
