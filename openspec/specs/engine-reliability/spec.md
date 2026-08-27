# engine-reliability Specification

## Purpose
Add reliability to the runtime engine: suspension/resume, optimistic
concurrency, and event idempotency (PRD §30-31, §42-43).

## Requirements

### Requirement: Suspension & resume

The engine SHALL suspend and resume a workflow instance without losing context.

#### Scenario: Suspend

- GIVEN a running workflow instance
- WHEN SuspendWorkflow is called
- THEN status is SUSPENDED
- AND state, context, history, and version are preserved

#### Scenario: Resume

- GIVEN a SUSPENDED instance
- WHEN ResumeWorkflow is called
- THEN it returns to RUNNING/WAITING from the saved state
- AND continues from where it left off

### Requirement: Optimistic concurrency

The engine SHALL detect conflicting concurrent updates via a version counter.

#### Scenario: Conflict

- GIVEN an instance at version N
- WHEN an update with expected version N-1 is applied
- THEN it returns CONFLICT
- AND does not change state

### Requirement: Event idempotency

The engine SHALL deduplicate events by idempotency key.

#### Scenario: Duplicate event

- GIVEN an event with an already-processed idempotency key
- THEN the engine treats it as a no-op
- AND does not apply the transition twice (PRD §30)
