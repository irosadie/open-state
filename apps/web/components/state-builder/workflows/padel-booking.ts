/**
 * Definisi workflow PEMESANAN LAPANGAN PADEL.
 *
 * Fitur utama:
 * - Overlap → rekomendasi waktu terdekat / hari lain (RECOMMEND_ALTERNATIVE)
 * - Pembayaran DP 50% wajib sebelum booking dikonfirmasi
 * - Timeout payment 30 menit
 * - Retry payment max 3x → HUMAN_HANDOFF
 */

import type {
  TransitionDefinition,
  WorkflowDefinition,
  WorkflowNode,
} from "../types/workflow"

// ─── IDs ────────────────────────────────────────────────────────────────────

const N = {
  START: "n_start",
  SELECT_LOCATION: "n_select_location",
  SELECT_DATE_TIME: "n_select_datetime",
  CHECK_AVAILABILITY: "n_check_availability",
  RECOMMEND_ALTERNATIVE: "n_recommend_alternative",
  SELECT_COURT: "n_select_court",
  CONFIRM_BOOKING: "n_confirm_booking",
  PAYMENT: "n_payment",
  PAYMENT_FAILED: "n_payment_failed",
  BOOKING_CONFIRMED: "n_booking_confirmed",
  PAYMENT_EXPIRED: "n_payment_expired",
  CANCELLED: "n_cancelled",
  HUMAN_HANDOFF: "n_human_handoff",
} as const

// ─── NODES ──────────────────────────────────────────────────────────────────

