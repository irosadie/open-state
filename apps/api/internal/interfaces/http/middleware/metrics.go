package middleware

import (
	"time"

	"github.com/labstack/echo/v4"
)

// MetricsRecorder is the minimal interface the HTTP layer needs to record RED
// metrics (PRD §84). It is satisfied by infrastructure/metrics.Registry.
type MetricsRecorder interface {
	ObserveHTTP(method, path string, status int, seconds float64)
}

// Metrics is an Echo middleware that records per-request RED metrics.
func Metrics(rec MetricsRecorder) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)
			if rec != nil {
				rec.ObserveHTTP(c.Request().Method, c.Path(), c.Response().Status, time.Since(start).Seconds())
			}
			return err
		}
	}
}
