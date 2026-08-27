"use client"

import {
  BaseEdge,
  EdgeLabelRenderer,
  type EdgeProps,
  getSmoothStepPath,
} from "@xyflow/react"
import type { TransitionDefinition } from "../types/workflow"

export interface TransitionEdgeData {
  transition?: TransitionDefinition
  /** warna edge yang sudah di-compute (terjamin beda antar cabang) */
  color?: string
}

/**
 * Custom edge untuk transisi antar state.
 *
 * - Garis lurus orthogonal (step path).
 * - Warna edge berasal dari data.color yang sudah di-assign di store
 *   (assignEdgeColors) sehingga TERJAMIN berbeda untuk setiap panah
 *   yang keluar dari node yang sama — tidak ada warna bertabrakan.
 * - Label selalu tampil, posisi dekat node sumber agar tidak bertumpuk.
 */
export function TransitionEdge({
  id,
  sourceX,
  sourceY,
  targetX,
  targetY,
  sourcePosition,
  targetPosition,
  markerEnd,
  data,
  label,
  selected,
}: EdgeProps) {
  const [edgePath, midX, midY] = getSmoothStepPath({
    sourceX,
    sourceY,
    sourcePosition,
    targetX,
    targetY,
    targetPosition,
    borderRadius: 8,
  })

  const edgeData = data as TransitionEdgeData | undefined
  const transition = edgeData?.transition
  const eventLabel =
    transition?.event ??
    (typeof label === "string" ? label : undefined) ??
    "event"
  // warna dari store (terjamin unik per cabang), fallback abu-abu
  const color = selected ? "#1f8fca" : (edgeData?.color ?? "#64748b")

  // posisi label dekat source (25%), bukan di tengah
  const labelX = sourceX + (midX - sourceX) * 0.25
  const labelY = sourceY + (midY - sourceY) * 0.25 + 18
  const offsetX =
    targetX - sourceX < -20 ? -10 : targetX - sourceX > 20 ? 10 : 0

  return (
    <>
      <BaseEdge
        id={id}
        path={edgePath}
        markerEnd={markerEnd}
        interactionWidth={20}
        style={{
          stroke: color,
          strokeWidth: selected ? 2.5 : 1.75,
        }}
      />
      <EdgeLabelRenderer>
        <div
          style={{
            position: "absolute",
            transform: `translate(-50%, 0%) translate(${labelX + offsetX}px, ${labelY}px)`,
            pointerEvents: "all",
            zIndex: 10,
          }}
        >
          <span
            className="block whitespace-nowrap rounded border border-slate-300 bg-white px-1.5 py-0.5 text-[10px] font-medium text-slate-700 shadow-sm"
            style={{ borderColor: color, color }}
          >
            {eventLabel}
          </span>
        </div>
      </EdgeLabelRenderer>
    </>
  )
}

export const edgeTypes = {
  transition: TransitionEdge,
}
