package capability

import (
	"context"
	"time"
)

// InputSchemaValidator validates an invocation payload against a capability's
// input schema (PRD §62).
type InputSchemaValidator interface {
	Validate(payload map[string]any, schema []byte) error
}

// RateLimiter enforces per-invocation rate limiting (PRD §62, §83).
type RateLimiter interface {
	Allow(ctx context.Context, key string) (bool, error)
}

// ProviderResolver maps a resolved capability to a CapabilityProvider.
// The default implementation falls back to the mock provider (PRD §2064).
type ProviderResolver interface {
	ResolveProvider(res *ResolvedCapability) CapabilityProvider
}

// CapabilityInvoker orchestrates the full security chain and provider
// invocation (PRD §153, §62).
type CapabilityInvoker struct {
	resolver    *CapabilityResolver
	providerRes ProviderResolver
	schema      InputSchemaValidator
	rateLimiter RateLimiter
	idempotency IdempotencyStore
}

// NewCapabilityInvoker builds the invoker with its collaborators.
func NewCapabilityInvoker(
	resolver *CapabilityResolver,
	providerRes ProviderResolver,
	schema InputSchemaValidator,
	rateLimiter RateLimiter,
	idempotency IdempotencyStore,
) *CapabilityInvoker {
	return &CapabilityInvoker{
		resolver:    resolver,
		providerRes: providerRes,
		schema:      schema,
		rateLimiter: rateLimiter,
		idempotency: idempotency,
	}
}

// Execute runs the capability invocation through the security chain and
// provider, returning a normalized result (PRD §62, §153).
func (ci *CapabilityInvoker) Execute(ctx context.Context, inv Invocation) (InvocationResult, error) {
	// 1. resolve capability + authorization (tenant → workflow → state)
	res, err := ci.resolver.Resolve(ctx, inv.TenantID, inv.Name, inv.WorkflowID, inv.StateID)
	if err != nil {
		return InvocationResult{}, err
	}

	// 2. input schema validation
	if ci.schema != nil && len(res.InputSchema) > 0 {
		if verr := ci.schema.Validate(inv.Payload, res.InputSchema); verr != nil {
			return InvocationResult{}, NewCapabilityError(ErrorKindValidation, "capability.validation_failed", "input schema validation failed: "+verr.Error())
		}
	}

	// 3. rate limiting (PRD 62, 83): scope key is tenant + capability so one
	// tenant/capability cannot abuse the provider.
	if ci.rateLimiter != nil {
		key := "tenant:" + inv.TenantID + ":capability:" + inv.CapabilityID
		ok, rerr := ci.rateLimiter.Allow(ctx, key)
		if rerr != nil {
			return InvocationResult{}, rerr
		}
		if !ok {
			return InvocationResult{}, NewCapabilityError(ErrorKindRateLimited, "capability.rate_limited", "rate limit exceeded")
		}
	}

	// 4. idempotency check
	key := buildIdempotencyKey(inv.WorkflowInstanceID, inv.ActionID)
	if ci.idempotency != nil && key != "" {
		if prev, hit, _ := ci.idempotency.Get(ctx, key); hit {
			return *prev, nil
		}
	}

	// 5. resolve provider & invoke (with retry + timeout)
	provider := ci.providerRes.ResolveProvider(res)
	result, err := ci.invokeWithPolicy(ctx, provider, inv)

	// 6. idempotency store on success (side-effecting)
	if err == nil && ci.idempotency != nil && key != "" {
		_ = ci.idempotency.Put(ctx, key, result)
	}

	return result, err
}

func (ci *CapabilityInvoker) invokeWithPolicy(ctx context.Context, provider CapabilityProvider, inv Invocation) (InvocationResult, error) {
	timeout := inv.Policy.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	var lastErr error
	attempts := inv.Policy.MaxRetry + 1
	if attempts < 1 {
		attempts = 1
	}

	for attempt := 0; attempt < attempts; attempt++ {
		callCtx, cancel := context.WithTimeout(ctx, timeout)
		res, err := provider.Invoke(callCtx, inv)
		cancel()

		if err == nil {
			return res, nil
		}

		lastErr = err
		if !retryableError(err, inv.Policy) {
			return InvocationResult{}, err
		}
		// last attempt: return error
		if attempt == attempts-1 {
			break
		}
		if !waitForRetry(ctx, backoff(100*time.Millisecond, attempt)) {
			break
		}
	}

	if lastErr == nil {
		lastErr = NewCapabilityError(ErrorKindExternal, "capability.failed", "capability invocation failed")
	}
	return InvocationResult{}, lastErr
}
