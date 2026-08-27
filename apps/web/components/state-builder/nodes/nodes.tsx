"use client"

import { Handle, Position } from "@xyflow/react"
import {
  CircleCheck,
  CircleStop,
  GitBranch,
  Play,
  Workflow as WorkflowIcon,
} from "lucide-react"
import type { WorkflowNodeKind } from "../types/workflow"
import type { FlowNodeData } from "../types/workflow.utils"
import { nodeKindColor } from "../types/workflow.utils"

const KIND_META: Record<
  WorkflowNodeKind,
  { label: string; icon: typeof Play }
> = {
  START: { label: "START", icon: Play },
  STATE: { label: "STATE", icon: WorkflowIcon },
  DECISION: { label: "DECISION", icon: GitBranch },
  EVENT: { label: "EVENT", icon: CircleStop },
  END: { label: "END", icon: CircleCheck },
}

/** Props yang diterima custom node dari React Flow (hanya yang kita pakai) */
interface BaseNodeProps {
  data: FlowNodeData
}

/** Handle atas (target) — dipakai di hampir semua node */
function TargetHandle() {
  return (
    <Handle
      type="target"
      position={Position.Top}
      className="!h-3 !w-3 !border-2 !border-white !bg-slate-400"
    />
  )
}

/** Handle bawah (source) — dipakai di semua node kecuali END */
function SourceHandle() {
  return (
    <Handle
      type="source"
      position={Position.Bottom}
      className="!h-3 !w-3 !border-2 !border-white !bg-slate-400"
    />
  )
}

/** Tampilan label utama node (ikon + nama) */
function NodeHeader({
  kind,
  name,
  color,
}: {
  kind: WorkflowNodeKind
  name: string
  color: string
}) {
  const meta = KIND_META[kind]
  const Icon = meta.icon
  return (
    <div className="flex items-center justify-center gap-1.5">
      <Icon className="h-3.5 w-3.5" style={{ color }} />
      <span className="truncate text-sm font-semibold text-slate-800">
        {name}
      </span>
    </div>
  )
}

/**
 * START / END — bentuk OVAL (terminator), khas flowchart.
 */
function TerminatorNode({ data }: BaseNodeProps) {
  const isEnd = data.kind === "END"
  const color = nodeKindColor(data.kind)
  const meta = KIND_META[data.kind]
  const Icon = meta.icon

  return (
    <div className="group relative">
      {!isEnd ? <TargetHandle /> : null}
      <div
        className="flex min-w-[140px] items-center gap-2 rounded-full border-2 px-5 py-2.5 shadow-sm transition-shadow group-hover:shadow-md"
        style={{
          borderColor: color,
          backgroundColor: `${color}14`,
        }}
      >
        <Icon className="h-4 w-4 shrink-0" style={{ color }} />
        <span className="truncate text-sm font-bold" style={{ color }}>
          {data.name || meta.label}
        </span>
      </div>
      {!isEnd ? <SourceHandle /> : null}
    </div>
  )
}

/**
 * STATE — bentuk KOTAK (process), khas flowchart.
 */
function StateNode({ data }: BaseNodeProps) {
  const color = nodeKindColor(data.kind)
  return (
    <div className="group relative min-w-[200px]">
      <TargetHandle />
      <div
        className="rounded-md border-2 bg-white px-3 py-2 shadow-sm transition-shadow group-hover:shadow-md"
        style={{ borderColor: color }}
      >
        <NodeHeader kind={data.kind} name={data.name} color={color} />
        {data.description ? (
          <div className="mt-0.5 truncate text-center text-xs text-slate-400">
            {data.description}
          </div>
        ) : null}
        {data.requiredContext.length > 0 || data.capabilities.length > 0 ? (
          <div className="mt-2 flex flex-wrap justify-center gap-1">
            {data.capabilities.slice(0, 2).map((cap) => (
              <span
                key={cap}
                className="rounded bg-primary-50 px-1.5 py-0.5 text-[10px] font-medium text-primary-700"
              >
                {cap}
              </span>
            ))}
            {data.requiredContext.length > 0 ? (
              <span className="rounded bg-slate-100 px-1.5 py-0.5 text-[10px] text-slate-500">
                ctx {data.requiredContext.length}
              </span>
            ) : null}
          </div>
        ) : null}
      </div>
      <SourceHandle />
    </div>
  )
}

/**
 * DECISION — bentuk BELAH KETUPAT (diamond), khas flowchart.
 * Menggunakan clip-path polygon. Teks berada di dalam diamond.
 */
function DecisionNode({ data }: BaseNodeProps) {
  const color = nodeKindColor(data.kind)
  const clip = "polygon(50% 0%, 100% 50%, 50% 100%, 0% 50%)"

  return (
    <div className="group relative">
      <TargetHandle />
      <div className="relative h-[120px] w-[220px]">
        {/* layer border (diamond berwarna, di belakang) */}
        <div
          className="absolute inset-0"
          style={{
            clipPath: clip,
            background: color,
          }}
        />
        {/* layer isi (putih, lebih kecil -> efek border) */}
        <div
          className="absolute inset-0 flex items-center justify-center"
          style={{
            clipPath: clip,
            transform: "scale(0.92)",
            background: "#ffffff",
            boxShadow: "0 1px 3px rgba(0,0,0,0.1)",
          }}
        >
          <div className="flex flex-col items-center p-4 text-center">
            <NodeHeader kind={data.kind} name={data.name} color={color} />
            <span className="mt-1 text-[10px] uppercase tracking-wide text-slate-400">
              branch
            </span>
          </div>
        </div>
      </div>
      <SourceHandle />
    </div>
  )
}

/**
 * EVENT — bentuk PARALLELOGRAM (input/output), khas flowchart.
 * Kotak dengan sisi miring kiri-kanan.
 */
function EventNode({ data }: BaseNodeProps) {
  const color = nodeKindColor(data.kind)
  return (
    <div className="group relative min-w-[200px]">
      <TargetHandle />
      <div
        className="rounded-md border-2 border-dashed bg-white px-3 py-2 shadow-sm transition-shadow group-hover:shadow-md"
        style={{ borderColor: color }}
      >
        <div className="flex items-center justify-center gap-1.5">
          <CircleStop className="h-3.5 w-3.5" style={{ color }} />
          <span className="truncate text-sm font-semibold text-slate-800">
            {data.name}
          </span>
        </div>
        <div className="mt-0.5 text-center text-[10px] uppercase tracking-wide text-slate-400">
          event
        </div>
      </div>
      <SourceHandle />
    </div>
  )
}

export const nodeTypes = {
  start: TerminatorNode,
  state: StateNode,
  decision: DecisionNode,
  event: EventNode,
  end: TerminatorNode,
}

export const nodeTypeList = [
  {
    type: "start",
    kind: "START",
    icon: Play,
    color: "#25a45f",
    label: "Start",
  },
  {
    type: "state",
    kind: "STATE",
    icon: WorkflowIcon,
    color: "#1f8fca",
    label: "State",
  },
  {
    type: "decision",
    kind: "DECISION",
    icon: GitBranch,
    color: "#f39e1a",
    label: "Decision",
  },
  {
    type: "event",
    kind: "EVENT",
    icon: CircleStop,
    color: "#1e8cc1",
    label: "Event",
  },
  {
    type: "end",
    kind: "END",
    icon: CircleCheck,
    color: "#d83838",
    label: "End",
  },
] as const
