/**
 * Definisi workflow ORDER DOKTER (konsultasi).
 *
 * Mencakup 2 skenario:
 *
 * Skenario A (jadwal tidak sesuai + antrian penuh + cancel):
 *   - User minta konsul dokter Raffi, hari ini jam 10
 *   - Sistem: dokter Raffi masuk jam 4 → OFFER_ALTERNATIVE_TIME
 *   - User: habis maghrib → sistem cek → FULL BOOKED → tawarkan alternatif
 *   - User: cancel → CANCELLED
 *
 * Skenario B (rekomendasi dokter berdasar kebutuhan + ganti dokter):
 *   - User minta dokter gizi → RECOMMEND_DOCTOR (5 dokter)
 *   - User: "rekom utk anak ngga mau makan" → sistem rekom A, B, C
 *   - User pilih A → jam 8 malam → user mau ganti → dokter B jam 3-5
 *   - User pilih jam 4.30 → cek antrian ok → BOOKING_CONFIRMED
 */

import type {
  TransitionDefinition,
  WorkflowDefinition,
  WorkflowNode,
} from "../types/workflow"

// ─── IDs ────────────────────────────────────────────────────────────────────

const N = {
  START: "n_start",
  SELECT_SPECIALTY: "n_select_specialty",
  CHECK_AVAILABILITY: "n_check_availability",
  RECOMMEND_DOCTOR: "n_recommend_doctor",
  OFFER_ALTERNATIVE_TIME: "n_offer_alternative_time",
  SELECT_TIME_SLOT: "n_select_time_slot",
  BOOKING_CONFIRMED: "n_booking_confirmed",
  CANCELLED: "n_cancelled",
} as const

// ─── NODES ──────────────────────────────────────────────────────────────────

