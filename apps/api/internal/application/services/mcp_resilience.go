package services

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	domaincap "github.com/irosadie/open-state/api/internal/domain/capability"
	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
	domainservices "github.com/irosadie/open-state/api/internal/domain/services"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type MCPGatewayMetrics interface {
	IncMCPGatewayInvocation(result string)
	ObserveMCPGatewayDuration(outcome string, seconds float64)
}

type MCPResilientProvider struct {
	inner   MCPGatewayProvider
	health  repositories.IMCPConnectionSecurityRepository
	audit   *AuditWriter
	metrics MCPGatewayMetrics
	mu      sync.Mutex
	states  map[string]*mcpProviderState
}

type mcpProviderState struct {
	semaphore        chan struct{}
	tokens           float64
	lastRefill       time.Time
	failures         int
	openUntil        time.Time
	halfOpenInFlight bool
}

func NewMCPResilientProvider(inner MCPGatewayProvider, connections repositories.IMCPConnectionRepository, audit *AuditWriter, metrics MCPGatewayMetrics) *MCPResilientProvider {
	var health repositories.IMCPConnectionSecurityRepository
	if value, ok := connections.(repositories.IMCPConnectionSecurityRepository); ok {
		health = value
	}
	return &MCPResilientProvider{inner: inner, health: health, audit: audit, metrics: metrics, states: make(map[string]*mcpProviderState)}
}

func (p *MCPResilientProvider) InvokeTool(ctx context.Context, connection *entities.MCPConnection, tool *entities.MCPDiscoveredTool, payload map[string]any, timeout time.Duration) (domainservices.MCPToolCallResult, error) {
	ctx, span := otel.Tracer("openstate.mcp.gateway").Start(ctx, "mcp.provider.invoke")
	defer span.End()
	if p.inner == nil || connection == nil || tool == nil {
		return domainservices.MCPToolCallResult{}, domaincap.NewCapabilityError(domaincap.ErrorKindUnavailable, "capability.gateway_unavailable", "MCP provider is unavailable")
	}
	span.SetAttributes(
		attribute.String("openstate.tenant_id", connection.TenantID),
		attribute.String("openstate.project_id", connection.ProjectID),
		attribute.String("openstate.connection_id", connection.ID),
		attribute.String("openstate.connection_alias", connection.Alias),
		attribute.String("openstate.tool_name", tool.Name),
	)
	policy := mcpResiliencePolicy(connection, timeout)
	state := p.stateFor(connection.ID, policy)
	if !p.allowRate(state, policy) {
		return p.fail(ctx, connection, tool, domaincap.NewCapabilityError(domaincap.ErrorKindRateLimited, "capability.rate_limited", "MCP provider rate limit exceeded"), "rate_limited")
	}
	if !p.enterCircuit(state) {
		return p.fail(ctx, connection, tool, domaincap.NewCapabilityError(domaincap.ErrorKindUnavailable, "capability.circuit_open", "MCP provider circuit is open"), "circuit_open")
	}
	if !p.acquire(ctx, state.semaphore) {
		p.leaveCircuit(state, false)
		return p.fail(ctx, connection, tool, domaincap.NewCapabilityError(domaincap.ErrorKindRateLimited, "capability.concurrency_limited", "MCP provider concurrency limit exceeded"), "concurrency_limited")
	}
	defer func() { <-state.semaphore }()

	options := domainservices.MCPCallOptionsFromContext(ctx)
	maxAttempts := 1
	if options.Idempotent {
		maxAttempts += policy.RetryMax
	}
	if maxAttempts > 6 {
		maxAttempts = 6
	}
	started := time.Now()
	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		callCtx, cancel := context.WithTimeout(ctx, policy.Timeout)
		result, err := p.inner.InvokeTool(callCtx, connection, tool, payload, policy.Timeout)
		cancel()
		if err == nil {
			p.success(ctx, connection, tool, state, time.Since(started))
			return result, nil
		}
		lastErr = err
		if attempt+1 >= maxAttempts || !retryableMCPError(err) {
			break
		}
		select {
		case <-ctx.Done():
			lastErr = ctx.Err()
		case <-time.After(retryBackoff(attempt)):
		}
		if ctx.Err() != nil {
			break
		}
	}
	p.leaveCircuit(state, false)
	p.failure(ctx, connection, tool, state, lastErr, time.Since(started))
	return domainservices.MCPToolCallResult{}, lastErr
}

