/**
 * Utilitas untuk membangun & memanipulasi WorkflowDefinition.
 */
import { type Edge, MarkerType, type Node } from "@xyflow/react"
import type {
  GuardCondition,
  GuardGroup,
  TransitionDefinition,
  WorkflowDefinition,
  WorkflowNode,
  WorkflowNodeKind,
  WorkflowValidationIssue,
  WorkflowValidationResult,
} from "./workflow"

export const WORKFLOW_SCHEMA_VERSION = 1

let counter = 0
/** generate id pendek yang unik */
export function uid(prefix = "n"): string {
  counter += 1
  return `${prefix}_${Date.now().toString(36)}_${counter.toString(36)}`
}

/** posisi default node saat ditambahkan */
export function nextNodePosition(
  nodes: WorkflowNode[],
  kind: WorkflowNodeKind,
): { x: number; y: number } {
  const base = nodes.length * 80
  // START biasanya di kiri atas, sisanya menyusul
  return {
    x: kind === "START" ? 0 : 120 + base,
    y: kind === "START" ? 0 : base,
  }
}

/** nama default untuk tipe node */
export function defaultNodeName(kind: WorkflowNodeKind): string {
  switch (kind) {
    case "START":
      return "START"
    case "END":
      return "END"
    case "DECISION":
      return "DECISION"
    case "EVENT":
      return "EVENT"
    case "STATE":
      return "NEW_STATE"
    default:
      return "STATE"
  }
}

/** warna node berdasarkan kind */
export function nodeKindColor(kind: WorkflowNodeKind): string {
  switch (kind) {
    case "START":
      return "#25a45f" // success
    case "END":
      return "#d83838" // danger
    case "DECISION":
      return "#f39e1a" // warning
    case "EVENT":
      return "#1e8cc1" // info
    default:
      return "#1f8fca" // primary
  }
}

export function createWorkflowNode(kind: WorkflowNodeKind): WorkflowNode {
  return {
    id: uid(kind === "START" ? "start" : kind === "END" ? "end" : "node"),
    kind,
    name: defaultNodeName(kind),
    description: "",
    requiredContext: [],
    capabilities: [],
    instructions: "",
    policy: {},
    isTerminal: kind === "END",
    position: { x: 0, y: 0 },
  }
}

export function createEmptyWorkflow(slug = "new-workflow"): WorkflowDefinition {
  return {
    slug,
    name: "Untitled Workflow",
    description: "",
    schemaVersion: WORKFLOW_SCHEMA_VERSION,
    status: "DRAFT",
    nodes: [],
    transitions: [],
    policy: {
      interruptible: "USER_REQUESTED",
      priority: 10,
    },
    triggers: [],
  }
}

/** Buat guard group kosong */
export function createGuardGroup(): GuardGroup {
  return { id: uid("g"), logic: "AND", conditions: [] }
}

/** Buat guard condition kosong */
export function createGuardCondition(): GuardCondition {
  return { id: uid("c"), field: "", operator: "==", value: "" }
}

export function createTransition(
  sourceStateId: string,
  event = "",
  targetStateId = "",
): TransitionDefinition {
  return {
    id: uid("tr"),
    sourceStateId,
    event,
    targetStateId,
    guards: [],
    priority: 10,
  }
}

/** Daftar node type yang valid untuk drag dari palette */
export const PALETTE_KINDS: WorkflowNodeKind[] = [
  "START",
  "STATE",
  "DECISION",
  "EVENT",
  "END",
]

/** ------------------------------------------------------------------ */
/** Validasi (PRD 54, 164) */
/** ------------------------------------------------------------------ */

