package config

import (
	"os"
	"strconv"
)

// RateLimitConfig holds the per-operation rate-limit settings (PRD §83). Each
// operation has a sustained rate (tokens/second) and a burst size. Values come
// from environment variables with documented defaults.
type RateLimitConfig struct {
	Login      RateLimitOperation
	Register   RateLimitOperation
	Capability RateLimitOperation
}

// RateLimitOperation is the rate + burst for a single operation.
type RateLimitOperation struct {
	Rate  float64
	Burst int
}

// Default rate-limit values. Tuned to block brute-force/abuse while not
// inconveniencing normal use.
const (
	defaultLoginRate       = 0.5 // ~1 request / 2s sustained
	defaultLoginBurst      = 10
	defaultRegisterRate    = 0.1 // ~1 request / 10s sustained
	defaultRegisterBurst   = 5
	defaultCapabilityRate  = 10 // 10 invocations / s sustained
	defaultCapabilityBurst = 30
)

// loadRateLimits reads rate-limit env vars (with defaults) and returns the
// consolidated RateLimitConfig.
func loadRateLimits() RateLimitConfig {
	return RateLimitConfig{
		Login: RateLimitOperation{
			Rate:  envFloat("RATE_LIMIT_LOGIN_RATE", defaultLoginRate),
			Burst: envInt("RATE_LIMIT_LOGIN_BURST", defaultLoginBurst),
		},
		Register: RateLimitOperation{
			Rate:  envFloat("RATE_LIMIT_REGISTER_RATE", defaultRegisterRate),
			Burst: envInt("RATE_LIMIT_REGISTER_BURST", defaultRegisterBurst),
		},
		Capability: RateLimitOperation{
			Rate:  envFloat("RATE_LIMIT_CAPABILITY_RATE", defaultCapabilityRate),
			Burst: envInt("RATE_LIMIT_CAPABILITY_BURST", defaultCapabilityBurst),
		},
	}
}

func envFloat(key string, def float64) float64 {
	raw := os.Getenv(key)
	if raw == "" {
		return def
	}
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil || v < 0 {
		return def
	}
	return v
}

func envInt(key string, def int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return def
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v < 1 {
		return def
	}
	return v
}