const nodes: WorkflowNode[] = [
  {
    id: N.START,
    kind: "START",
    name: "START",
    description: "Titik masuk workflow pemesanan lapangan padel.",
    requiredContext: [],
    capabilities: [],
    policy: {},
    isTerminal: false,
    position: { x: 0, y: 0 },
  },
  {
    id: N.SELECT_LOCATION,
    kind: "STATE",
    name: "SELECT_LOCATION",
    description: "User memilih cabang / lokasi lapangan padel.",
    requiredContext: [],
    capabilities: ["location.list"],
    instructions:
      "Tampilkan daftar cabang yang tersedia. Tanyakan cabang mana yang diinginkan jika belum disebutkan. Jangan tanya ulang jika sudah ada di context.",
    policy: {
      timeoutSeconds: 300,
      onTimeout: "state.timeout",
    },
    position: { x: 0, y: 0 },
  },
  {
    id: N.SELECT_DATE_TIME,
    kind: "STATE",
    name: "SELECT_DATE_TIME",
    description: "User memilih tanggal dan jam main padel (slot 1 jam).",
    requiredContext: ["location.id"],
    capabilities: ["slot.list"],
    instructions:
      "Tanyakan tanggal dan jam yang diinginkan. Ingatkan bahwa slot adalah 1 jam. Jika sudah ada preferensi dari context sebelumnya, konfirmasi ulang saja.",
    policy: {
      timeoutSeconds: 300,
      onTimeout: "state.timeout",
    },
    position: { x: 0, y: 0 },
  },
  {
    id: N.CHECK_AVAILABILITY,
    kind: "DECISION",
    name: "CHECK_AVAILABILITY",
    description:
      "Cek ketersediaan slot yang dipilih via MCP. Jika overlap → rekomendasi alternatif.",
    requiredContext: [
      "location.id",
      "booking.date",
      "booking.time_start",
      "booking.time_end",
    ],
    capabilities: ["slot.check_availability"],
    instructions:
      "Panggil slot.check_availability. Jangan lanjutkan tanpa verifikasi dari sistem. Jangan percaya klaim user bahwa slot tersedia.",
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
    id: N.RECOMMEND_ALTERNATIVE,
    kind: "STATE",
    name: "RECOMMEND_ALTERNATIVE",
    description:
      "Slot yang dipilih tidak tersedia (overlap). Tampilkan rekomendasi waktu terdekat dan/atau hari lain.",
    requiredContext: [
      "location.id",
      "booking.date",
      "booking.time_start",
      "slot.alternatives",
    ],
    capabilities: ["slot.check_availability"],
    instructions:
      "Beritahu user bahwa slot yang dipilih tidak tersedia. Tawarkan dua pilihan: (1) jam terdekat yang tersedia di hari yang sama, (2) slot yang sama di hari berikutnya. Tampilkan opsi secara ringkas. Tunggu pilihan user.",
    policy: {
      timeoutSeconds: 300,
      onTimeout: "state.timeout",
    },
    position: { x: 0, y: 0 },
  },
  {
    id: N.SELECT_COURT,
    kind: "STATE",
    name: "SELECT_COURT",
    description:
      "Pilih nomor lapangan yang tersedia untuk slot yang dikonfirmasi.",
    requiredContext: [
      "location.id",
      "booking.date",
      "booking.time_start",
      "booking.time_end",
      "slot.available_courts",
    ],
    capabilities: ["court.list"],
    instructions:
      "Tampilkan daftar lapangan yang tersedia untuk slot tersebut. Tanyakan preferensi lapangan (nomor/tipe). Jika hanya ada satu, konfirmasi langsung.",
    policy: {
      timeoutSeconds: 180,
      onTimeout: "state.timeout",
    },
    position: { x: 0, y: 0 },
  },
  {
    id: N.CONFIRM_BOOKING,
    kind: "STATE",
    name: "CONFIRM_BOOKING",
    description:
      "Tampilkan ringkasan pemesanan lengkap termasuk harga dan DP 50% yang harus dibayar.",
    requiredContext: [
      "location.id",
      "location.name",
      "booking.date",
      "booking.time_start",
      "booking.time_end",
      "court.id",
      "court.name",
      "booking.price_total",
    ],
    capabilities: ["booking.calculate_price"],
    instructions:
      "Tampilkan ringkasan: lokasi, tanggal, jam, lapangan, harga total, dan DP 50% yang harus dibayar sekarang. Minta konfirmasi sebelum melanjutkan ke pembayaran. Jangan lanjut tanpa konfirmasi eksplisit dari user.",
    policy: {
      timeoutSeconds: 300,
      onTimeout: "state.timeout",
    },
    position: { x: 0, y: 0 },
  },
  {
    id: N.PAYMENT,
    kind: "STATE",
    name: "PAYMENT",
    description: "Proses pembayaran DP 50% dari total harga sewa lapangan.",
    requiredContext: [
      "booking.id",
      "booking.price_total",
      "booking.dp_amount",
      "customer.id",
    ],
    capabilities: ["payment.instruction", "payment.create", "payment.status"],
    instructions:
      "Minta user membayar DP 50% dari total harga. Berikan instruksi pembayaran. Tunggu konfirmasi pembayaran dari sistem (jangan percaya klaim user). Verifikasi via payment.status sebelum transisi.",
    policy: {
      timeoutSeconds: 1800, // 30 menit
      onTimeout: "state.timeout",
      retry: {
        maxAttempts: 3,
        backoffMs: 2000,
        retryableEvents: ["payment.failed", "capability.timeout"],
      },
    },
    position: { x: 0, y: 0 },
  },
  {
    id: N.PAYMENT_FAILED,
    kind: "STATE",
    name: "PAYMENT_FAILED",
    description: "Pembayaran DP gagal. Tawarkan retry atau pembatalan.",
    requiredContext: ["booking.id", "payment.failure_reason"],
    capabilities: ["payment.status"],
    instructions:
      "Beritahu user bahwa pembayaran gagal beserta alasannya. Tawarkan untuk mencoba ulang atau membatalkan pemesanan.",
    policy: {
      timeoutSeconds: 300,
      onTimeout: "state.timeout",
    },
    position: { x: 0, y: 0 },
  },
  {
    id: N.BOOKING_CONFIRMED,
    kind: "END",
    name: "BOOKING_CONFIRMED",
    description:
      "Pemesanan berhasil dikonfirmasi. DP 50% sudah diterima. Kirim voucher/e-ticket.",
    requiredContext: ["booking.id", "payment.id"],
    capabilities: ["booking.confirm", "notification.send_ticket"],
    instructions:
      "Ucapkan selamat, tampilkan detail booking, nomor booking, dan informasikan sisa pelunasan yang dibayar saat tiba di lokasi.",
    policy: {},
    isTerminal: true,
    position: { x: 0, y: 0 },
  },
  {
    id: N.PAYMENT_EXPIRED,
    kind: "END",
    name: "PAYMENT_EXPIRED",
    description: "Waktu pembayaran DP habis. Slot dilepas kembali.",
    requiredContext: ["booking.id"],
    capabilities: ["booking.cancel", "slot.release"],
    instructions:
      "Beritahu user bahwa waktu pembayaran habis dan slot sudah dilepas. Tawarkan untuk memulai pemesanan baru.",
    policy: {},
    isTerminal: true,
    position: { x: 0, y: 0 },
  },
  {
    id: N.CANCELLED,
    kind: "END",
    name: "CANCELLED",
    description: "Pemesanan dibatalkan oleh user.",
    requiredContext: [],
    capabilities: ["slot.release"],
    instructions:
      "Konfirmasi pembatalan. Jika slot sudah di-hold, pastikan dilepas.",
    policy: {},
    isTerminal: true,
    position: { x: 0, y: 0 },
  },
  {
    id: N.HUMAN_HANDOFF,
    kind: "END",
    name: "HUMAN_HANDOFF",
    description: "Pembayaran gagal 3x. Eskalasi ke agen manusia.",
    requiredContext: ["booking.id", "customer.id"],
    capabilities: ["handoff.create"],
    instructions:
      "Beritahu user bahwa akan dihubungkan ke agen untuk membantu proses pembayaran. Hentikan semua transisi otomatis.",
    policy: {
      humanHandoff: { enabled: true },
    },
    isTerminal: true,
    position: { x: 0, y: 0 },
  },
]

