package middleware

import (
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
)

// CORSConfig hardens cross-origin access (PRD §74). Only explicit allow-list
// origins receive CORS headers; in production, an empty allow-list denies all
// cross-origin requests (same-origin still works).
type CORSConfig struct {
	// AllowedOrigins is the explicit origin allow-list. Empty denies cross-origin.
	AllowedOrigins []string
}

// CORS returns an Echo middleware with hardened CORS settings: explicit origin
// allow-list, restricted methods, and the headers the API actually uses.
func CORS(cfg CORSConfig) echo.MiddlewareFunc {
	methods := []string{echo.GET, echo.POST, echo.PUT, echo.PATCH, echo.DELETE, echo.OPTIONS}
	headers := []string{"Authorization", "Content-Type", "X-Tenant-ID", "X-Project-ID"}

	return echomw.CORSWithConfig(echomw.CORSConfig{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowMethods:     methods,
		AllowHeaders:     headers,
		AllowCredentials: len(cfg.AllowedOrigins) > 0,
	})
}