func (p *MCPResilientProvider) stateFor(connectionID string, policy mcpConnectionResiliencePolicy) *mcpProviderState {
	p.mu.Lock()
	defer p.mu.Unlock()
	state := p.states[connectionID]
	if state == nil {
		state = &mcpProviderState{semaphore: make(chan struct{}, policy.MaxConcurrency), tokens: float64(policy.RateLimitBurst), lastRefill: time.Now()}
		p.states[connectionID] = state
	}
	return state
}

func (p *MCPResilientProvider) allowRate(state *mcpProviderState, policy mcpConnectionResiliencePolicy) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	now := time.Now()
	if elapsed := now.Sub(state.lastRefill).Seconds(); elapsed > 0 {
		state.tokens = minFloat(float64(policy.RateLimitBurst), state.tokens+elapsed*policy.RateLimitPerSecond)
		state.lastRefill = now
	}
	if state.tokens < 1 {
		return false
	}
	state.tokens--
	return true
}

func (p *MCPResilientProvider) enterCircuit(state *mcpProviderState) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	now := time.Now()
	if state.openUntil.IsZero() {
		return true
	}
	if now.Before(state.openUntil) {
		return false
	}
	if state.halfOpenInFlight {
		return false
	}
	state.halfOpenInFlight = true
	return true
}

func (p *MCPResilientProvider) leaveCircuit(state *mcpProviderState, success bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	state.halfOpenInFlight = false
	if success {
		state.failures = 0
		state.openUntil = time.Time{}
	}
}

func (p *MCPResilientProvider) acquire(ctx context.Context, semaphore chan struct{}) bool {
	select {
	case semaphore <- struct{}{}:
		return true
	case <-ctx.Done():
		return false
	}
}

func (p *MCPResilientProvider) success(ctx context.Context, connection *entities.MCPConnection, tool *entities.MCPDiscoveredTool, state *mcpProviderState, duration time.Duration) {
	p.leaveCircuit(state, true)
	p.recordHealth(ctx, connection, entities.MCPHealthHealthy, "", 0, nil)
	p.observe(ctx, connection, tool, "success", duration)
}

func (p *MCPResilientProvider) failure(ctx context.Context, connection *entities.MCPConnection, tool *entities.MCPDiscoveredTool, state *mcpProviderState, err error, duration time.Duration) {
	p.mu.Lock()
	state.failures++
	threshold := connection.CircuitFailureThreshold
	if threshold <= 0 {
		threshold = 5
	}
	opened := state.failures >= threshold
	if opened {
		recovery := connection.CircuitRecoverySeconds
		if recovery <= 0 {
			recovery = 30
		}
		state.openUntil = time.Now().Add(time.Duration(recovery) * time.Second)
	}
	failures := state.failures
	state.halfOpenInFlight = false
	p.mu.Unlock()
	status := entities.MCPHealthUnavailable
	if opened {
		status = entities.MCPHealthCircuitOpen
	}
	p.recordHealth(ctx, connection, status, safeMCPFailureCode(err), failures, openedTime(opened))
	p.observe(ctx, connection, tool, outcomeForMCPError(err), duration)
}

func (p *MCPResilientProvider) fail(ctx context.Context, connection *entities.MCPConnection, tool *entities.MCPDiscoveredTool, err error, outcome string) (domainservices.MCPToolCallResult, error) {
	p.observe(ctx, connection, tool, outcome, 0)
	return domainservices.MCPToolCallResult{}, err
}

func (p *MCPResilientProvider) recordHealth(ctx context.Context, connection *entities.MCPConnection, status entities.MCPConnectionHealthStatus, reason string, failures int, openedAt *time.Time) {
	if p.health == nil {
		return
	}
	_, _ = p.health.RecordHealth(ctx, repositories.MCPConnectionHealthUpdateInput{TenantID: connection.TenantID, ProjectID: connection.ProjectID, ID: connection.ID, HealthStatus: status, HealthReason: optionalSafeReason(reason), LastSuccessAt: successTime(status), ConsecutiveFailures: failures, CircuitOpenedAt: openedAt, Actor: "system"})
}

