## MODIFIED Requirements

### Requirement: State Builder persists drafts to the API
The system SHALL let the State Builder save and load workflow definition drafts through
the Builder API instead of browser-local PGlite or localStorage.

- Initial save SHALL create the workflow and persist the complete graph; subsequent
  saves SHALL update the same server-side draft with its optimistic version.
- Opening a saved workflow by its builder route identifier SHALL fetch and
  materialize the server-side draft into React Flow.
- Autosave and manual save SHALL visibly distinguish saving, saved, and failed
  states, and SHALL retain the unsaved in-memory graph after an error.
- A stale-save conflict SHALL tell the operator to reload rather than silently
  overwriting the server draft.
- The State Builder in-memory React Flow state (nodes, edges, undo/redo, validation)
  SHALL remain unchanged.

#### Scenario: Save draft to API
- **WHEN** the operator edits a workflow in the State Builder and auto-save triggers
- **THEN** the complete draft graph is persisted to the Builder API and the save status is reflected in the UI.

#### Scenario: Load draft from API
- **WHEN** the State Builder opens with a persisted workflow id
- **THEN** the server draft is fetched and materialized into nodes and edges.

#### Scenario: Save conflict
- **WHEN** another operator has saved the same workflow after the current editor loaded it
- **THEN** the current editor sees a conflict state and its local graph is not silently discarded.

### Requirement: React-query hooks for workflows
The system SHALL expose react-query hooks for workflow list, get, create/upsert draft,
publish, version detail, version history, and version comparison.

- Hooks SHALL call the API through the axios instance, never directly inside JSX.
- Mutations SHALL invalidate workflow and version queries affected by a successful
  save or publish.

#### Scenario: Create-or-update draft hook
- **WHEN** the create-or-update draft mutation resolves successfully
- **THEN** the workflow list and affected workflow query are invalidated so downstream views reflect the change.

#### Scenario: Publish hook
- **WHEN** a publish mutation resolves successfully
- **THEN** the workflow, version-history, and version-comparison queries for that workflow are invalidated.

## ADDED Requirements

### Requirement: Publish control in State Builder
The State Builder SHALL provide a Publish action for the currently open workflow.

- The control SHALL be disabled while saving or publishing and SHALL make the
  pending draft save complete before publishing.
- The control SHALL require a locally valid graph before calling the API and SHALL
  display server validation or concurrency errors returned by the API.
- On success, it SHALL display the new published version and refresh version history.

#### Scenario: Publish valid saved draft
- **WHEN** an operator activates Publish for a locally valid, saved workflow
- **THEN** the Builder publishes the current server draft and reports the new version number.

#### Scenario: Block invalid local graph
- **WHEN** an operator activates Publish while local validation has errors
- **THEN** no publish request is sent and the Builder directs the operator to the validation errors.

### Requirement: Version history and graph diff in State Builder
The State Builder SHALL let an operator inspect a workflow's immutable published
versions and compare any two distinct versions.

- History SHALL show version number, published status, current-version marker, and
  publication time in newest-first order.
- The operator SHALL be able to choose base and target versions from that history.
- The diff view SHALL separately present added, removed, and changed nodes and
  transitions, including changed field names.
- The history and diff controls SHALL not allow mutation or restoration of a
  published version.

#### Scenario: Inspect history
- **WHEN** an operator opens version history for a workflow with published versions
- **THEN** the Builder displays every version newest first and identifies the current version.

#### Scenario: Compare selected versions
- **WHEN** an operator selects two distinct published versions
- **THEN** the Builder displays the graph diff returned by the API for that ordered pair.

#### Scenario: Workflow has no published versions
- **WHEN** an operator opens version history for a draft-only workflow
- **THEN** the Builder clearly indicates that no published versions exist and offers no comparison.
