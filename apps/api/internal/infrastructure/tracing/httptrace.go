package tracing

import (
	"log/slog"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// HTTPTrace provides an Echo middleware that creates a server span per request
// (PRD §84). It extracts an incoming trace context from the traceparent header
// (distributed tracing), sets the span as the parent for downstream operations,
// and records method/path/status attributes.
type HTTPTrace struct {
	tracer trace.Tracer
	logger *slog.Logger
}

// NewHTTPTrace builds an HTTPTrace middleware.
func NewHTTPTrace(logger *slog.Logger) *HTTPTrace {
	return &HTTPTrace{tracer: otel.Tracer("openstate.http"), logger: logger}
}

// Middleware returns the Echo middleware that instruments inbound requests.
func (m *HTTPTrace) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			ctx := req.Context()

			// Extract incoming trace context (traceparent) for distributed tracing.
			ctx = otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier(req.Header))

			spanCtx, span := m.tracer.Start(ctx, "HTTP "+req.Method)
			defer span.End()

			span.SetAttributes(
				attribute.String("http.request.method", req.Method),
				attribute.String("url.path", req.URL.Path),
			)

			// Propagate the span context so downstream services (DB, worker) become
			// children of this span.
			c.SetRequest(req.WithContext(spanCtx))

			err := next(c)
			span.SetAttributes(attribute.Int("http.response.status_code", c.Response().Status))
			return err
		}
	}
}
