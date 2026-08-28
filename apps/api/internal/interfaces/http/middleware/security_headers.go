package middleware

import (
	"fmt"

	"github.com/labstack/echo/v4"
)

// SecurityHeadersConfig configures the security headers middleware (PRD §139,
// §84). Safe defaults are applied when values are empty.
type SecurityHeadersConfig struct {
	// ContentSecurityPolicy string, e.g. "default-src 'self'". Empty uses a
	// restrictive default for a JSON API.
	ContentSecurityPolicy string
	// EnableHSTS toggles Strict-Transport-Security. Disabled for plain-HTTP dev.
	EnableHSTS bool
	// HSTSMaxAge in seconds (default 31536000 = 1 year).
	HSTSMaxAge int
}

// SecurityHeaders returns an Echo middleware that sets recommended security
// headers on every response (PRD §139).
func SecurityHeaders(cfg SecurityHeadersConfig) echo.MiddlewareFunc {
	csp := cfg.ContentSecurityPolicy
	if csp == "" {
		csp = "default-src 'self'"
	}
	maxAge := cfg.HSTSMaxAge
	if maxAge <= 0 {
		maxAge = 31536000
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			h := c.Response().Header()
			h.Set("X-Content-Type-Options", "nosniff")
			h.Set("X-Frame-Options", "DENY")
			h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
			h.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
			h.Set("X-XSS-Protection", "0")
			h.Set("Content-Security-Policy", csp)

			if cfg.EnableHSTS {
				h.Set("Strict-Transport-Security", fmt.Sprintf("max-age=%d; includeSubDomains", maxAge))
			}

			return next(c)
		}
	}
}
