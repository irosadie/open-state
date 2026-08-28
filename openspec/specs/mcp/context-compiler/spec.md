# mcp/context-compiler Specification

## Purpose
Define the LLM context compiler that assembles the minimal per-turn context for a
3rd-party LLM/RAG client, separating available vs missing context, distinguishing
memory from workflow data, and redacting PII before it leaves the platform (PRD 22, 90).

## Requirements

### Requirement: Minimal per-turn context

The platform SHALL compile a minimal, focused context for a given turn (PRD 22).

#### Scenario: Compile context for a turn

- **WHEN** a client requests context for a conversation/instance turn
- **THEN** the compiler returns only the fields relevant to the current state and
  intent
- **AND** omits redundant or off-scope data

### Requirement: Available vs missing context

The compiled context SHALL distinguish available context from missing/required context.

#### Scenario: Signal gaps

- **WHEN** the engine requires context that is not yet present for the current state
- **THEN** the compiler flags it as missing/required so the client can request it
- **AND** lists the available context separately

### Requirement: Memory vs workflow split

The compiler SHALL keep persistent memory references separate from workflow runtime
data (PRD 24, 43.2).

#### Scenario: No data leaking between scopes

- **WHEN** the compiler builds context
- **THEN** long-lived customer/user memory is presented separately from the current
  workflow instance state, so deleting a workflow instance never affects memory.

### Requirement: PII redaction

The compiler SHALL redact PII before returning context to any client (PRD 90).

#### Scenario: Redacted output

- **WHEN** compiled context contains PII (names, identifiers, contact data, free-text)
- **THEN** the returned context SHALL contain masked/omitted PII according to the
  configured redaction policy.

### Requirement: Redactor port

PII redaction SHALL be pluggable through a portable redactor interface (PRD 90, 169).

#### Scenario: Replaceable redaction

- **WHEN** an operator configures a different redaction implementation
- **THEN** the compiler uses the injected redactor without code change to the compiler
  itself.
