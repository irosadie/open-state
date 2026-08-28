package tracing

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Config controls the OpenTelemetry TracerProvider.
type Config struct {
	// OTLPEndpoint is the OTLP/HTTP collector endpoint. Empty disables tracing
	// (no-op) so the service starts without a collector.
	OTLPEndpoint string
	// ServiceName identifies this service in traces.
	ServiceName string
}

// Setup creates a TracerProvider, sets it as the global provider, and returns a
// shutdown function. When no OTLP endpoint is configured, a no-op provider is
// used so the service always starts. Call the returned shutdown on process exit
// to flush pending spans (PRD §84).
func Setup(ctx context.Context, cfg Config, logger *slog.Logger) (shutdown func(context.Context) error) {
	if cfg.OTLPEndpoint == "" {
		logger.Info("OTel tracing disabled (no OTEL_EXPORTER_OTLP_ENDPOINT)")
		return func(context.Context) error { return nil }
	}

	res := resource.NewWithAttributes(
		resource.Default().SchemaURL(),
		attribute.String("service.name", cfg.ServiceName),
	)

	exporter, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpointURL(cfg.OTLPEndpoint))
	if err != nil {
		logger.Warn("OTel exporter setup failed; tracing disabled", "error", err.Error())
		return func(context.Context) error { return nil }
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	logger.Info("OTel tracing enabled", "endpoint", cfg.OTLPEndpoint, "service", cfg.ServiceName)
	return tp.Shutdown
}
