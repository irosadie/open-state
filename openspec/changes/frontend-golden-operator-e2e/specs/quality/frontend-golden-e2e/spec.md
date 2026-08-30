## Purpose

Provide deterministic browser-level golden journeys that verify State Builder and operator runtime workflows against an isolated, tenant-scoped OpenState stack.

## ADDED Requirements

### Requirement: Browser golden E2E harness is isolated and deterministic

The platform SHALL provide a dedicated browser E2E suite that runs separately from unit/component tests against a disposable local application stack. Each run SHALL use synthetic fixture data, a fresh isolated data scope, and deterministic service responses so repeated runs produce the same semantic outcomes.

#### Scenario: Browser suite runs against an isolated stack

- **WHEN** the frontend golden E2E command runs
- **THEN** it starts or connects only to the designated disposable web, API, database, and cache services
- **AND** it resets and seeds the synthetic fixture scope before executing journeys
- **AND** it does not reuse production data or a prior run's mutations.

#### Scenario: Golden journey is repeated

- **WHEN** the same golden manifest runs repeatedly against freshly seeded fixtures
- **THEN** its expected workflow version, graph checkpoints, runtime lifecycle outcome, and authorization result are identical for every run.

### Requirement: Golden fixtures declare safe semantic checkpoints

The platform SHALL version-control fixture manifests for browser journeys. Each manifest SHALL declare the actor, active tenant, starting resources, user actions, and expected semantic checkpoints. Fixtures and diagnostics SHALL contain only synthetic safe data and SHALL NOT contain credentials, raw prompts/responses, or RAG documents.

#### Scenario: Fixture declares a Builder journey

- **WHEN** the Builder golden manifest is loaded
- **THEN** it identifies the synthetic Editor, tenant, draft graph, intended edit, and expected saved/published/version-diff checkpoints
- **AND** the journey asserts those business outcomes rather than a full DOM or screenshot snapshot.

#### Scenario: Fixture contains disallowed sensitive data

- **WHEN** fixture validation finds a credential, raw provider payload, or raw RAG document value
- **THEN** the suite fails before browser execution
- **AND** the disallowed value is not emitted in test diagnostics.

### Requirement: State Builder golden journey validates lifecycle behavior

The browser suite SHALL provide an Editor golden journey that verifies permission-aware Builder access, deterministic draft edit/save/reload persistence, valid publish, version history, and graph diff behavior. It SHALL also verify invalid publish prevention and stale-save conflict behavior without losing the local graph.

#### Scenario: Editor saves and publishes a valid graph

- **WHEN** a fixture Editor opens the seeded Builder draft, makes the declared valid graph edit, saves it, and publishes it
- **THEN** the UI acknowledges the persisted draft and published version
- **AND** reopening the workflow shows the saved graph
- **AND** history identifies the current published version in newest-first order
- **AND** the selected version pair presents the expected added, removed, and changed graph elements.

#### Scenario: Editor attempts invalid or stale Builder mutation

- **WHEN** the fixture presents an invalid graph for publish or a stale version for save
- **THEN** invalid publish sends no publish request and directs the Editor to validation feedback
- **AND** stale save displays conflict/reload guidance
- **AND** the unsaved local graph remains available to the Editor.

### Requirement: Operator golden journey validates runtime inspection and commands

The browser suite SHALL provide an Operator golden journey that verifies permission-aware runtime discovery, instance detail, safe context/timeline presentation, permitted Debug View presentation, and confirmed lifecycle commands. It SHALL verify the resulting persisted lifecycle state and audit record through non-browser test verification utilities.

#### Scenario: Operator inspects a runtime instance

- **WHEN** a fixture Operator opens the tenant-scoped Runtime Inspector and selects a seeded instance
- **THEN** the UI displays the expected workflow/version, current state, safe available or missing context, and chronological timeline
- **AND** the permitted Debug View displays only fixture source, status, duration, correlation information, reason codes, and sanitized provider metadata.

#### Scenario: Operator issues lifecycle commands

- **WHEN** the fixture Operator confirms suspend for a running instance, resume for a suspended instance, and retry for a failed instance
- **THEN** the UI reports each corresponding lifecycle result
- **AND** test verification confirms the expected persisted state and audit actor/action/outcome for each instance.

### Requirement: Golden journeys prove authorization and tenant boundaries in the browser

The browser suite SHALL verify that a journey actor can access only route and action permissions granted for the active tenant. A denied route or action SHALL present the application's access-denied behavior and SHALL NOT load protected data merely to discover denial.

#### Scenario: Operator opens an unrelated management route

- **WHEN** the fixture Operator navigates directly to a tenant or capability-management route without the required permission
- **THEN** the application presents access-denied feedback
- **AND** the protected route data request is not issued.

#### Scenario: Fixture actor encounters another tenant's resource

- **WHEN** the fixture actor searches for or opens an identifier seeded only in the sentinel tenant
- **THEN** the application does not display that tenant's resource data
- **AND** test verification confirms no cross-tenant lifecycle or audit mutation occurred.

### Requirement: Browser golden tests do not contact external providers

The browser golden suite SHALL restrict external network access to the local disposable test stack. It SHALL use deterministic local fixture references for any integration metadata and SHALL fail if the browser or services attempt to contact a real LLM, RAG, MCP, observability, or capability provider.

#### Scenario: Runtime trace includes a provider reference

- **WHEN** a fixture runtime trace includes an external-provider reference
- **THEN** the UI renders only the sanctioned sanitized metadata from the fixture
- **AND** no external provider network request is made.

### Requirement: Browser golden suite is a CI regression gate

The frontend golden E2E suite SHALL run in a dedicated CI job on relevant application, test-infrastructure, and workflow changes. A failed semantic checkpoint SHALL fail the job. Failure diagnostics SHALL be bounded and contain only approved synthetic fixture data.

#### Scenario: Golden checkpoint regresses in CI

- **WHEN** a Builder or Operator checkpoint differs from its fixture expectation
- **THEN** the CI job fails
- **AND** provides approved diagnostic artifacts sufficient to identify the failed step
- **AND** does not upload production or disallowed sensitive data.
