/**
 * Domain model untuk State Builder (workflow definition).
 *
 * Sesuai PRD section 12, 33, 35, 36, 48, 25, 26.
 * Model ini dipisahkan dari model React Flow (node/edge) —
 * React Flow hanya untuk visual, ini adalah model domain yang
 * nanti dikirim ke backend sebagai workflow definition (DSL).
 */

export type WorkflowNodeKind = "START" | "STATE" | "DECISION" | "END" | "EVENT"

/** Lifecycle workflow definition sesuai PRD section 9 */
export type WorkflowStatus =
  | "DRAFT"
  | "VALIDATING"
  | "VALID"
  | "PUBLISHED"
  | "ARCHIVED"

export type GuardOperator =
  | "=="
  | "!="
  | ">"
  | ">="
  | "<"
  | "<="
  | "IN"
  | "NOT_IN"
  | "EXISTS"
  | "NOT_EXISTS"

export interface GuardCondition {
  id?: string
  /** field yang dicek, misal: payment.status */
  field: string
  operator: GuardOperator
  /** value pembanding (opsional untuk EXISTS/NOT_EXISTS) */
  value?: string
}

export interface GuardGroup {
  id?: string
  /** logical operator yang menggabungkan conditions: AND / OR */
  logic: "AND" | "OR"
  conditions: GuardCondition[]
}

/** Transisi sesuai PRD section 33 */
export interface TransitionDefinition {
  id: string
  /** id state asal */
  sourceStateId: string
  /** event yang memicu transisi, misal: payment.success */
  event: string
  /** id state target */
  targetStateId: string
  /** guard untuk transisi ini */
  guards: GuardGroup[]
  /** prioritas numerik, lebih kecil = dievaluasi lebih dulu (PRD 34) */
  priority: number
  label?: string
}

/** Retry policy sesuai PRD section 48 */
export interface RetryPolicy {
  maxAttempts: number
  backoffMs: number
  retryableEvents: string[]
}

/** Human handoff (PRD 49) */
export interface HumanHandoffPolicy {
  enabled: boolean
}

export interface StatePolicy {
  /** timeout state dalam detik (PRD 25) */
  timeoutSeconds?: number
  /** event yang dipicu saat timeout, misal: state.timeout */
  onTimeout?: string
  retry?: RetryPolicy
  humanHandoff?: HumanHandoffPolicy
}

/** Node/State definition sesuai PRD section 12 */
export interface WorkflowNode {
  id: string
  kind: WorkflowNodeKind
  /** nama node / state, misal: PAYMENT */
  name: string
  /** deskripsi state */
  description?: string
  /** context yang dibutuhkan (PRD 36) */
  requiredContext: string[]
  /** capability yang diizinkan (PRD 18) */
  capabilities: string[]
  /** instruksi untuk LLM (PRD 103) */
  instructions?: string
  /** guard yang melekat pada state (opsional) */
  guardGroups?: GuardGroup[]
  /** policy state: timeout, retry, handoff */
  policy: StatePolicy
  /** true jika ini terminal state (PRD 45) */
  isTerminal?: boolean
  /** posisi visual (dari React Flow, disimpan juga di sini agar persist) */
  position: { x: number; y: number }
}

export interface WorkflowPolicy {
  /** max duration workflow dalam detik (PRD 26) */
  maxDurationSeconds?: number
  /** apakah workflow bisa diinterupsi (PRD 42) */
  interruptible: "NEVER" | "USER_REQUESTED" | "HIGH_PRIORITY" | "ALWAYS"
  /** prioritas tie-breaker (PRD 41) */
  priority: number
}

export interface WorkflowTrigger {
  event: string
  source: "event" | "api" | "intent" | "webhook" | "schedule"
}

/**
 * Definisi workflow penuh (DSL). Sesuai PRD section 161.
 * Ini representasi yang disimpan & divalidasi.
 */
export interface WorkflowDefinition {
  /** slug unique per tenant (PRD 5) */
  slug: string
  name: string
  description?: string
  schemaVersion: number
  status: WorkflowStatus
  /** id node START */
  entryNodeId?: string
  nodes: WorkflowNode[]
  transitions: TransitionDefinition[]
  policy: WorkflowPolicy
  triggers: WorkflowTrigger[]
  updatedBy?: string
  updatedAt?: string
}

/** Hasil validasi workflow (PRD 54, 164) */
export interface WorkflowValidationIssue {
  severity: "error" | "warning"
  code:
    | "MISSING_START"
    | "UNREACHABLE_STATE"
    | "MISSING_TARGET"
    | "DUPLICATE_TRANSITION"
    | "AMBIGUOUS_TRANSITION"
    | "INVALID_GUARD"
    | "INVALID_CAPABILITY"
    | "INVALID_CONTEXT_REFERENCE"
    | "DEAD_END"
    | "NO_TERMINAL"
    | "CYCLE_NO_EXIT"
    | "MISSING_TIMEOUT_BEHAVIOR"
    | "UNUSED_CAPABILITY"
    | "UNUSED_CONTEXT"
  message: string
  /** node/edge yang berkaitan, untuk UX error pointing (PRD 165) */
  nodeId?: string
  edgeId?: string
}

export interface WorkflowValidationResult {
  valid: boolean
  issues: WorkflowValidationIssue[]
}
