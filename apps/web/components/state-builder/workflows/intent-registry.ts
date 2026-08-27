/**
 * Intent Registry — mapping intent percakapan → workflow (state machine).
 *
 * Inilah yang akan dikembalikan MCP pada resolusi percakapan:
 * percakapan → intent → workflow → current state.
 *
 * Setiap intent terhubung ke SATU workflow. LLM mengklasifikasikan intent
 * dari bahasa user; engine yang menentukan state machine-nya.
 */
import type { IntentRegistry } from "../types/workflow"
import { orderDokterWorkflow } from "./order-dokter"
import { orderMakananWorkflow } from "./order-makanan"
import { padelBookingWorkflow } from "./padel-booking"

export const INTENT_REGISTRY: IntentRegistry = {
  schemaVersion: 1,
  intents: [
    {
      id: "BOOKING_PADEL",
      name: "Pemesanan Lapangan Padel",
      description:
        "User ingin memesan / booking lapangan padel (atau olahraga lapangan).",
      workflowSlug: padelBookingWorkflow.slug,
      entryEvent: "padel.booking.requested",
      examples: [
        "mau booking lapangan padel",
        "pesan lapangan padel besok jam 7",
        "saya mau main padel, ada slot?",
      ],
      priority: 10,
    },
    {
      id: "ORDER_MAKANAN",
      name: "Order Makanan (Kafe)",
      description: "User ingin memesan makanan/minuman (kafe, kopi, dsb).",
      workflowSlug: orderMakananWorkflow.slug,
      entryEvent: "order.makanan.requested",
      examples: ["saya mau kopi latte", "pesan nasi goreng", "mau order makan"],
      priority: 10,
    },
    {
      id: "ORDER_DOKTER",
      name: "Order Dokter (Konsultasi)",
      description:
        "User ingin konsultasi / buat antrian dokter atau tenaga medis.",
      workflowSlug: orderDokterWorkflow.slug,
      entryEvent: "consultation.requested",
      examples: [
        "konsul sama dokter raffi",
        "buat antrian dokter gizi",
        "mau konsultasi kesehatan",
      ],
      priority: 10,
    },
  ],
}

/** Cari intent berdasarkan id */
export function getIntent(intentId: string) {
  return INTENT_REGISTRY.intents.find((i) => i.id === intentId)
}

/** Cari intent berdasarkan workflow slug */
export function getIntentByWorkflow(slug: string) {
  return INTENT_REGISTRY.intents.find((i) => i.workflowSlug === slug)
}