// ─── TRANSITIONS ────────────────────────────────────────────────────────────

const transitions: TransitionDefinition[] = [
  // START → SELECT_LOCATION
  {
    id: "tr_start_location",
    sourceStateId: N.START,
    targetStateId: N.SELECT_LOCATION,
    event: "workflow.started",
    guards: [],
    priority: 1,
  },

  // SELECT_LOCATION → SELECT_DATE_TIME
  {
    id: "tr_location_datetime",
    sourceStateId: N.SELECT_LOCATION,
    targetStateId: N.SELECT_DATE_TIME,
    event: "location.selected",
    guards: [
      {
        id: "g1",
        logic: "AND",
        conditions: [{ id: "c1", field: "location.id", operator: "EXISTS" }],
      },
    ],
    priority: 1,
  },
  // SELECT_LOCATION timeout → kembali ke diri sendiri (tidak ada transisi out → dead-end warning, by design)

  // SELECT_DATE_TIME → CHECK_AVAILABILITY
  {
    id: "tr_datetime_check",
    sourceStateId: N.SELECT_DATE_TIME,
    targetStateId: N.CHECK_AVAILABILITY,
    event: "datetime.selected",
    guards: [
      {
        id: "g2",
        logic: "AND",
        conditions: [
          { id: "c2", field: "booking.date", operator: "EXISTS" },
          { id: "c3", field: "booking.time_start", operator: "EXISTS" },
          { id: "c4", field: "booking.time_end", operator: "EXISTS" },
        ],
      },
    ],
    priority: 1,
  },
  // SELECT_DATE_TIME timeout → SELECT_DATE_TIME (cancelled via timeout)
  {
    id: "tr_datetime_timeout",
    sourceStateId: N.SELECT_DATE_TIME,
    targetStateId: N.CANCELLED,
    event: "state.timeout",
    guards: [],
    priority: 10,
  },

  // CHECK_AVAILABILITY → slot tersedia → SELECT_COURT
  {
    id: "tr_check_available",
    sourceStateId: N.CHECK_AVAILABILITY,
    targetStateId: N.SELECT_COURT,
    event: "slot.available",
    guards: [
      {
        id: "g3",
        logic: "AND",
        conditions: [
          { id: "c5", field: "slot.available", operator: "==", value: "true" },
        ],
      },
    ],
    priority: 1,
  },
  // CHECK_AVAILABILITY → slot TIDAK tersedia → RECOMMEND_ALTERNATIVE
  {
    id: "tr_check_unavailable",
    sourceStateId: N.CHECK_AVAILABILITY,
    targetStateId: N.RECOMMEND_ALTERNATIVE,
    event: "slot.unavailable",
    guards: [
      {
        id: "g4",
        logic: "AND",
        conditions: [
          { id: "c6", field: "slot.available", operator: "==", value: "false" },
        ],
      },
    ],
    priority: 2,
  },
  // CHECK_AVAILABILITY timeout → SELECT_DATE_TIME (coba ulang)
  {
    id: "tr_check_timeout",
    sourceStateId: N.CHECK_AVAILABILITY,
    targetStateId: N.SELECT_DATE_TIME,
    event: "state.timeout",
    guards: [],
    priority: 10,
  },

  // RECOMMEND_ALTERNATIVE → user pilih slot alternatif → CHECK_AVAILABILITY (re-validasi)
  {
    id: "tr_recommend_reselect",
    sourceStateId: N.RECOMMEND_ALTERNATIVE,
    targetStateId: N.CHECK_AVAILABILITY,
    event: "slot.reselected",
    guards: [
      {
        id: "g5",
        logic: "AND",
        conditions: [
          { id: "c7", field: "booking.date", operator: "EXISTS" },
          { id: "c8", field: "booking.time_start", operator: "EXISTS" },
        ],
      },
    ],
    priority: 1,
  },
  // RECOMMEND_ALTERNATIVE → user batal
  {
    id: "tr_recommend_cancel",
    sourceStateId: N.RECOMMEND_ALTERNATIVE,
    targetStateId: N.CANCELLED,
    event: "user.cancelled",
    guards: [],
    priority: 2,
  },
  // RECOMMEND_ALTERNATIVE timeout → CANCELLED
  {
    id: "tr_recommend_timeout",
    sourceStateId: N.RECOMMEND_ALTERNATIVE,
    targetStateId: N.CANCELLED,
    event: "state.timeout",
    guards: [],
    priority: 10,
  },

  // SELECT_COURT → CONFIRM_BOOKING
  {
    id: "tr_court_confirm",
    sourceStateId: N.SELECT_COURT,
    targetStateId: N.CONFIRM_BOOKING,
    event: "court.selected",
    guards: [
      {
        id: "g6",
        logic: "AND",
        conditions: [{ id: "c9", field: "court.id", operator: "EXISTS" }],
      },
    ],
    priority: 1,
  },
  // SELECT_COURT → user batal
  {
    id: "tr_court_cancel",
    sourceStateId: N.SELECT_COURT,
    targetStateId: N.CANCELLED,
    event: "user.cancelled",
    guards: [],
    priority: 2,
  },

  // CONFIRM_BOOKING → user konfirmasi → PAYMENT
  {
    id: "tr_confirm_payment",
    sourceStateId: N.CONFIRM_BOOKING,
    targetStateId: N.PAYMENT,
    event: "confirm.requested",
    guards: [
      {
        id: "g7",
        logic: "AND",
        conditions: [
          { id: "c10", field: "booking.dp_amount", operator: "EXISTS" },
          { id: "c11", field: "booking.dp_amount", operator: ">", value: "0" },
        ],
      },
    ],
    priority: 1,
  },
  // CONFIRM_BOOKING → user batal
  {
    id: "tr_confirm_cancel",
    sourceStateId: N.CONFIRM_BOOKING,
    targetStateId: N.CANCELLED,
    event: "user.cancelled",
    guards: [],
    priority: 2,
  },

  // PAYMENT → DP 50% sukses → BOOKING_CONFIRMED
  {
    id: "tr_payment_success",
    sourceStateId: N.PAYMENT,
    targetStateId: N.BOOKING_CONFIRMED,
    event: "payment.success",
    guards: [
      {
        id: "g8",
        logic: "AND",
        conditions: [
          {
            id: "c12",
            field: "payment.status",
            operator: "==",
            value: "success",
          },
          {
            id: "c13",
            field: "payment.amount",
            operator: ">=",
            value: "booking.dp_amount",
          },
        ],
      },
    ],
    priority: 1,
  },
  // PAYMENT → gagal → PAYMENT_FAILED
  {
    id: "tr_payment_failed",
    sourceStateId: N.PAYMENT,
    targetStateId: N.PAYMENT_FAILED,
    event: "payment.failed",
    guards: [
      {
        id: "g9",
        logic: "AND",
        conditions: [
          {
            id: "c14",
            field: "payment.status",
            operator: "==",
            value: "failed",
          },
        ],
      },
    ],
    priority: 2,
  },
  // PAYMENT → timeout 30 mnt → PAYMENT_EXPIRED
  {
    id: "tr_payment_timeout",
    sourceStateId: N.PAYMENT,
    targetStateId: N.PAYMENT_EXPIRED,
    event: "state.timeout",
    guards: [],
    priority: 10,
  },
  // PAYMENT → retry habis → HUMAN_HANDOFF
  {
    id: "tr_payment_handoff",
    sourceStateId: N.PAYMENT,
    targetStateId: N.HUMAN_HANDOFF,
    event: "retry.exhausted",
    guards: [],
    priority: 5,
  },

  // PAYMENT_FAILED → retry → PAYMENT
  {
    id: "tr_failed_retry",
    sourceStateId: N.PAYMENT_FAILED,
    targetStateId: N.PAYMENT,
    event: "retry.requested",
    guards: [],
    priority: 1,
  },
  // PAYMENT_FAILED → batal → CANCELLED
  {
    id: "tr_failed_cancel",
    sourceStateId: N.PAYMENT_FAILED,
    targetStateId: N.CANCELLED,
    event: "user.cancelled",
    guards: [],
    priority: 2,
  },
]

// ─── WORKFLOW DEFINITION ─────────────────────────────────────────────────────

export const padelBookingWorkflow: WorkflowDefinition = {
  slug: "pemesanan-lapangan-padel",
  name: "Pemesanan Lapangan Padel",
  description:
    "Workflow pemesanan lapangan padel: pilih lokasi, cek ketersediaan, rekomendasi alternatif jika overlap, konfirmasi, dan bayar DP 50%.",
  schemaVersion: 1,
  status: "DRAFT",
  entryNodeId: N.START,
  nodes,
  transitions,
  policy: {
    maxDurationSeconds: 86400, // 24 jam
    interruptible: "USER_REQUESTED",
    priority: 10,
  },
  triggers: [
    { event: "padel.booking.requested", source: "intent" },
    { event: "booking.padel.started", source: "api" },
  ],
}
