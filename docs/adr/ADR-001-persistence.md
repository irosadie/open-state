# ADR-001: Persistence Strategy — Postgres-first + Repository Abstraction

- **Status:** Accepted
- **Tanggal:** 2026-08-27
- **Konteks:** PRD, Epic #1, Issue #3

## Konteks
Platform open-source ingin memberi kebebasan kepada user untuk memilih
database (PostgreSQL, MySQL, SQLite, MongoDB/dll). Namun fitur production-grade
(outbox pattern, JSONB query, optimistic concurrency, transactional) paling
kuat didukung PostgreSQL.

Tantangan: mendukung banyak DB sekaligus sejak awal membuat fitur harus
diturunkan ke "lowest common denominator", menaikkan biaya & risiko bug.

## Keputusan
**PostgreSQL sebagai database utama (default & production-grade), dengan
seluruh engine dibangun di atas abstraksi repository (persistence port).**

```text
Runtime Engine (domain-pure, tidak tahu DB)
        │  interface: WorkflowRepo, InstanceRepo, EventRepo, ContextRepo
        ▼
Persistence Ports (Go interfaces)
   ├── PostgresAdapter        ← implementasi pertama & utama
   ├── MySQLAdapter           ← (ditambahkan kemudian, opsional)
   ├── SQLiteAdapter          ← (ditambahkan kemudian, opsional)
   └── MongoDBAdapter         ← (ditambahkan kemudian, opsional)
```

## Konsekuensi
### Positif
- Engine 100% bergantung pada interface → portabel & testable.
- Fitur production-grade (outbox, JSONB, optimistic locking) dapat dipakai
  penuh di PostgreSQL.
- Adapter DB lain (MySQL/SQLite/Mongo) dapat ditambahkan **tanpa merombak
  engine** — ini yang membuat "user bisa pilih DB" menjadi mungkin.
- Prinsip PRD 72 ("domain logic independent from DB") terpenuhi.

### Negatif / trade-off
- Adapter non-Postgres hanya mengimplementasi subset fitur yang tersedia di
  DB tsb; fitur kaya Postgres (mis. outbox robust, advisory lock) tidak penuh
  di adapter lain.
- Perlu disiplin agar query tetap dalam batas yang portable atau dienkapsulasi
  per adapter.

## Alternatif yang Dipertimbangkan
- **Multi-DB sejak awal** (4 adapter langsung): ditolak karena mahal, riskan,
  dan memaksa fitur turun ke paling umum.
- **SQLite-first + Postgres prod**: ditolak karena SQLite tidak mendukung fitur
  production (concurrency, outbox kuat) yang dibutuhkan PRD 65, 68.

## Referensi
- PRD section 68 (Database), 72 (Go Backend), 175 (modular monolith)
- Epic #1, Issue #3 (Data & Persistence)
