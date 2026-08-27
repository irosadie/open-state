## Overview

Hierarchical context resolution for the runtime engine. Domain-pure, no
DB/HTTP/LLM. Builds on `engine-domain-core`.

## Decisions

### D1. Context keyed by dot-path
Keys like `customer.name`, `booking.date`. Precedence: later scope wins.

### D2. Missing = requiredContext - available
A state declares `requiredContext`; the resolver computes what is absent so the
LLM (via MCP later) asks only for the missing pieces (PRD §36-37).

### D3. Memory vs workflow split
`MemoryContext` (persistent customer data) is separate from
`WorkflowContext` (transient order/booking data). Deleting one must not delete
the other (PRD §24, §43.2).

### D4. Sensitivity flag
Each entry carries `sensitive bool` so PII redaction can be applied later
(Epic #4) without structural change.

## Risks / Notes
- **Precedence rules** must be documented to avoid confusion.
- Keep context package free of DB — values are passed in, not fetched here.
