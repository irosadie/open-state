/**
 * Intent Resolver — percakapan → intent → workflow → initial state.
 *
 * Ini adalah basis untuk tool MCP `resolve_intent` / `get_active_workflow`.
 * Pada implementasi runtime, klasifikasi bahasa dilakukan LLM (dengan
 * bantuan `IntentDefinition.examples` & `description`); di sini kita
 * sediakan logika resolusi deterministik setelah intent teridentifikasi.
 */
import type { IntentDefinition, WorkflowDefinition } from "../types/workflow"
import { INTENT_REGISTRY, getIntentByWorkflow } from "./intent-registry"
import { orderDokterWorkflow } from "./order-dokter"
import { orderMakananWorkflow } from "./order-makanan"
import { padelBookingWorkflow } from "./padel-booking"

export interface ResolvedIntent {
  intent: IntentDefinition
  workflow: WorkflowDefinition
  entryEvent: string
  initialStateId: string | undefined
}

/** Daftar semua workflow (single source of truth untuk resolver) */
export const ALL_WORKFLOWS: WorkflowDefinition[] = [
  padelBookingWorkflow,
  orderMakananWorkflow,
  orderDokterWorkflow,
]

/** Dapatkan workflow dari slug */
export function getWorkflowBySlug(
  slug: string,
): WorkflowDefinition | undefined {
  return ALL_WORKFLOWS.find((w) => w.slug === slug)
}

/**
 * Resolve intent (yang sudah diklasifikasikan LLM) menjadi:
 * workflow + entry event + initial state.
 */
export function resolveIntent(intentId: string): ResolvedIntent | null {
  const intent = INTENT_REGISTRY.intents.find((i) => i.id === intentId)
  if (!intent) return null
  const workflow = getWorkflowBySlug(intent.workflowSlug)
  if (!workflow) return null

  return {
    intent,
    workflow,
    entryEvent: intent.entryEvent ?? "workflow.started",
    initialStateId: workflow.entryNodeId,
  }
}

/**
 * Cari intent yang mungkin cocok dengan teks user, berbasis keyword sederhana.
 * NOTE: pada produksi ini digantikan LLM intent classification (PRD 21).
 * Helper ini untuk dev/testing cepat.
 */
export function resolveIntentFromText(text: string): ResolvedIntent | null {
  const lower = text.toLowerCase()
  for (const intent of INTENT_REGISTRY.intents) {
    const hit = intent.examples?.some((ex) => lower.includes(ex.toLowerCase()))
    if (hit) return resolveIntent(intent.id)
  }
  // fallback keyword
  if (/(padel|lapangan|main\s)/.test(lower))
    return resolveIntent("BOOKING_PADEL")
  if (/(kopi|makan|minum|order|pesan)/.test(lower))
    return resolveIntent("ORDER_MAKANAN")
  if (/(dokter|konsul|kesehatan|antrian)/.test(lower))
    return resolveIntent("ORDER_DOKTER")
  return null
}

export { getIntentByWorkflow }
