# ADR-001: Persistence Strategy — Postgres-first + Repository Abstraction

- **Status:** Accepted
- **Date:** 2026-08-27
- **Context:** PRD, Epic #1, Issue #3

## Context
The open-source platform wants to give users the freedom to choose their
database (PostgreSQL, MySQL, SQLite, MongoDB, etc.). However, production-grade
features (outbox pattern, JSONB query, optimistic concurrency, transactional)
are best supported by PostgreSQL.

Challenge: supporting many DBs at once from the start forces features to be
lowered to the "lowest common denominator", increasing cost and bug risk.

## Decision
**PostgreSQL as the primary database (default & production-grade), with the
entire engine built on top of a repository abstraction (persistence port).**

```text
Runtime Engine (domain-pure, DB-agnostic)
        │  interface: WorkflowRepo, InstanceRepo, EventRepo, ContextRepo
        ▼
Persistence Ports (Go interfaces)
   ├── PostgresAdapter        ← first & primary implementation
   ├── MySQLAdapter           ← (added later, optional)
   ├── SQLiteAdapter          ← (added later, optional)
   └── MongoDBAdapter         ← (added later, optional)
```

## Consequences
### Positive
- Engine depends 100% on interfaces → portable & testable.
- Production-grade features (outbox, JSONB, optimistic locking) can be fully
  used on PostgreSQL.
- Other DB adapters (MySQL/SQLite/Mongo) can be added **without reworking the
  engine** — this is what makes "user can choose DB" possible.
- PRD 72 principle ("domain logic independent from DB") is satisfied.

### Negative / trade-off
- Non-Postgres adapters only implement the subset of features available on that
  DB; rich Postgres features (e.g., robust outbox, advisory lock) are not fully
  available on other adapters.
- Requires discipline so queries stay within portable bounds or are encapsulated
  per adapter.

## Alternatives Considered
- **Multi-DB from the start** (4 adapters at once): rejected because it is
  expensive, risky, and forces features down to the most common denominator.
- **SQLite-first + Postgres prod**: rejected because SQLite does not support the
  production features (concurrency, robust outbox) required by PRD 65, 68.

## References
- PRD section 68 (Database), 72 (Go Backend), 175 (modular monolith)
- Epic #1, Issue #3 (Data & Persistence)