func (p *MCPResilientProvider) observe(ctx context.Context, connection *entities.MCPConnection, tool *entities.MCPDiscoveredTool, outcome string, duration time.Duration) {
	if p.metrics != nil {
		p.metrics.IncMCPGatewayInvocation(outcome)
		p.metrics.ObserveMCPGatewayDuration(outcome, duration.Seconds())
	}
	options := domainservices.MCPCallOptionsFromContext(ctx)
	if p.audit != nil {
		p.audit.Write(ctx, connection.TenantID, "system", entities.AuditActionMCPGatewayInvocation, "mcp_connection", connection.ID, nil, map[string]any{"projectId": connection.ProjectID, "connectionAlias": connection.Alias, "tool": tool.Name, "outcome": outcome, "durationMs": duration.Milliseconds(), "correlationId": options.CorrelationID}, nil)
	}
}

type mcpConnectionResiliencePolicy struct {
	Timeout            time.Duration
	MaxConcurrency     int
	RateLimitPerSecond float64
	RateLimitBurst     int
	RetryMax           int
}

func mcpResiliencePolicy(connection *entities.MCPConnection, fallback time.Duration) mcpConnectionResiliencePolicy {
	timeout := time.Duration(connection.TimeoutMS) * time.Millisecond
	if timeout <= 0 {
		timeout = fallback
	}
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	maxConcurrency := connection.MaxConcurrency
	if maxConcurrency <= 0 {
		maxConcurrency = 4
	}
	rate := connection.RateLimitPerSecond
	if rate <= 0 {
		rate = 10
	}
	burst := connection.RateLimitBurst
	if burst <= 0 {
		burst = 20
	}
	retries := connection.RetryMax
	if retries < 0 {
		retries = 0
	}
	return mcpConnectionResiliencePolicy{Timeout: timeout, MaxConcurrency: maxConcurrency, RateLimitPerSecond: rate, RateLimitBurst: burst, RetryMax: retries}
}

func retryableMCPError(err error) bool {
	var capabilityErr *domaincap.CapabilityError
	if !errors.As(err, &capabilityErr) {
		return false
	}
	return capabilityErr.Kind == domaincap.ErrorKindTimeout || capabilityErr.Kind == domaincap.ErrorKindUnavailable || capabilityErr.Kind == domaincap.ErrorKindExternal
}

func retryBackoff(attempt int) time.Duration { return time.Duration(25*(attempt+1)) * time.Millisecond }
func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
func openedTime(opened bool) *time.Time {
	if !opened {
		return nil
	}
	now := time.Now()
	return &now
}
func successTime(status entities.MCPConnectionHealthStatus) *time.Time {
	if status != entities.MCPHealthHealthy {
		return nil
	}
	now := time.Now()
	return &now
}
func optionalSafeReason(reason string) *string {
	if strings.TrimSpace(reason) == "" {
		return nil
	}
	reason = strings.TrimSpace(reason)
	if len(reason) > 128 {
		reason = reason[:128]
	}
	return &reason
}

func safeMCPFailureCode(err error) string {
	var capabilityErr *domaincap.CapabilityError
	if errors.As(err, &capabilityErr) && safeMCPCode(capabilityErr.Code) {
		return capabilityErr.Code
	}
	return "capability.provider_failed"
}

func safeMCPCode(code string) bool {
	if len(code) == 0 || len(code) > 96 {
		return false
	}
	for _, r := range code {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') && r != '.' && r != '_' && r != '-' {
			return false
		}
	}
	return strings.HasPrefix(code, "capability.") || strings.HasPrefix(code, "mcp_")
}

func outcomeForMCPError(err error) string {
	var capabilityErr *domaincap.CapabilityError
	if errors.As(err, &capabilityErr) {
		return string(capabilityErr.Kind)
	}
	return "failed"
}