const nodes: WorkflowNode[] = [
  {
    id: N.START,
    kind: "START",
    name: "START",
    description: "Titik masuk workflow konsultasi dokter.",
    requiredContext: [],
    capabilities: [],
    policy: {},
    isTerminal: false,
    position: { x: 0, y: 0 },
  },
  {
    id: N.SELECT_SPECIALTY,
    kind: "STATE",
    name: "SELECT_SPECIALTY",
    description:
      "User menyebut dokter spesifik (misal dokter Raffi) atau kategori (misal dokter gizi).",
    requiredContext: ["doctor.specialty"],
    capabilities: ["doctor.lookup", "catalog.doctor_list"],
    instructions:
      "Terima permintaan user: dokter spesifik atau kategori spesialisasi. Jika belum jelas, tanyakan. Jangan tanya ulang yang sudah ada di context.",
    policy: { timeoutSeconds: 300, onTimeout: "state.timeout" },
    position: { x: 0, y: 0 },
  },
  {
    id: N.CHECK_AVAILABILITY,
    kind: "DECISION",
    name: "CHECK_AVAILABILITY",
    description:
      "Cek jadwal & antrian dokter yang diminta. Mengarahkan ke: slot tersedia, jadwal tak sesuai, atau rekomendasi dokter lain.",
    requiredContext: ["doctor.id", "booking.date"],
    capabilities: ["schedule.check", "queue.check"],
    instructions:
      "Cek jadwal & antrian via sistem (bukan klaim user). Tentukan arah: dokter tersedia, jadwal tidak sesuai permintaan, atau butuh rekomendasi dokter lain.",
    policy: {
      timeoutSeconds: 30,
      onTimeout: "state.timeout",
      retry: {
        maxAttempts: 2,
        backoffMs: 1000,
        retryableEvents: ["capability.timeout"],
      },
    },
    position: { x: 0, y: 0 },
  },
  {
    id: N.RECOMMEND_DOCTOR,
    kind: "STATE",
    name: "RECOMMEND_DOCTOR",
    description:
      "Rekomendasikan dokter sesuai kategori/kebutuhan user (misal dokter anak yang menangani susah makan).",
    requiredContext: ["doctor.specialty", "patient.needs"],
    capabilities: ["doctor.recommend", "catalog.doctor_list"],
    instructions:
      "Jika user minta rekomendasi berdasarkan kebutuhan (misal anak susah makan), berikan daftar dokter yang sesuai. Tampilkan 2-3 opsi. Tunggu pilihan user.",
    policy: { timeoutSeconds: 300, onTimeout: "state.timeout" },
    position: { x: 0, y: 0 },
  },
  {
    id: N.OFFER_ALTERNATIVE_TIME,
    kind: "STATE",
    name: "OFFER_ALTERNATIVE_TIME",
    description:
      "Jadwal yang diminta tidak tersedia / slot penuh. Tawarkan waktu alternatif atau jadwal berikutnya.",
    requiredContext: ["doctor.id", "booking.date", "schedule.alternatives"],
    capabilities: ["schedule.check"],
    instructions:
      "Beritahu user jadwal yang diminta tidak tersedia. Tawarkan alternatif (waktu lain hari itu, atau jadwal berikutnya). Tunggu pilihan: pilih slot, ganti waktu, atau batal.",
    policy: { timeoutSeconds: 300, onTimeout: "state.timeout" },
    position: { x: 0, y: 0 },
  },
  {
    id: N.SELECT_TIME_SLOT,
    kind: "STATE",
    name: "SELECT_TIME_SLOT",
    description:
      "User memilih jam. Sistem cek antrian untuk memastikan slot masih tersedia.",
    requiredContext: ["doctor.id", "booking.date", "booking.time_start"],
    capabilities: ["schedule.check", "queue.check", "booking.reserve"],
    instructions:
      "Tampilkan jam yang tersedia. Saat user memilih, cek antrian/slot terlebih dahulu (jangan percaya klaim). Jika slot tersedia, konfirmasi.",
    policy: { timeoutSeconds: 300, onTimeout: "state.timeout" },
    position: { x: 0, y: 0 },
  },
  {
    id: N.BOOKING_CONFIRMED,
    kind: "END",
    name: "BOOKING_CONFIRMED",
    description: "Antrian konsultasi berhasil dipesan.",
    requiredContext: ["booking.id", "doctor.id", "booking.time_start"],
    capabilities: ["booking.confirm", "notification.send_confirmation"],
    instructions: "Konfirmasi antrian: dokter, tanggal, jam, nomor antrian.",
    policy: {},
    isTerminal: true,
    position: { x: 0, y: 0 },
  },
  {
    id: N.CANCELLED,
    kind: "END",
    name: "CANCELLED",
    description: "Permintaan konsultasi dibatalkan.",
    requiredContext: [],
    capabilities: ["booking.cancel"],
    instructions: "Konfirmasi pembatalan.",
    policy: {},
    isTerminal: true,
    position: { x: 0, y: 0 },
  },
]

// ─── TRANSITIONS ────────────────────────────────────────────────────────────