export function validateWorkflow(
  wf: WorkflowDefinition,
): WorkflowValidationResult {
  const issues: WorkflowValidationIssue[] = []

  if (!wf.entryNodeId) {
    issues.push({
      severity: "error",
      code: "MISSING_START",
      message:
        "Workflow belum memiliki node START. Tambahkan node START terlebih dahulu.",
    })
  }

  const starts = wf.nodes.filter((n) => n.kind === "START")
  if (starts.length === 0) {
    issues.push({
      severity: "error",
      code: "MISSING_START",
      message: "Tidak ada node START di canvas.",
    })
  }

  // dead-end state (bukan terminal, tapi tidak punya transisi keluar)
  const terminalIds = new Set(
    wf.nodes.filter((n) => n.isTerminal || n.kind === "END").map((n) => n.id),
  )
  const hasTransitionsFrom = (stateId: string) =>
    wf.transitions.some((t) => t.sourceStateId === stateId)

  for (const node of wf.nodes) {
    if (!terminalIds.has(node.id) && !hasTransitionsFrom(node.id)) {
      issues.push({
        severity: node.kind === "DECISION" ? "error" : "warning",
        code: "DEAD_END",
        message: `State "${node.name}" tidak memiliki transisi keluar.`,
        nodeId: node.id,
      })
    }
  }

  // missing target
  const nodeIds = new Set(wf.nodes.map((n) => n.id))
  for (const tr of wf.transitions) {
    if (!nodeIds.has(tr.targetStateId)) {
      issues.push({
        severity: "error",
        code: "MISSING_TARGET",
        message: `Transisi "${tr.event || "?"}" menuju state yang tidak ada.`,
        edgeId: tr.id,
      })
    }
  }

  // duplicate transition: event+source yang sama
  const seen = new Set<string>()
  for (const tr of wf.transitions) {
    const key = `${tr.sourceStateId}:${tr.event}`
    if (seen.has(key)) {
      issues.push({
        severity: "error",
        code: "DUPLICATE_TRANSITION",
        message: `Transisi duplikat untuk event "${tr.event}".`,
        edgeId: tr.id,
      })
    }
    seen.add(key)
  }

  // no terminal state
  if (!terminalIds.size) {
    issues.push({
      severity: "error",
      code: "NO_TERMINAL",
      message: "Workflow tidak memiliki node terminal (END).",
    })
  }

  // unreachable states (dari START)
  const reachable = reachableStates(wf)
  for (const node of wf.nodes) {
    if (node.kind !== "START" && !reachable.has(node.id)) {
      issues.push({
        severity: "error",
        code: "UNREACHABLE_STATE",
        message: `State "${node.name}" tidak dapat dicapai dari START.`,
        nodeId: node.id,
      })
    }
  }

  // cycle tanpa exit (warning)
  if (hasCycleWithoutExit(wf, terminalIds)) {
    issues.push({
      severity: "warning",
      code: "CYCLE_NO_EXIT",
      message:
        "Terdapat siklus tanpa jalur keluar. Pastikan ada transisi keluar dari siklus.",
    })
  }

  return { valid: issues.every((i) => i.severity === "warning"), issues }
}

function reachableStates(wf: WorkflowDefinition): Set<string> {
  const reachable = new Set<string>()
  const start = wf.nodes.find((n) => n.kind === "START")
  if (!start) return reachable
  const stack = [start.id]
  while (stack.length > 0) {
    const id = stack.pop()
    if (id === undefined) break
    if (reachable.has(id)) continue
    reachable.add(id)
    for (const tr of wf.transitions.filter((t) => t.sourceStateId === id)) {
      stack.push(tr.targetStateId)
    }
  }
  return reachable
}

function hasCycleWithoutExit(
  wf: WorkflowDefinition,
  _terminalIds: Set<string>,
): boolean {
  // deteksi cycle dengan DFS sederhana
  const visited = new Set<string>()
  const inStack = new Set<string>()
  const outEdges = new Map<string, string[]>()
  for (const tr of wf.transitions) {
    const list = outEdges.get(tr.sourceStateId) ?? []
    list.push(tr.targetStateId)
    outEdges.set(tr.sourceStateId, list)
  }

  function dfs(nodeId: string): boolean {
    if (inStack.has(nodeId)) return true
    if (visited.has(nodeId)) return false
    visited.add(nodeId)
    inStack.add(nodeId)
    const targets = outEdges.get(nodeId) ?? []
    for (const t of targets) {
      if (dfs(t)) return true
    }
    inStack.delete(nodeId)
    return false
  }

  for (const node of wf.nodes) {
    if (dfs(node.id)) return true
  }
  return false
}

/** ------------------------------------------------------------------ */
/** Konversi React Flow <-> domain model */
/** ------------------------------------------------------------------ */

/** React Flow node type yang kita daftarkan */
export type FlowNodeType = "start" | "state" | "decision" | "event" | "end"

export function domainKindToFlowType(kind: WorkflowNodeKind): FlowNodeType {
  switch (kind) {
    case "START":
      return "start"
    case "DECISION":
      return "decision"
    case "EVENT":
      return "event"
    case "END":
      return "end"
    default:
      return "state"
  }
}

export function flowTypeToDomainKind(type: FlowNodeType): WorkflowNodeKind {
  switch (type) {
    case "start":
      return "START"
    case "decision":
      return "DECISION"
    case "event":
      return "EVENT"
    case "end":
      return "END"
    default:
      return "STATE"
  }
}

/** Data yang disimpan di node React Flow */
export interface FlowNodeData extends WorkflowNode, Record<string, unknown> {
  sourceStateId: string
}

/** Konversi domain -> React Flow nodes */
export function toFlowNodes(wf: WorkflowDefinition): Node[] {
  return wf.nodes.map((n) => ({
    id: n.id,
    type: domainKindToFlowType(n.kind),
    position: n.position,
    data: { ...n, sourceStateId: n.id } as FlowNodeData,
  }))
}

/**
 * Palet warna per index edge keluar dari satu node.
 * Terjamin berbeda satu sama lain, kontras, mudah dibedakan.
 * Urutan: biru, ungu, oranye, teal, pink, coklat, indigo, lime
 */
const EDGE_PALETTE = [
  "#2563eb", // blue
  "#7c3aed", // violet
  "#ea580c", // orange
  "#0891b2", // cyan
  "#db2777", // pink
  "#b45309", // amber
  "#4f46e5", // indigo
  "#65a30d", // lime
]

/** Warna tone untuk event yang jelas positive/negative */
const TONE_POSITIVE = "#16a34a" // green
const TONE_NEGATIVE = "#dc2626" // red

