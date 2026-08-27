/**
 * Export / Import workflow definition ke/dari JSON.
 *
 * Sesuai PRD section 120:
 * - Format harus versioned
 * - Menyertakan: workflow definition, states, transitions, guards, policies,
 *   capability references, schema version
 * - TIDAK menyertakan: secrets, credentials, tenant tokens, runtime data
 */
import type { WorkflowDefinition } from "../types/workflow"
import { WORKFLOW_SCHEMA_VERSION } from "../types/workflow.utils"

/** Format file export workflow */
export const EXPORT_FORMAT = "application/json"
/** Ekstensi & nama file export */
export const EXPORT_FILENAME_PREFIX = "workflow"

export interface WorkflowExportEnvelope {
  format: "openstate.workflow"
  version: number
  exportedAt: string
  workflow: WorkflowDefinition
}

/** Bangun envelope export dari definisi workflow */
export function buildWorkflowExport(
  workflow: WorkflowDefinition,
): WorkflowExportEnvelope {
  return {
    format: "openstate.workflow",
    version: WORKFLOW_SCHEMA_VERSION,
    exportedAt: new Date().toISOString(),
    // clone untuk mencegah mutasi data asli di store
    workflow: JSON.parse(JSON.stringify(workflow)),
  }
}

/** Serialisasi workflow ke string JSON (pretty-print) */
export function serializeWorkflow(workflow: WorkflowDefinition): string {
  const envelope = buildWorkflowExport(workflow)
  return JSON.stringify(envelope, null, 2)
}

/** Nama file export, aman untuk filesystem */
export function exportFilename(workflow: WorkflowDefinition): string {
  const slug = workflow.slug.replace(/[^a-z0-9-_]+/gi, "-") || "workflow"
  return `${EXPORT_FILENAME_PREFIX}-${slug}.json`
}

/** Download workflow sebagai file JSON di browser */
export function downloadWorkflow(workflow: WorkflowDefinition): void {
  const json = serializeWorkflow(workflow)
  const blob = new Blob([json], { type: EXPORT_FORMAT })
  const url = URL.createObjectURL(blob)
  const a = document.createElement("a")
  a.href = url
  a.download = exportFilename(workflow)
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)
}

/**
 * Parse & validasi file export JSON.
 * Mengembalikan WorkflowDefinition jika valid, atau melempar Error.
 */
export function parseWorkflowExport(text: string): {
  workflow: WorkflowDefinition
  envelope: WorkflowExportEnvelope
} {
  const parsed = JSON.parse(text) as WorkflowExportEnvelope

  if (!parsed || parsed.format !== "openstate.workflow") {
    throw new Error("Format file tidak dikenali. Bukan export workflow.")
  }
  if (!parsed.workflow || !Array.isArray(parsed.workflow.nodes)) {
    throw new Error("Struktur workflow tidak valid.")
  }

  return { workflow: parsed.workflow, envelope: parsed }
}
