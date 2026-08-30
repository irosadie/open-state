## Context

The MCP interface currently receives an intent string but the application treats it as a workflow ID or slug. The seed process logs an intent-to-workflow association without persisting it, while the frontend has a separate static registry. External LLM clients therefore have no backend-owned catalog containing the metadata and examples needed for reliable routing. See `proposal.md` and the delta specs for the intended behavior.

## Goals / Non-Goals

**Goals:**

- Make canonical intent IDs and their workflow mappings durable and tenant/project scoped.
- Give MCP clients a compact catalog with descriptions and example utterances.
- Keep workflow state authoritative in the existing State Engine and expose it through existing runtime tools.
- Preserve clean separation between domain, application, infrastructure, and MCP interface layers.

**Non-Goals:**

- Add an LLM, embedding model, or fuzzy classifier to the backend.
- Make the frontend registry the source of truth.
- Change workflow execution, authorization, or capability filtering.

## Decisions

### Persist intents as a first-class project-owned catalog

Add an `intents` table with a generated database key, a canonical `intent_key` such as `BOOKING_PADEL`, tenant and project scope, display name, description, JSON examples, and a required `workflow_id` mapping. Enforce uniqueness for `(tenant_id, project_id, intent_key)` and use foreign keys to prevent mappings to deleted projects or workflows.

The read model joins the mapped workflow and exposes only published workflows as routable. This keeps draft and archived workflow definitions out of LLM routing while allowing the catalog to remain explicitly managed.

Alternative considered: infer one intent from every workflow row. Rejected because a workflow slug is not a canonical intent contract and provides no durable examples or one-to-many routing metadata.

### Make `list_intents` the LLM discovery step

Register a read-only MCP tool with required `tenant` and `project` arguments. Its stable response contains an `intents` array with `id`, `projectId`, `name`, `description`, `examples`, and `workflowSlug`. The tool returns an empty array for a valid scope with no published mappings and a validation error for incomplete scope.

The MCP tool description explicitly tells the model to use the returned canonical `id` with `resolve_intent`. The LLM remains responsible for interpreting the user utterance; OpenState supplies the bounded choices and validates the selected mapping.

Alternative considered: expose all workflow definitions to the model. Rejected because it leaks implementation details, gives the model no stable intent vocabulary, and conflicts with tenant/project isolation.

### Resolve only catalog keys

Change the intent application service and MCP handler so `resolve_intent` looks up the canonical key inside the requested tenant/project and then returns the mapped workflow. Do not retain fallback resolution by arbitrary workflow ID or slug; this prevents bypassing the intent catalog. The existing `get_active_workflow` and `get_context` tools remain the source for conversation runtime state.

Normalize surrounding whitespace and case for lookup, while always returning the stored canonical casing. Unknown or non-routable keys return the existing not-found tool error shape.

### Keep seed data representative

Extend the demo seed records with descriptions and example utterances, including Indonesian phrases equivalent to “saya mau order lapangan” and “saya mau booking lapangan padel” for `BOOKING_PADEL`. Seed each mapping after its project/workflow exists and make the operation idempotent using the scoped canonical key.

### Keep generated database code synchronized

Add the migration and sqlc queries first, regenerate the checked-in database code with the repository's configured sqlc command, then adapt the domain repository and application service. MCP unit tests will use a fake intent port; repository/integration tests will verify the SQL-backed scope and publication filter.

## Risks / Trade-offs

- [Risk] Existing callers that pass workflow IDs or slugs to `resolve_intent` will stop resolving. → [Mitigation] Document the canonical-ID contract, add a focused not-found regression test, and update the demo/client examples to call `list_intents` first.
- [Risk] A catalog entry can become stale when a workflow is archived or deleted. → [Mitigation] Join/filter on published workflows for reads and use workflow foreign-key deletion semantics.
- [Risk] Example utterances can be incomplete for a real tenant. → [Mitigation] Keep examples editable catalog data and return descriptions so tenants can expand coverage without changing the routing code.
- [Risk] The migration is not applied in an existing environment. → [Mitigation] Include the numbered goose migration in deployment steps and verify startup/seed against a migrated database.

## Migration Plan

1. Deploy the new goose migration and regenerated backend code.
2. Deploy the service with `list_intents` and canonical resolution.
3. Run the idempotent demo seed (or tenant-specific catalog provisioning) to create mappings.
4. Update MCP clients to call `list_intents`, select an `id`, call `resolve_intent`, then use the existing runtime tools.

Rollback is application-compatible at the database level: stop using the new tool and roll back the service binary while leaving the additive `intents` table in place. Reverting the migration is only safe after catalog rows are no longer needed and should be performed through the normal migration rollback process.
