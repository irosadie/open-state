/**
 * Definisi workflow ORDER MAKANAN (minuman kafe).
 *
 * Skenario yang dicakup:
 * - User minta produk (misal: kopi latte)
 * - Produk kosong → sistem merekomendasikan produk lain dalam kategori sama
 *   (misal: aren coffee) via RECOMMEND_ALTERNATIVE
 * - Sistem minta nama & alamat (COLLECT_CUSTOMER) — jangan minta ulang
 *   jika alamat sudah ada di context (PRD 37)
 * - Sistem minta payment (PAYMENT)
 * - User interupsi: ganti produk (misal: aren → sanger)
 *   → CHANGE_PRODUCT → cek stok → kembali, alamat sudah ada → lanjut PAYMENT
 *
 * Ini menguji: rekomendasi, context preservation, interupsi/suspend,
 * dan do-not-ask-known-context (PRD 24, 37, 38, 42, 43).
 */

import type {
  TransitionDefinition,
  WorkflowDefinition,
  WorkflowNode,
} from "../types/workflow"

// ─── IDs ────────────────────────────────────────────────────────────────────

const N = {
  START: "n_start",
  SELECT_PRODUCT: "n_select_product",
  CHECK_STOCK: "n_check_stock",
  RECOMMEND_ALTERNATIVE: "n_recommend_alternative",
  COLLECT_CUSTOMER: "n_collect_customer",
  PAYMENT: "n_payment",
  PAYMENT_FAILED: "n_payment_failed",
  CHANGE_PRODUCT: "n_change_product",
  ORDER_CONFIRMED: "n_order_confirmed",
  PAYMENT_EXPIRED: "n_payment_expired",
  CANCELLED: "n_cancelled",
} as const

// ─── NODES ──────────────────────────────────────────────────────────────────

