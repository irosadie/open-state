package middleware

import (
	"log/slog"
	"time"

	"github.com/labstack/echo/v4"
)

// RequestLogger is an Echo middleware that emits a structured slog record per
// request with the standard fields (PRD §84): method, path, status,
// duration_ms, request_id, and (when present) user_id and tenant_id.
func RequestLogger(logger *slog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)

			logger.Info("http request",
				"method", c.Request().Method,
				"path", c.Request().URL.Path,
				"status", c.Response().Status,
				"duration_ms", time.Since(start).Milliseconds(),
				"request_id", requestID(c),
				"user_id", userID(c),
				"tenant_id", c.Request().Header.Get(TenantHeader),
				"error", errorString(err),
			)
			return err
		}
	}
}

func requestID(c echo.Context) string {
	if id := c.Response().Header().Get(echo.HeaderXRequestID); id != "" {
		return id
	}
	if id := c.Request().Header.Get(echo.HeaderXRequestID); id != "" {
		return id
	}
	return ""
}

func userID(c echo.Context) string {
	if id, ok := c.Get(UserIDKey).(string); ok {
		return id
	}
	return ""
}

func errorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
