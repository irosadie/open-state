/**
 * Legacy browser bridge for the State Builder (PRD 128, epic #5).
 *
 * The server-side workflow draft is now authoritative. The draft key is read
 * only as an explicit migration fallback for work created before API drafts;
 * API identity/version keys remain a small browser convenience for reopening
 * the last workflow.
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

export function loadDraftLocal(): WorkflowDefinition | null {
  return safeParse<WorkflowDefinition>(localStorage.getItem(DRAFT_KEY))
}

/** Remove a legacy local draft after an explicit, successful migration. */
export function clearDraftLocal(): void {
  try {
    localStorage.removeItem(DRAFT_KEY)
  } catch {
    // ignore
  }
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
