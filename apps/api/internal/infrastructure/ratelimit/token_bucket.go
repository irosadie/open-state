package ratelimit

import (
	"context"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// Config defines the rate and burst for a named rate-limit operation (PRD §83).
type Config struct {
	// Rate is the sustained number of tokens added per second.
	Rate float64
	// Burst is the maximum size of the bucket (allowed burst of requests).
	Burst int
}

// TokenBucket is an in-memory token-bucket rate limiter (PRD §83) implementing
// the domain RateLimiter port. It maintains one independent bucket per key using
// golang.org/x/time/rate. It is safe for concurrent use.
//
// Limitation: state is in-memory and resets on process restart; a shared store
// (e.g. Redis) is a future enhancement. This is acceptable for single-instance
// brute-force/abuse protection.
type TokenBucket struct {
	mu      sync.Mutex
	config  Config
	buckets map[string]*rate.Limiter
}

// NewTokenBucket returns a TokenBucket limiting each key to the given config.
func NewTokenBucket(config Config) *TokenBucket {
	return &TokenBucket{
		config:  config,
		buckets: make(map[string]*rate.Limiter),
	}
}

// Allow reports whether the request for key is within the configured rate.
func (t *TokenBucket) Allow(ctx context.Context, key string) (bool, error) {
	limiter := t.limiterFor(key)
	return limiter.Allow(), nil
}

// limiterFor returns (creating if needed) the token bucket for key. Buckets are
// created lazily so keys not yet seen consume no memory. A background cleanup of
// idle buckets is intentionally omitted for simplicity; buckets are tiny.
func (t *TokenBucket) limiterFor(key string) *rate.Limiter {
	t.mu.Lock()
	defer t.mu.Unlock()

	if l, ok := t.buckets[key]; ok {
		return l
	}
	l := rate.NewLimiter(rate.Limit(t.config.Rate), t.config.Burst)
	t.buckets[key] = l
	return l
}

// WaitForReset is a helper used to compute a Retry-After hint for a key. It
// returns how long until the bucket can accept one more request. This mutates
// the bucket by reserving one token, so it is intended to be called only after
// an Allow() denial when computing the response header.
func (t *TokenBucket) WaitForReset(ctx context.Context, key string) time.Duration {
	limiter := t.limiterFor(key)
	r := limiter.ReserveN(time.Now(), 1)
	if !r.OK() {
		// No tokens available and rate is effectively zero; hint one second.
		return time.Second
	}
	delay := r.Delay()
	r.Cancel()
	return delay
}
