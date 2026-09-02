## Context

The secure MCP server already resolves provider routing internally and checks the active state before invoking a logical capability. Its current failure projection is safe but underspecified for an LLM, while event proposals do not carry an idempotency key through the MCP/orchestrator boundary.

## Goals / Non-Goals

**Goals:**

- Make gateway failures unambiguous and fail closed for LLM clients.
- Preserve provider routing confidentiality.
- Make secure workflow start and event retries safe within the existing tenant-scoped idempotency ledger.
- Keep advisory-mode compatibility behavior unchanged.

**Non-Goals:**

- Automatic provider registration or fallback selection.
- Changes to the provider mock protocol or scenario data.
- Removing read-only diagnostic tools from the server.

## Decisions

1. **Use explicit safe failure metadata.** Secure capability errors will include `hardStop: true` and `nextAction: STOP`, while retaining the existing normalized `code`, `kind`, and safe message. Mapping, connection, catalog, and tool failures receive distinct codes so operators can fix configuration without leaking secrets.

2. **Enforce guidance at the MCP contract boundary.** Initialization instructions and secure tool descriptions will explicitly prohibit scope escalation, fallback capability selection, direct provider calls, and transitions after a failed invocation. The gateway authorization check remains the authoritative enforcement layer.

3. **Reuse the existing idempotency ledger.** The orchestrator service will expose backward-compatible idempotent methods used by secure MCP handlers. Outcomes store only an instance/event identifier and scope; retries load the original tenant-scoped resource. Advisory callers keep the legacy method path. The database upsert must update the operation scope when the engine first records the same event key, otherwise a valid MCP retry is misclassified as a cross-operation conflict.

4. **Use the same key for retries.** The secure tool descriptions require a stable key for `start_workflow` and `propose_event`, making a transport retry distinguishable from a new user action.

5. **Keep static mock catalog tools caller-compatible.** A canned provider response does not consume input fields, so a static fixture with an explicitly empty input schema is registered as permissive. This lets the mock receive search context from the gateway while preserving strict schemas for state-changing tools.

6. **Expose state context names verbatim.** Secure lifecycle guidance tells the client to copy `requiredContext` keys exactly into event payloads, including dotted names such as `doctor.specialty`, so guarded transitions do not fail because a model renamed a context field.

## Risks / Trade-offs

- **Risk:** A ledger write can fail after a workflow or event side effect succeeds. → Return the side-effect outcome only when the original operation and ledger record both succeed; surface an unavailable error otherwise so the caller does not assume a safe retry.
- **Risk:** Existing secure clients omit the new keys. → Keep the Go compatibility handlers usable, but advertise and enforce required keys in secure MCP schemas so new clients cannot omit them.
- **Risk:** A model may still attempt an invalid fallback despite instructions. → The active-state authorization check rejects it without contacting the provider, and regression tests cover that behavior.

## Migration Plan

Deploy the server changes, reconnect the secure State MCP client so it receives the updated instructions/tool schemas, then refresh the project provider catalog and bindings separately. Existing idempotency records remain compatible because new records use explicit scopes.
