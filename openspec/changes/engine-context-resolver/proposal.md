## Why

The engine core (Proposal `engine-domain-core`) executes state machines, but a
conversation also needs **hierarchical context resolution** — which data is
available, which is missing, and distinguishing persistent memory from
workflow/transient data (PRD §23, §24). This drives the "do not ask for known
context" behavior (PRD §37) and the MCP `get_context` tool later (Epic #4).

## What Changes

- **New Go package** `apps/api/internal/domain/engine/context.go` — context
  resolver (domain-pure).
- **Context hierarchy**: tenant → conversation → workflow → state → turn →
  RAG → MCP results (PRD §23).
- **Available vs missing**: computes `missing_context` from a state's
  `requiredContext` (PRD §36).
- **Memory vs workflow data split**: persistent memory (customer) kept separate
  from transient workflow data (order/booking) (PRD §24, §43.2).
- **Redaction hook**: marks sensitive keys for later PII redaction (implementation
  in Epic #4, but the structure is defined here).
- **Unit tests**.

## Capabilities

### New Capabilities

- `engine/context-resolver`: hierarchical context resolution + missing-context
  detection.
- `engine/context-scope`: persistent-memory vs workflow-data separation.

## Impact

- **`apps/api/internal/domain/engine/context.go`** — new file (+ tests).
- No DB/HTTP/LLM changes.
- Depends on `engine-domain-core` (types like `WorkflowNode.requiredContext`).

## Non-Goals

- No MCP `get_context` tool (Epic #4).
- No actual PII redaction (Epic #4).
- No persistence of context (Epic #3).
