package metrics

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

// Registry holds the application metrics and is registered with Prometheus.
type Registry struct {
	registry *prometheus.Registry

	// RED metrics for HTTP requests (PRD §84).
	httpRequestsTotal   *prometheus.CounterVec
	httpRequestDuration *prometheus.HistogramVec

	// Application metrics.
	auditEntriesTotal     *prometheus.CounterVec
	capabilityInvocations *prometheus.CounterVec
	authFailuresTotal     *prometheus.CounterVec
	mcpGatewayInvocations *prometheus.CounterVec
	mcpGatewayDuration    *prometheus.HistogramVec
}

// New builds the metrics registry, registering runtime/process collectors and
// the application metric vectors (PRD §84).
func New() *Registry {
	reg := prometheus.NewRegistry()
	reg.MustRegister(collectors.NewGoCollector())
	reg.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	r := &Registry{
		registry: reg,
		httpRequestsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total HTTP requests by method, path and status.",
		}, []string{"method", "path", "status"}),
		httpRequestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds.",
			Buckets: prometheus.DefBuckets,
		}, []string{"method", "path"}),
		auditEntriesTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "audit_entries_total",
			Help: "Total audit entries written by action.",
		}, []string{"action"}),
		capabilityInvocations: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "capability_invocations_total",
			Help: "Total capability invocations by result.",
		}, []string{"result"}),
		authFailuresTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "auth_failures_total",
			Help: "Total authentication failures (login/register).",
		}, []string{"reason"}),
		mcpGatewayInvocations: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "mcp_gateway_invocations_total",
			Help: "MCP gateway invocations by safe outcome class.",
		}, []string{"outcome"}),
		mcpGatewayDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name: "mcp_gateway_duration_seconds",
			Help: "MCP gateway provider duration by safe outcome class.",
			Buckets: prometheus.DefBuckets,
		}, []string{"outcome"}),
	}

	reg.MustRegister(r.httpRequestsTotal, r.httpRequestDuration, r.auditEntriesTotal, r.capabilityInvocations, r.authFailuresTotal, r.mcpGatewayInvocations, r.mcpGatewayDuration)
	return r
}

// Prometheus returns the underlying Prometheus registry for HTTP exposition.
func (r *Registry) Prometheus() *prometheus.Registry { return r.registry }

// ObserveHTTP records a request metric (RED, PRD §84).
func (r *Registry) ObserveHTTP(method, path string, status int, seconds float64) {
	r.httpRequestsTotal.WithLabelValues(method, path, strconv.Itoa(status)).Inc()
	r.httpRequestDuration.WithLabelValues(method, path).Observe(seconds)
}

// IncAudit increments the audit volume metric for an action.
func (r *Registry) IncAudit(action string) {
	r.auditEntriesTotal.WithLabelValues(action).Inc()
}

// IncCapabilityInvocation increments the capability invocation metric by result.
func (r *Registry) IncCapabilityInvocation(result string) {
	r.capabilityInvocations.WithLabelValues(result).Inc()
}

// IncAuthFailure increments the auth failure metric by reason.
func (r *Registry) IncAuthFailure(reason string) {
	r.authFailuresTotal.WithLabelValues(reason).Inc()
}

func (r *Registry) IncMCPGatewayInvocation(outcome string) {
	r.mcpGatewayInvocations.WithLabelValues(outcome).Inc()
}

func (r *Registry) ObserveMCPGatewayDuration(outcome string, seconds float64) {
	r.mcpGatewayDuration.WithLabelValues(outcome).Observe(seconds)
}
