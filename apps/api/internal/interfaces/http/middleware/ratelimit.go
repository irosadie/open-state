package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	domainsvc "github.com/irosadie/open-state/api/internal/domain/services"
	"github.com/labstack/echo/v4"
)

// RateLimitKeyFunc builds the scope key for a request. Returning ok=false means
// the key cannot be derived (e.g. missing identity) and the request proceeds
// without rate limiting (fail-open for key derivation, not for the limiter).
type RateLimitKeyFunc func(c echo.Context) (key string, ok bool)

// RateLimit is an Echo middleware that rate-limits requests per scope key (PRD
// §83). It fails open: if the limiter returns an error, the request proceeds and
// the error is logged, so a limiter outage does not lock out users. When the
// limit is exceeded it returns 429 with a Retry-After header.
func RateLimit(limiter domainsvc.RateLimiter, buildKey RateLimitKeyFunc) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			key, ok := buildKey(c)
			if !ok {
				// Cannot derive a scope key; do not rate limit this request.
				return next(c)
			}

			allowed, err := limiter.Allow(c.Request().Context(), key)
			if err != nil {
				// Fail open: log and proceed so an outage does not lock out users.
				c.Logger().Warnf("rate limiter error (key=%s): %v", key, err)
				return next(c)
			}
			if !allowed {
				c.Response().Header().Set("Retry-After", "1")
				return &echo.HTTPError{
					Code:    http.StatusTooManyRequests,
					Message: "rate limit exceeded",
				}
			}
			return next(c)
		}
	}
}

// LoginKey builds a scope key from the request email (login brute-force per
// account, PRD §83). Falls back to the client IP when the body cannot be bound.
// The body is read non-destructively and restored so the controller can still
// bind it.
func LoginKey(c echo.Context) (string, bool) {
	body, _ := io.ReadAll(c.Request().Body)
	c.Request().Body = io.NopCloser(bytes.NewReader(body))
	if len(body) > 0 {
		var req struct {
			Email string `json:"email"`
		}
		if err := json.Unmarshal(body, &req); err == nil && req.Email != "" {
			return "route:login:email:" + req.Email, true
		}
	}
	if ip := clientIP(c); ip != "" {
		return "route:login:ip:" + ip, true
	}
	return "", false
}

// RegisterKey builds a scope key from the client IP (mass-registration
// protection, PRD §83). There is no account identity before registration.
func RegisterKey(c echo.Context) (string, bool) {
	if ip := clientIP(c); ip != "" {
		return "route:register:ip:" + ip, true
	}
	return "", false
}

// clientIP extracts the client IP, honoring the X-Forwarded-For header when
// behind a proxy and falling back to the remote address.
func clientIP(c echo.Context) string {
	if xff := c.Request().Header.Get("X-Forwarded-For"); xff != "" {
		if first := strings.TrimSpace(strings.SplitN(xff, ",", 2)[0]); first != "" {
			return first
		}
	}
	if ra := c.RealIP(); ra != "" {
		return ra
	}
	return ""
}
