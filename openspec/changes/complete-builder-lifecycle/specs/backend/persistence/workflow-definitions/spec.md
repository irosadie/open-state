## ADDED Requirements

### Requirement: Durable editable draft definition
The system SHALL persist one complete editable `WorkflowDefinition` draft on the
workflow root, independently of its immutable published version snapshots.

- The draft definition SHALL contain the entire graph envelope, including nodes,
  transitions, guards, policies, triggers, and node positions.
- Creating a workflow draft SHALL persist its initial definition.
- Updating a workflow draft SHALL atomically replace its definition and increment
  the workflow's optimistic-lock version only when the caller's expected version
  matches.
- Published `workflow_versions.definition` snapshots SHALL never be changed by a
  draft update.
- The draft definition SHALL be scoped by the same tenant and project boundaries
  as its workflow root.

#### Scenario: Save a graph draft
- **WHEN** an operator saves a changed graph with the current workflow version
- **THEN** a later read returns the same persisted graph and a workflow version
  incremented by one.

#### Scenario: Stale graph save is rejected
- **WHEN** an operator saves a graph with a stale workflow version
- **THEN** no part of the stored draft is changed and the operation fails with a
  conflict.

#### Scenario: Draft edits preserve published history
- **WHEN** a workflow with published versions receives a later draft update
- **THEN** every existing published version retains its original definition.

### Requirement: Repository access to draft definitions
The workflow-definition persistence contract SHALL support reading and
optimistically updating the complete draft definition for a tenant- and
project-scoped workflow.

#### Scenario: Repository returns a scoped draft
- **WHEN** a consumer requests a workflow by tenant, project, and workflow id
- **THEN** the returned workflow includes its persisted draft definition and no
  workflow outside that scope can be returned.
