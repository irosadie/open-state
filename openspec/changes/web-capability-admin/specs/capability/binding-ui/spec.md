# capability/binding-ui Specification

## Purpose

Define the admin UI for managing capability bindings across tenant, workflow, and state
scopes, consuming the binding endpoints of the capability admin API.

## ADDED Requirements

### Requirement: List bindings

The admin UI SHALL display the bindings for a capability.

#### Scenario: View bindings

- **WHEN** an admin opens the bindings panel for a capability
- **THEN** the UI shows each binding's scope type, scope id, and permission

### Requirement: Create binding

The admin UI SHALL let an admin bind a capability to a scope.

#### Scenario: Submit binding

- **WHEN** an admin selects a scope type and scope id and submits
- **THEN** the UI validates the payload and calls the create-binding endpoint
- **AND** surfaces validation or conflict errors inline
- **AND** refreshes the binding list

### Requirement: Delete binding

The admin UI SHALL let an admin remove a binding.

#### Scenario: Remove binding

- **WHEN** an admin confirms removing a binding
- **THEN** the UI calls the delete-binding endpoint
- **AND** removes it from the list
