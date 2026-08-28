package services

import "context"

// RateLimiter is the domain port for rate limiting (PRD §83). It is used by the
// HTTP layer (auth brute-force protection) and the capability invocation chain
// (abuse protection). `Allow` returns true when the request is permitted and
// false when the configured rate for `key` has been exceeded.
type RateLimiter interface {
	Allow(ctx context.Context, key string) (bool, error)
}