const nodes: WorkflowNode[] = [
  {
    id: N.START,
    kind: "START",
    name: "START",
    description: "Titik masuk workflow order makanan.",
    requiredContext: [],
    capabilities: [],
    policy: {},
    isTerminal: false,
    position: { x: 0, y: 0 },
  },
  {
    id: N.SELECT_PRODUCT,
    kind: "STATE",
    name: "SELECT_PRODUCT",
    description: "User memilih/menyebutkan produk yang ingin dipesan.",
    requiredContext: ["order.items"],
    capabilities: ["catalog.list"],
    instructions:
      "Terima produk yang diminta user. Konfirmasi produk & jumlah. Jika user belum menyebut produk, tanyakan. Jangan tanya ulang produk yang sudah ada di context.",
    policy: { timeoutSeconds: 300, onTimeout: "state.timeout" },
    position: { x: 0, y: 0 },
  },
  {
    id: N.CHECK_STOCK,
    kind: "DECISION",
    name: "CHECK_STOCK",
    description:
      "Cek stok produk via MCP. Jika kosong → rekomendasi alternatif kategori sama.",
    requiredContext: ["order.items", "product.sku"],
    capabilities: ["product.check_stock"],
    instructions:
      "Panggil product.check_stock. JANGAN percaya klaim user bahwa produk tersedia. Verifikasi dari sistem sebelum lanjut.",
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
      "Produk kosong. Tampilkan rekomendasi produk lain dalam kategori yang sama.",
    requiredContext: ["product.category", "product.alternatives"],
    capabilities: ["product.suggest_alternatives"],
    instructions:
      "Beritahu user bahwa produk yang diminta kosong. Tampilkan 1-3 rekomendasi produk dalam kategori yang sama. Tunggu pilihan user.",
    policy: { timeoutSeconds: 300, onTimeout: "state.timeout" },
    position: { x: 0, y: 0 },
  },
  {
    id: N.COLLECT_CUSTOMER,
    kind: "STATE",
    name: "COLLECT_CUSTOMER",
    description:
      "Kumpulkan nama & alamat customer. Jangan minta ulang jika sudah ada di context.",
    requiredContext: ["customer.name", "customer.address"],
    capabilities: ["customer.lookup"],
    instructions:
      "Pastikan nama & alamat lengkap. Jika customer.name atau customer.address SUDAH ada di context (memory/workflow), JANGAN tanya ulang. Hanya minta yang kurang. Verifikasi alamat yang ada.",
    policy: { timeoutSeconds: 300, onTimeout: "state.timeout" },
    position: { x: 0, y: 0 },
  },
  {
    id: N.PAYMENT,
    kind: "STATE",
    name: "PAYMENT",
    description: "Proses pembayaran pesanan.",
    requiredContext: ["order.id", "order.total", "customer.id"],
    capabilities: ["payment.instruction", "payment.create", "payment.status"],
    instructions:
      "Minta user melakukan pembayaran. Berikan instruksi. Tunggu verifikasi dari sistem (jangan percaya klaim user). Terima event product.change_requested bila user ingin ganti produk di tengah.",
    policy: {
      timeoutSeconds: 900,
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
    description: "Pembayaran gagal. Tawarkan retry / ganti produk / batal.",
    requiredContext: ["order.id", "payment.failure_reason"],
    capabilities: ["payment.status"],
    instructions:
      "Beritahu user pembayaran gagal + alasan. Tawarkan coba lagi atau batal.",
    policy: { timeoutSeconds: 300, onTimeout: "state.timeout" },
    position: { x: 0, y: 0 },
  },
  {
    id: N.CHANGE_PRODUCT,
    kind: "STATE",
    name: "CHANGE_PRODUCT",
    description:
      "User ingin ganti produk di tengah proses. Verifikasi stok produk baru.",
    requiredContext: ["product.sku", "order.items"],
    capabilities: ["product.check_stock", "catalog.list"],
    instructions:
      "User ingin mengganti produk. Tanyakan produk pengganti. Cek stok produk baru. Setelah ganti, PERTAHANKAN konteks yang sudah ada (nama & alamat) dan lanjut ke payment.",
    policy: { timeoutSeconds: 300, onTimeout: "state.timeout" },
    position: { x: 0, y: 0 },
  },
  {
    id: N.ORDER_CONFIRMED,
    kind: "END",
    name: "ORDER_CONFIRMED",
    description: "Pesanan berhasil & dibayar. Kirim konfirmasi/struk.",
    requiredContext: ["order.id", "payment.id"],
    capabilities: ["order.confirm", "notification.send_receipt"],
    instructions: "Konfirmasi pesanan, tampilkan ringkasan & estimasi selesai.",
    policy: {},
    isTerminal: true,
    position: { x: 0, y: 0 },
  },
  {
    id: N.PAYMENT_EXPIRED,
    kind: "END",
    name: "PAYMENT_EXPIRED",
    description: "Waktu pembayaran habis. Pesanan dibatalkan otomatis.",
    requiredContext: ["order.id"],
    capabilities: ["order.cancel"],
    instructions:
      "Beritahu user waktu pembayaran habis dan pesanan dibatalkan.",
    policy: {},
    isTerminal: true,
    position: { x: 0, y: 0 },
  },
  {
    id: N.CANCELLED,
    kind: "END",
    name: "CANCELLED",
    description: "Pesanan dibatalkan.",
    requiredContext: [],
    capabilities: ["order.cancel"],
    instructions: "Konfirmasi pembatalan pesanan.",
    policy: {},
    isTerminal: true,
    position: { x: 0, y: 0 },
  },
]

// ─── TRANSITIONS ────────────────────────────────────────────────────────────

const transitions: TransitionDefinition[] = [
  // START → SELECT_PRODUCT
  {
    id: "tr_start_select",
    sourceStateId: N.START,
    targetStateId: N.SELECT_PRODUCT,
    event: "order.started",
    guards: [],
    priority: 1,
  },

  // SELECT_PRODUCT → CHECK_STOCK (produk dipilih)
  {
    id: "tr_select_stock",
    sourceStateId: N.SELECT_PRODUCT,
    targetStateId: N.CHECK_STOCK,
    event: "product.requested",
    guards: [
      {
        id: "g1",
        logic: "AND",
        conditions: [
          { id: "c1", field: "product.sku", operator: "EXISTS" },
          { id: "c2", field: "order.items", operator: "EXISTS" },
        ],
      },
    ],
    priority: 1,
  },
  // SELECT_PRODUCT timeout → CANCELLED
  {
    id: "tr_select_timeout",
    sourceStateId: N.SELECT_PRODUCT,
    targetStateId: N.CANCELLED,
    event: "state.timeout",
    guards: [],
    priority: 10,
  },
  // SELECT_PRODUCT → batal
  {
    id: "tr_select_cancel",
    sourceStateId: N.SELECT_PRODUCT,
    targetStateId: N.CANCELLED,
    event: "user.cancelled",
    guards: [],
    priority: 5,
  },

  // CHECK_STOCK → tersedia → COLLECT_CUSTOMER
  {
    id: "tr_stock_available",
    sourceStateId: N.CHECK_STOCK,
    targetStateId: N.COLLECT_CUSTOMER,
    event: "product.in_stock",
    guards: [
      {
        id: "g2",
        logic: "AND",
        conditions: [
          {
            id: "c3",
            field: "product.in_stock",
            operator: "==",
            value: "true",
          },
        ],
      },
    ],
    priority: 1,
  },
  // CHECK_STOCK → kosong → RECOMMEND_ALTERNATIVE
  {
    id: "tr_stock_unavailable",
    sourceStateId: N.CHECK_STOCK,
    targetStateId: N.RECOMMEND_ALTERNATIVE,
    event: "product.out_of_stock",
    guards: [
      {
        id: "g3",
        logic: "AND",
        conditions: [
          {
            id: "c4",
            field: "product.in_stock",
            operator: "==",
            value: "false",
          },
        ],
      },
    ],
    priority: 2,
  },
  // CHECK_STOCK timeout → kembali ke SELECT_PRODUCT
  {
    id: "tr_stock_timeout",
    sourceStateId: N.CHECK_STOCK,
    targetStateId: N.SELECT_PRODUCT,
    event: "state.timeout",
    guards: [],
    priority: 10,
  },

  // RECOMMEND_ALTERNATIVE → user pilih alternatif → CHECK_STOCK (re-verify)
  {
    id: "tr_reco_select",
    sourceStateId: N.RECOMMEND_ALTERNATIVE,
    targetStateId: N.CHECK_STOCK,
    event: "product.selected",
    guards: [
      {
        id: "g4",
        logic: "AND",
        conditions: [{ id: "c5", field: "product.sku", operator: "EXISTS" }],
      },
    ],
    priority: 1,
  },
  // RECOMMEND_ALTERNATIVE → batal
  {
    id: "tr_reco_cancel",
    sourceStateId: N.RECOMMEND_ALTERNATIVE,
    targetStateId: N.CANCELLED,
    event: "user.cancelled",
    guards: [],
    priority: 2,
  },
  // RECOMMEND_ALTERNATIVE timeout → CANCELLED
  {
    id: "tr_reco_timeout",
    sourceStateId: N.RECOMMEND_ALTERNATIVE,
    targetStateId: N.CANCELLED,
    event: "state.timeout",
    guards: [],
    priority: 10,
  },

  // COLLECT_CUSTOMER → lengkap → PAYMENT
  {
    id: "tr_customer_payment",
    sourceStateId: N.COLLECT_CUSTOMER,
    targetStateId: N.PAYMENT,
    event: "customer.ready",
    guards: [
      {
        id: "g5",
        logic: "AND",
        conditions: [
          { id: "c6", field: "customer.name", operator: "EXISTS" },
          { id: "c7", field: "customer.address", operator: "EXISTS" },
        ],
      },
    ],
    priority: 1,
  },
  // COLLECT_CUSTOMER → batal
  {
    id: "tr_customer_cancel",
    sourceStateId: N.COLLECT_CUSTOMER,
    targetStateId: N.CANCELLED,
    event: "user.cancelled",
    guards: [],
    priority: 2,
  },

  // PAYMENT → sukses → ORDER_CONFIRMED
  {
    id: "tr_payment_success",
    sourceStateId: N.PAYMENT,
    targetStateId: N.ORDER_CONFIRMED,
    event: "payment.success",
    guards: [
      {
        id: "g6",
        logic: "AND",
        conditions: [
          {
            id: "c8",
            field: "payment.status",
            operator: "==",
            value: "success",
          },
          {
            id: "c9",
            field: "payment.amount",
            operator: ">=",
            value: "order.total",
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
        id: "g7",
        logic: "AND",
        conditions: [
          {
            id: "c10",
            field: "payment.status",
            operator: "==",
            value: "failed",
          },
        ],
      },
    ],
    priority: 2,
  },
  // PAYMENT → INTERUPSI ganti produk → CHANGE_PRODUCT
  {
    id: "tr_payment_change",
    sourceStateId: N.PAYMENT,
    targetStateId: N.CHANGE_PRODUCT,
    event: "product.change_requested",
    guards: [],
    priority: 3,
  },
  // PAYMENT timeout → PAYMENT_EXPIRED
  {
    id: "tr_payment_timeout",
    sourceStateId: N.PAYMENT,
    targetStateId: N.PAYMENT_EXPIRED,
    event: "state.timeout",
    guards: [],
    priority: 10,
  },
  // PAYMENT retry habis → CANCELLED
  {
    id: "tr_payment_exhausted",
    sourceStateId: N.PAYMENT,
    targetStateId: N.CANCELLED,
    event: "retry.exhausted",
    guards: [],
    priority: 6,
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
  // PAYMENT_FAILED → ganti produk → CHANGE_PRODUCT
  {
    id: "tr_failed_change",
    sourceStateId: N.PAYMENT_FAILED,
    targetStateId: N.CHANGE_PRODUCT,
    event: "product.change_requested",
    guards: [],
    priority: 3,
  },

  // CHANGE_PRODUCT → stok ok → CHECK_STOCK → (alamat sudah ada) → PAYMENT
  {
    id: "tr_change_verify",
    sourceStateId: N.CHANGE_PRODUCT,
    targetStateId: N.CHECK_STOCK,
    event: "product.available",
    guards: [
      {
        id: "g8",
        logic: "AND",
        conditions: [
          {
            id: "c11",
            field: "product.in_stock",
            operator: "==",
            value: "true",
          },
        ],
      },
    ],
    priority: 1,
  },
  // CHANGE_PRODUCT → stok kosong → RECOMMEND_ALTERNATIVE
  {
    id: "tr_change_unavailable",
    sourceStateId: N.CHANGE_PRODUCT,
    targetStateId: N.RECOMMEND_ALTERNATIVE,
    event: "product.unavailable",
    guards: [
      {
        id: "g9",
        logic: "AND",
        conditions: [
          {
            id: "c12",
            field: "product.in_stock",
            operator: "==",
            value: "false",
          },
        ],
      },
    ],
    priority: 2,
  },
  // CHANGE_PRODUCT → tetap produk lama → lanjut ke COLLECT_CUSTOMER (cek alamat)
  {
    id: "tr_change_keep",
    sourceStateId: N.CHANGE_PRODUCT,
    targetStateId: N.COLLECT_CUSTOMER,
    event: "keep.current",
    guards: [],
    priority: 3,
  },
  // CHANGE_PRODUCT → batal
  {
    id: "tr_change_cancel",
    sourceStateId: N.CHANGE_PRODUCT,
    targetStateId: N.CANCELLED,
    event: "user.cancelled",
    guards: [],
    priority: 5,
  },
]

// ─── WORKFLOW DEFINITION ─────────────────────────────────────────────────────

export const orderMakananWorkflow: WorkflowDefinition = {
  slug: "order-makanan",
  name: "Order Makanan (Kafe)",
  description:
    "Workflow pemesanan makanan/minuman: pilih produk, cek stok, rekomendasi alternatif saat stok kosong, kumpulkan nama & alamat, pembayaran, dan mendukung interupsi ganti produk di tengah proses.",
  schemaVersion: 1,
  status: "DRAFT",
  entryNodeId: N.START,
  nodes,
  transitions,
  policy: {
    maxDurationSeconds: 3600, // 1 jam
    interruptible: "USER_REQUESTED",
    priority: 10,
  },
  triggers: [
    { event: "order.makanan.requested", source: "intent" },
    { event: "food.order.started", source: "api" },
  ],
}