/** Tentukan apakah event punya tone yang jelas */
function eventTone(event: string): "positive" | "negative" | null {
  const e = event.toLowerCase()
  const POS = [
    "success",
    "available",
    "confirmed",
    "reserved",
    "completed",
    "approved",
    "valid",
  ]
  const NEG = [
    "failed",
    "unavailable",
    "cancelled",
    "cancel",
    "timeout",
    "expired",
    "error",
    "rejected",
    "insufficient",
    "denied",
    "invalid",
    "exhausted",
    "abort",
  ]
  if (POS.some((p) => e.includes(p))) return "positive"
  if (NEG.some((n) => e.includes(n))) return "negative"
  return null
}

/**
 * Assign warna per edge, terjamin berbeda untuk setiap edge
 * yang keluar dari node yang sama.
 *
 * Logika:
 * 1. Kumpulkan semua edge per sourceStateId.
 * 2. Untuk setiap group, assign warna:
 *    - positive tone → hijau
 *    - negative tone → merah
 *    - neutral → ambil dari EDGE_PALETTE[i] (berbeda per index)
 *    - TAPI jika 2+ edge dari source yang sama punya warna sama
 *      (misal 2 neutral) → override dengan palet agar berbeda.
 */
function assignEdgeColors(
  transitions: Array<{ id: string; sourceStateId: string; event: string }>,
): Map<string, string> {
  const colorMap = new Map<string, string>()

  // Group per source
  const bySource = new Map<string, typeof transitions>()
  for (const t of transitions) {
    const group = bySource.get(t.sourceStateId) ?? []
    group.push(t)
    bySource.set(t.sourceStateId, group)
  }

  for (const group of bySource.values()) {
    const only = group[0]
    if (group.length === 1 && only) {
      // Hanya 1 edge keluar: pakai tone, atau biru jika netral
      const tone = eventTone(only.event)
      colorMap.set(
        only.id,
        tone === "positive"
          ? TONE_POSITIVE
          : tone === "negative"
            ? TONE_NEGATIVE
            : "#2563eb",
      )
      continue
    }

    // Banyak edge keluar dari node yang sama.
    // Aturan WAJIB: setiap panah dari node yang sama punya warna BEDA.
    //
    // Strategi: urutkan edge agar yang punya tone jelas (positive/negative)
    // diproses lebih dulu sehingga dapat warna hijau/merah; edge neutral
    // dan edge yang tone-nya sudah terpakai mengambil warna palet unik.
    // Hasilnya: semua warna berbeda satu sama lain, dan tone tetap terbaca.
    const usedColors = new Set<string>()

    // Urutkan: positive & negative lebih dulu, sisanya (neutral/duplikat) belakangan
    const sorted = [...group].sort((a, b) => {
      const ta = eventTone(a.event)
      const tb = eventTone(b.event)
      const prio = (t: string | null) => (t === null ? 2 : 1)
      return prio(ta) - prio(tb)
    })

    for (const t of sorted) {
      const tone = eventTone(t.event)

      // warna utama berdasarkan tone (belum dipakai)
      let color = ""
      if (tone === "positive" && !usedColors.has(TONE_POSITIVE)) {
        color = TONE_POSITIVE
      } else if (tone === "negative" && !usedColors.has(TONE_NEGATIVE)) {
        color = TONE_NEGATIVE
      }

      // jika tone tidak tersedia/tidak jelas → ambil warna palet yang belum dipakai
      if (!color) {
        for (let i = 0; i < EDGE_PALETTE.length; i++) {
          const c = EDGE_PALETTE[i]
          if (c && !usedColors.has(c)) {
            color = c
            break
          }
        }
      }
      // fallback ekstrem jika palet habis
      if (!color) color = "#64748b"

      usedColors.add(color)
      colorMap.set(t.id, color)
    }
  }

  return colorMap
}

/** Konversi domain -> React Flow edges */
export function toFlowEdges(wf: WorkflowDefinition): Edge[] {
  const colorMap = assignEdgeColors(wf.transitions)

  return wf.transitions.map((t) => {
    const color = colorMap.get(t.id) ?? "#64748b"
    return {
      id: t.id,
      source: t.sourceStateId,
      target: t.targetStateId,
      type: "transition",
      label: t.event || "event",
      markerEnd: {
        type: MarkerType.ArrowClosed,
        width: 18,
        height: 18,
        color,
      },
      data: { transition: t, color },
    }
  })
}

/** Konversi React Flow nodes -> domain nodes */
export function fromFlowNodes(nodes: Node[]): WorkflowNode[] {
  return nodes.map((n) => {
    const data = n.data as FlowNodeData
    return {
      id: n.id,
      kind: data.kind,
      name: data.name ?? defaultNodeName(data.kind),
      description: data.description ?? "",
      requiredContext: data.requiredContext ?? [],
      capabilities: data.capabilities ?? [],
      instructions: data.instructions ?? "",
      guardGroups: data.guardGroups ?? [],
      policy: data.policy ?? {},
      isTerminal: data.isTerminal,
      position: n.position,
    }
  })
}
