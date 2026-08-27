# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- **Intent Registry** — formal mapping of conversation → intent → workflow →
  state resolution (PRD §40.1).
- **Intent Resolver** — dev/testing intent classification helper.
- **Example workflows**:
  - `order-makanan` (food ordering with stock check, recommendation,
    mid-flow product change)
  - `order-dokter` (doctor consultation with schedule/queue check,
    needs-based recommendation, doctor switch)
- **ADR-001** — Postgres-first + repository abstraction (persistence strategy).
- **PRD updates**:
  - §37.1 Less-Click / Minimal Friction
  - §38.1 Item Recommendation
  - §38.2 Schedule Mismatch & Queue
  - §38.3 Needs-Based Recommendation & Provider Switch
  - §40.1 Intent Registry
  - §43.1 Mid-Flow Interruption
  - §43.2 Context Preservation & Resume
- **Documentation** — rewritten README, CONTRIBUTING, SECURITY, CODE_OF_CONDUCT.

### Changed
- State Builder now supports selecting example workflows via dropdown.

## [0.1.0] - 2026-08-27

### Added
- Monorepo scaffold (Go + Next.js + Turborepo + bun).
- State Builder UI (React Flow): flowchart nodes, transitions, guards,
  auto-layout, validation, import/export JSON.
- PGlite draft persistence (browser-local), undo/redo, search, toasts.
- PADEL booking example workflow (overlap recommendation + DP 50%).
- Auth flow (register/login via BFF proxy).
