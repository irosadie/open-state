## Purpose

Give authorized operators an Admin Console interface for inspecting persisted
workflow execution and sanitized per-turn debug evidence without directly
integrating the browser with AI or provider systems (PRD 142, 143).

## ADDED Requirements

### Requirement: Runtime Inspector navigation and discovery

The Admin Console SHALL provide an authorized Runtime Inspector route where users
can search and filter workflow instances, then open an instance detail view.

- The list SHALL show workflow/version, current state, lifecycle status, last
  activity, and correlation id when available.
- The UI SHALL use OpenState runtime APIs only; it SHALL NOT call LLM, RAG, MCP,
  OTLP, or provider dashboards from the browser.

#### Scenario: Operator opens an instance from the list

- **WHEN** an operator selects an instance from the Runtime Inspector list
- **THEN** the UI opens its detail view with the data returned by OpenState
- **AND** does not establish a browser connection to an external provider

### Requirement: Runtime detail presentation

The instance detail UI SHALL present the pinned workflow/version, current state,
sanitized available and missing context, and ordered runtime timeline.

#### Scenario: Timeline explains current state

- **WHEN** an operator views an instance with recorded events and transitions
- **THEN** the UI renders those entries in chronological order
- **AND** identifies the current state and correlated event activity

### Requirement: Debug View presentation

The detail UI SHALL provide a Debug View that renders the recorded per-turn stages
and their status, source, duration, correlation id, reason codes, and sanitized
provider metadata.

- External LLM/RAG/MCP stages SHALL be visibly labelled as external-provider
  metadata.
- A stage that was not recorded SHALL be displayed as unavailable, not as a
  successful or failed provider result.
- Debug View SHALL be hidden or denied when the signed-in user lacks `debug:read`.

#### Scenario: Trace shows an external RAG reference

- **WHEN** the selected turn contains a recorded RAG integration reference
- **THEN** the UI shows its provider alias, status, duration, and sanitized summary
- **AND** does not display raw retrieved documents or invoke the RAG system

### Requirement: Truthful loading and failure states

The Runtime Inspector SHALL show loading, empty, forbidden, and error states that
do not imply runtime data exists when it could not be retrieved.

#### Scenario: Debug access is denied

- **WHEN** the runtime detail request succeeds but Debug View is forbidden
- **THEN** the UI keeps permitted runtime detail visible
- **AND** clearly states that debug evidence is not authorized
