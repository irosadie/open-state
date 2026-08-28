/**
 * Local durability bridge for the State Builder (PRD 128, epic #5).
 *
 * The backend currently persists the full workflow definition only at publish
 * (immutable version); the draft *body* (nodes/graph) has no server column yet.
 * Until that lands, the draft definition and the authoritative API workflow id
 * are mirrored to the browser so an in-progress edit survives a refresh. The
 * API remains the source of truth for workflow identity (create/update/publish).
 */
import type { WorkflowDefinition } from "../types/workflow"

const DRAFT_KEY = "openstate:sb:draft"
const API_ID_KEY = "openstate:sb:apiWorkflowId"
const API_VERSION_KEY = "openstate:sb:apiVersion"

function safeParse<T>(raw: string | null): T | null {
  if (!raw) return null
  try {
    return JSON.parse(raw) as T
  } catch {
    return null
  }
}

export function saveDraftLocal(wf: WorkflowDefinition): void {
  try {
    localStorage.setItem(DRAFT_KEY, JSON.stringify(wf))
  } catch {
    // storage unavailable (e.g. SSR) — non-fatal
  }
}

export function loadDraftLocal(): WorkflowDefinition | null {
  return safeParse<WorkflowDefinition>(localStorage.getItem(DRAFT_KEY))
}

export function saveApiId(id: string | null): void {
  try {
    if (id) localStorage.setItem(API_ID_KEY, id)
    else localStorage.removeItem(API_ID_KEY)
  } catch {
    // ignore
  }
}

export function loadApiId(): string | null {
  return localStorage.getItem(API_ID_KEY)
}

export function saveApiVersion(version: number): void {
  try {
    localStorage.setItem(API_VERSION_KEY, String(version))
  } catch {
    // ignore
  }
}

export function loadApiVersion(): number | null {
  const raw = localStorage.getItem(API_VERSION_KEY)
  if (!raw) return null
  const n = Number(raw)
  return Number.isFinite(n) ? n : null
}