const transitions: TransitionDefinition[] = [
  // START → SELECT_SPECIALTY
  {
    id: "tr_start_specialty",
    sourceStateId: N.START,
    targetStateId: N.SELECT_SPECIALTY,
    event: "consultation.requested",
    guards: [],
    priority: 1,
  },

  // SELECT_SPECIALTY → CHECK_AVAILABILITY
  {
    id: "tr_specialty_check",
    sourceStateId: N.SELECT_SPECIALTY,
    targetStateId: N.CHECK_AVAILABILITY,
    event: "doctor.requested",
    guards: [
      {
        id: "g1",
        logic: "AND",
        conditions: [
          { id: "c1", field: "doctor.specialty", operator: "EXISTS" },
        ],
      },
    ],
    priority: 1,
  },
  // SELECT_SPECIALTY → CANCELLED
  {
    id: "tr_specialty_cancel",
    sourceStateId: N.SELECT_SPECIALTY,
    targetStateId: N.CANCELLED,
    event: "user.cancelled",
    guards: [],
    priority: 5,
  },
  // SELECT_SPECIALTY timeout → CANCELLED
  {
    id: "tr_specialty_timeout",
    sourceStateId: N.SELECT_SPECIALTY,
    targetStateId: N.CANCELLED,
    event: "state.timeout",
    guards: [],
    priority: 10,
  },

  // CHECK_AVAILABILITY → dokter tersedia → SELECT_TIME_SLOT
  {
    id: "tr_check_available",
    sourceStateId: N.CHECK_AVAILABILITY,
    targetStateId: N.SELECT_TIME_SLOT,
    event: "doctor.available",
    guards: [
      {
        id: "g2",
        logic: "AND",
        conditions: [
          {
            id: "c2",
            field: "schedule.available",
            operator: "==",
            value: "true",
          },
        ],
      },
    ],
    priority: 1,
  },
  // CHECK_AVAILABILITY → jadwal tak sesuai → OFFER_ALTERNATIVE_TIME
  {
    id: "tr_check_mismatch",
    sourceStateId: N.CHECK_AVAILABILITY,
    targetStateId: N.OFFER_ALTERNATIVE_TIME,
    event: "doctor.schedule_mismatch",
    guards: [
      {
        id: "g3",
        logic: "AND",
        conditions: [
          {
            id: "c3",
            field: "schedule.available",
            operator: "==",
            value: "false",
          },
        ],
      },
    ],
    priority: 2,
  },
  // CHECK_AVAILABILITY → butuh rekomendasi (dokter spesifik tidak tersedia/kategori)
  {
    id: "tr_check_recommend",
    sourceStateId: N.CHECK_AVAILABILITY,
    targetStateId: N.RECOMMEND_DOCTOR,
    event: "doctor.unavailable",
    guards: [],
    priority: 3,
  },
  // CHECK_AVAILABILITY → user minta rekomendasi berdasar kebutuhan
  {
    id: "tr_check_recommend_needs",
    sourceStateId: N.CHECK_AVAILABILITY,
    targetStateId: N.RECOMMEND_DOCTOR,
    event: "doctor.recommend",
    guards: [],
    priority: 4,
  },
  // CHECK_AVAILABILITY timeout → CANCELLED
  {
    id: "tr_check_timeout",
    sourceStateId: N.CHECK_AVAILABILITY,
    targetStateId: N.CANCELLED,
    event: "state.timeout",
    guards: [],
    priority: 10,
  },

  // RECOMMEND_DOCTOR → user pilih dokter → CHECK_AVAILABILITY (jadwal dokter baru)
  {
    id: "tr_reco_select",
    sourceStateId: N.RECOMMEND_DOCTOR,
    targetStateId: N.CHECK_AVAILABILITY,
    event: "doctor.selected",
    guards: [
      {
        id: "g4",
        logic: "AND",
        conditions: [{ id: "c4", field: "doctor.id", operator: "EXISTS" }],
      },
    ],
    priority: 1,
  },
  // RECOMMEND_DOCTOR → CANCELLED
  {
    id: "tr_reco_cancel",
    sourceStateId: N.RECOMMEND_DOCTOR,
    targetStateId: N.CANCELLED,
    event: "user.cancelled",
    guards: [],
    priority: 5,
  },
  // RECOMMEND_DOCTOR timeout → CANCELLED
  {
    id: "tr_reco_timeout",
    sourceStateId: N.RECOMMEND_DOCTOR,
    targetStateId: N.CANCELLED,
    event: "state.timeout",
    guards: [],
    priority: 10,
  },

  // OFFER_ALTERNATIVE_TIME → user pilih slot alternatif → re-check → SELECT_TIME_SLOT
  {
    id: "tr_alt_reserve",
    sourceStateId: N.OFFER_ALTERNATIVE_TIME,
    targetStateId: N.SELECT_TIME_SLOT,
    event: "slot.reserved",
    guards: [
      {
        id: "g5",
        logic: "AND",
        conditions: [
          { id: "c5", field: "booking.time_start", operator: "EXISTS" },
        ],
      },
    ],
    priority: 1,
  },
  // OFFER_ALTERNATIVE_TIME → user ganti waktu → SELECT_TIME_SLOT
  {
    id: "tr_alt_change",
    sourceStateId: N.OFFER_ALTERNATIVE_TIME,
    targetStateId: N.SELECT_TIME_SLOT,
    event: "time.change_requested",
    guards: [],
    priority: 2,
  },
  // OFFER_ALTERNATIVE_TIME → CANCELLED
  {
    id: "tr_alt_cancel",
    sourceStateId: N.OFFER_ALTERNATIVE_TIME,
    targetStateId: N.CANCELLED,
    event: "user.cancelled",
    guards: [],
    priority: 5,
  },
  // OFFER_ALTERNATIVE_TIME timeout → CANCELLED
  {
    id: "tr_alt_timeout",
    sourceStateId: N.OFFER_ALTERNATIVE_TIME,
    targetStateId: N.CANCELLED,
    event: "state.timeout",
    guards: [],
    priority: 10,
  },

  // SELECT_TIME_SLOT → slot tersedia → BOOKING_CONFIRMED
  {
    id: "tr_slot_confirm",
    sourceStateId: N.SELECT_TIME_SLOT,
    targetStateId: N.BOOKING_CONFIRMED,
    event: "slot.confirmed",
    guards: [
      {
        id: "g6",
        logic: "AND",
        conditions: [
          { id: "c6", field: "queue.available", operator: "==", value: "true" },
          { id: "c7", field: "booking.time_start", operator: "EXISTS" },
        ],
      },
    ],
    priority: 1,
  },
  // SELECT_TIME_SLOT → slot penuh → OFFER_ALTERNATIVE_TIME
  {
    id: "tr_slot_full",
    sourceStateId: N.SELECT_TIME_SLOT,
    targetStateId: N.OFFER_ALTERNATIVE_TIME,
    event: "slot.full",
    guards: [
      {
        id: "g7",
        logic: "AND",
        conditions: [
          {
            id: "c8",
            field: "queue.available",
            operator: "==",
            value: "false",
          },
        ],
      },
    ],
    priority: 2,
  },
  // SELECT_TIME_SLOT → CANCELLED
  {
    id: "tr_slot_cancel",
    sourceStateId: N.SELECT_TIME_SLOT,
    targetStateId: N.CANCELLED,
    event: "user.cancelled",
    guards: [],
    priority: 5,
  },
  // SELECT_TIME_SLOT timeout → CANCELLED
  {
    id: "tr_slot_timeout",
    sourceStateId: N.SELECT_TIME_SLOT,
    targetStateId: N.CANCELLED,
    event: "state.timeout",
    guards: [],
    priority: 10,
  },
]

// ─── WORKFLOW DEFINITION ─────────────────────────────────────────────────────

export const orderDokterWorkflow: WorkflowDefinition = {
  slug: "order-dokter",
  name: "Order Dokter (Konsultasi)",
  description:
    "Workflow konsultasi dokter: pilih spesialisasi/dokter, cek jadwal & antrian, rekomendasi dokter sesuai kebutuhan, tawarkan waktu alternatif saat slot penuh/jadwal tak sesuai, dan dukung ganti dokter/waktu di tengah proses.",
  schemaVersion: 1,
  status: "DRAFT",
  entryNodeId: N.START,
  nodes,
  transitions,
  policy: {
    maxDurationSeconds: 7200, // 2 jam
    interruptible: "USER_REQUESTED",
    priority: 10,
  },
  triggers: [
    { event: "consultation.requested", source: "intent" },
    { event: "doctor.consultation.started", source: "api" },
  ],
}
