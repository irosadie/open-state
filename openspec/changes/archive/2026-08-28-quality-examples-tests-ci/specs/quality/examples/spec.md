# quality/examples Specification

## Purpose

Define reproducible, seedable example workflows and their intent-registry entries so
the platform can be exercised end-to-end with real data (PRD §40.1, §123). The
examples cover three business intents: PADEL court booking, food ordering (Order
Makanan), and doctor appointment (Order Dokter).

## ADDED Requirements

### Requirement: Example workflows

The platform SHALL ship at least three example workflows that can be seeded into a
tenant's project.

#### Scenario: Seed the padel workflow

- **GIVEN** a demo tenant/project
- **WHEN** the seed runs
- **THEN** a `padel-court-booking` workflow is registered with entry triggers
  (`padel.booking.requested` / `booking.padel.started`)
- **AND** it reflects the canonical definition at `docs/padel-booking.workflow.json`

#### Scenario: Seed the food ordering workflow

- **GIVEN** a demo tenant/project
- **WHEN** the seed runs
- **THEN** an `order-food` workflow is registered and executable end-to-end

#### Scenario: Seed the doctor ordering workflow

- **GIVEN** a demo tenant/project
- **WHEN** the seed runs
- **THEN** an `order-doctor` workflow is registered and executable end-to-end

### Requirement: Intent registry entries

The platform SHALL register intent entries that map each example workflow to an
intent id (PRD §40.1).

#### Scenario: Intent to workflow mapping

- **WHEN** the seed runs
- **THEN** `BOOKING_PADEL` maps to `padel-court-booking`
- **AND** `ORDER_FOOD` maps to `order-food`
- **AND** `ORDER_DOCTOR` maps to `order-doctor`
- **AND** each intent carries sample `examples` phrases for classification and
  testing (PRD 125)

### Requirement: Idempotent seeding

The seed SHALL be idempotent so it can run repeatedly without duplicating rows.

#### Scenario: Re-running the seed

- **GIVEN** the seed has already run once
- **WHEN** it runs again
- **THEN** existing workflows and intents are updated (upsert) rather than duplicated
- **AND** no duplicate workflow or intent rows are created

### Requirement: Tenant and project scoping

Seeded examples SHALL be scoped to a dedicated demo tenant and project so they do not
pollute production data.

#### Scenario: Isolated demo scope

- **GIVEN** the seed runs in any environment
- **THEN** the example workflows and intents belong to the demo tenant/project
- **AND** they are invisible to other tenants (PRD §4)
