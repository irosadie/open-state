# mcp/rag-provider Specification

## Purpose

Define the `RAGProvider` port so the State Engine can request relevant knowledge from an
external RAG backend without depending on a concrete implementation (PRD 169, 171).

## ADDED Requirements

### Requirement: RAG provider port

The platform SHALL define a `RAGProvider` interface exposing a `Retrieve` operation
(PRD 171).

#### Scenario: Retrieve relevant knowledge

- **WHEN** the engine requires knowledge for a state/context
- **THEN** it calls `Retrieve(ctx, query)` and receives a ranked set of relevant
  documents/chunks.

### Requirement: Portable and injectable

The RAG provider SHALL be injected as a domain port; the engine SHALL NOT depend on a
concrete RAG backend.

#### Scenario: Replaceable RAG backend

- **WHEN** a different RAG backend (vector DB, search API, etc.) is provided
- **THEN** the engine/application layer requires no changes (PRD 169).

### Requirement: Retrieval result

The `Retrieve` result SHALL be a normalized structure the compiler can merge into
context.

#### Scenario: Normalized result

- **WHEN** `Retrieve` returns
- **THEN** the result contains the source text and optional metadata/relevance, usable
  by the context compiler without leaking backend-specific types.

### Requirement: No built-in LLM call

The RAG provider SHALL NOT require a platform-initiated LLM call (PRD 170).

#### Scenario: Knowledge only, no generation

- **WHEN** the engine retrieves knowledge
- **THEN** it returns retrieved text only; text generation stays with the 3rd-party
  LLM client.
