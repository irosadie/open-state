/**
 * Layer persistence draft workflow menggunakan PGlite
 * (PostgreSQL embedded di browser, via WASM).
 *
 * Keunggulan vs localStorage:
 * - Data tersimpan dalam tabel SQL terstruktur (bukan blobs)
 * - Bisa di-migrate ke PostgreSQL server-side nanti
 * - Query-able (list, filter, cari)
 *
 * Ini solusi sementara sebelum backend tersedia — draft tidak hilang
 * saat refresh, dan siap disinkronkan ke backend PostgreSQL.
 */
import { PGlite } from "@electric-sql/pglite"
import type { WorkflowDefinition } from "../types/workflow"

const DB_NAMESPACE = "openstate"

let dbPromise: Promise<PGlite> | null = null

/** Inisialisasi PGlite singleton (lazy) + buat tabel jika belum ada */
export function getDb(): Promise<PGlite> {
  if (!dbPromise) {
    dbPromise = (async () => {
      const db = new PGlite(`idb://${DB_NAMESPACE}`)
      await db.exec(`
        CREATE TABLE IF NOT EXISTS workflow_drafts (
          slug TEXT PRIMARY KEY,
          name TEXT NOT NULL,
          definition JSONB NOT NULL,
          updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
        );
      `)
      return db
    })()
  }
  return dbPromise
}

/** Simpan draft workflow ke PGlite (upsert by slug) */
export async function saveDraft(workflow: WorkflowDefinition): Promise<void> {
  const db = await getDb()
  await db.query(
    `INSERT INTO workflow_drafts (slug, name, definition, updated_at)
     VALUES ($1, $2, $3, now())
     ON CONFLICT (slug) DO UPDATE SET
       name = EXCLUDED.name,
       definition = EXCLUDED.definition,
       updated_at = now()`,
    [workflow.slug, workflow.name, JSON.stringify(workflow)],
  )
}

/** Muat draft berdasarkan slug. Mengembalikan null jika tidak ada. */
export async function loadDraft(
  slug: string,
): Promise<WorkflowDefinition | null> {
  const db = await getDb()
  const result = await db.query<{ definition: string }>(
    "SELECT definition FROM workflow_drafts WHERE slug = $1",
    [slug],
  )
  if (result.rows.length === 0) return null
  const row = result.rows[0]
  if (!row) return null
  return JSON.parse(row.definition) as WorkflowDefinition
}

/** Daftar semua draft (slug + name + updated_at) */
export async function listDrafts(): Promise<
  Array<{ slug: string; name: string; updatedAt: string }>
> {
  const db = await getDb()
  const result = await db.query<{
    slug: string
    name: string
    updated_at: string
  }>(
    "SELECT slug, name, updated_at FROM workflow_drafts ORDER BY updated_at DESC",
  )
  return result.rows.map((r) => ({
    slug: r.slug,
    name: r.name,
    updatedAt: r.updated_at,
  }))
}

/** Hapus draft berdasarkan slug */
export async function deleteDraft(slug: string): Promise<void> {
  const db = await getDb()
  await db.query("DELETE FROM workflow_drafts WHERE slug = $1", [slug])
}
