## Purpose

Let local MCP clients exercise realistic provider write lifecycles and observable read-after-write effects without requiring third-party credentials or persistent infrastructure.

## ADDED Requirements

### Requirement: Provider writes mutate isolated scenario state

The provider mock SHALL apply each declared write tool to the active scenario's in-memory state. State SHALL be isolated by process and reset to the scenario's initial data when the process starts.

#### Scenario: Restart a provider scenario

- **WHEN** a provider mock process is restarted with the same scenario
- **THEN** writes from the prior process are absent and the configured initial state is restored

### Requirement: Padel booking and payment writes are stateful

The padel scenario SHALL create bookings for available slots and create and verify payments only for existing bookings. Padel read tools SHALL expose the resulting booking availability and payment status.

#### Scenario: Book a padel slot and verify payment

- **WHEN** a client books an available padel slot, creates its payment, and verifies that payment
- **THEN** the booking has a deterministic reference, the slot is unavailable, and the payment status is `PAID`

#### Scenario: Attempt a duplicate padel booking

- **WHEN** a client books an already reserved padel slot
- **THEN** the tool returns an MCP error and leaves state unchanged

### Requirement: Food cart, order, and payment writes are stateful

The food-order scenario SHALL add valid menu items to a cart, create an order from the cart, and create and verify payments for that order. Cart and order read tools SHALL return the resulting state.

#### Scenario: Create a food order from a cart

- **WHEN** a client adds menu items to a cart and creates an order
- **THEN** the cart totals and created order reflect the selected items and use deterministic identifiers

#### Scenario: Create an order without cart items

- **WHEN** a client attempts to create an order from an empty cart
- **THEN** the tool returns an MCP error and creates no order

### Requirement: Doctor appointment and payment writes are stateful

The doctor scenario SHALL reserve, confirm, cancel, and directly book valid appointment slots, and SHALL create and verify payments for existing appointments. Appointment and schedule reads SHALL expose the resulting availability and status.

#### Scenario: Reserve and confirm a doctor appointment

- **WHEN** a client reserves an available appointment slot and confirms the reservation
- **THEN** the scenario returns deterministic references and the slot becomes unavailable

#### Scenario: Cancel an appointment

- **WHEN** a client cancels an existing confirmed appointment
- **THEN** the appointment becomes cancelled and its slot becomes available again

### Requirement: Write errors are MCP-visible

The provider mock SHALL return MCP tool errors for missing or invalid input, missing referenced records, unavailable slots, and invalid lifecycle transitions. Failed writes SHALL NOT partially mutate state.

#### Scenario: Verify an unknown payment

- **WHEN** a client requests payment verification for an unknown payment identifier
- **THEN** the tool returns an MCP error and creates no payment record
