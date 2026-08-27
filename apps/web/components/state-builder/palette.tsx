"use client"

import { GripVertical } from "lucide-react"
import { useCallback, useState } from "react"
import { nodeTypeList } from "./nodes/nodes"

/** DND drag payload type */
export const NODE_DRAG_TYPE = "application/x-workflow-node"

/** Type payload untuk reorder di dalam palette */
const REORDER_DRAG_TYPE = "application/x-palette-reorder"

/** Kunci localStorage untuk urutan palette */
const PALETTE_ORDER_KEY = "state-builder:palette-order"

/** Urutan default sesuai deklarasi nodeTypeList */
const DEFAULT_ORDER = nodeTypeList.map((n) => n.type)

/** Load urutan tersimpan; fallback ke default jika rusak */
function loadOrder(): string[] {
  try {
    const raw = localStorage.getItem(PALETTE_ORDER_KEY)
    if (!raw) return [...DEFAULT_ORDER]
    const parsed = JSON.parse(raw) as unknown
    if (!Array.isArray(parsed)) return [...DEFAULT_ORDER]
    const order = parsed.filter(
      (t): t is string => typeof t === "string" && DEFAULT_ORDER.includes(t as never),
    )
    // pastikan semua tipe ada (antisipasi tipe baru ditambahkan)
    for (const t of DEFAULT_ORDER) {
      if (!order.includes(t)) order.push(t)
    }
    return order
  } catch {
    return [...DEFAULT_ORDER]
  }
}

/**
 * Palette berisi daftar node type yang bisa di-drag ke canvas.
 * - Drag pada badan item → menambah node ke canvas.
 * - Drag pada ikon grip (kiri) → mengubah urutan item dalam palette.
 * Urutan disimpan ke localStorage agar persisten.
 */
export function Palette() {
  const [order, setOrder] = useState<string[]>(loadOrder)
  const [dragType, setDragType] = useState<string | null>(null)

  const reorder = useCallback((fromType: string, toType: string) => {
    setOrder((prev) => {
      if (fromType === toType) return prev
      const next = [...prev]
      const fromIdx = next.indexOf(fromType)
      const toIdx = next.indexOf(toType)
      if (fromIdx < 0 || toIdx < 0) return prev
      next.splice(fromIdx, 1)
      next.splice(toIdx, 0, fromType)
      try {
        localStorage.setItem(PALETTE_ORDER_KEY, JSON.stringify(next))
      } catch {
        // abaikan error storage
      }
      return next
    })
  }, [])

  const items = order
    .map((type) => nodeTypeList.find((n) => n.type === type))
    .filter((n): n is (typeof nodeTypeList)[number] => n != null)

  return (
    <div className="flex flex-col gap-2 p-3">
      <p className="flex items-center justify-between text-xs font-semibold uppercase tracking-wide text-slate-400">
        <span>Nodes</span>
        <span className="text-[9px] font-normal normal-case text-slate-300">
          geser grip utk urut
        </span>
      </p>
      {items.map((item) => {
        const Icon = item.icon
        const isReorder = dragType === REORDER_DRAG_TYPE
        return (
          <div
            key={item.type}
            draggable
            onDragStart={(e) => {
              e.dataTransfer.setData(NODE_DRAG_TYPE, item.kind)
              e.dataTransfer.effectAllowed = "move"
            }}
            onDragOver={(e) => {
              // hanya reorder jika drag berasal dari grip
              if (e.dataTransfer.types.includes(REORDER_DRAG_TYPE)) {
                e.preventDefault()
                e.dataTransfer.dropEffect = "move"
              }
            }}
            onDrop={(e) => {
              e.preventDefault()
              const fromType = e.dataTransfer.getData(REORDER_DRAG_TYPE)
              if (fromType) reorder(fromType, item.type)
            }}
            className={`flex cursor-grab items-center gap-2 rounded-lg border border-slate-200 bg-white px-2.5 py-2 text-sm text-slate-700 shadow-sm hover:border-slate-300 hover:shadow-md active:cursor-grabbing ${
              isReorder ? "opacity-50" : ""
            }`}
          >
            {/* Grip handle: hanya untuk reorder, bukan drag ke canvas */}
            <span
              draggable
              title="Tahan untuk mengubah urutan"
              className="flex cursor-grab items-center self-stretch rounded px-0.5 text-slate-300 hover:bg-slate-100 hover:text-slate-500 active:cursor-grabbing"
              onDragStart={(e) => {
                e.stopPropagation()
                e.dataTransfer.setData(REORDER_DRAG_TYPE, item.type)
                e.dataTransfer.effectAllowed = "move"
                setDragType(REORDER_DRAG_TYPE)
              }}
              onDragEnd={() => setDragType(null)}
            >
              <GripVertical className="h-3.5 w-3.5" />
            </span>
            <span
              className="flex h-5 w-5 items-center justify-center rounded-md text-white"
              style={{ backgroundColor: item.color }}
            >
              <Icon className="h-3 w-3" />
            </span>
            <span className="font-medium">{item.label}</span>
          </div>
        )
      })}
    </div>
  )
}
