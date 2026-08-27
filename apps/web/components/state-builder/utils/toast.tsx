"use client"

import { CheckCircle2, Info, TriangleAlert, XCircle } from "lucide-react"
import { useEffect, useState } from "react"

type ToastKind = "success" | "error" | "info" | "warning"

interface ToastItem {
  id: number
  kind: ToastKind
  message: string
}

const KIND_STYLE: Record<ToastKind, string> = {
  success: "border-success-300 bg-success-50 text-success-800",
  error: "border-danger-300 bg-danger-50 text-danger-800",
  info: "border-primary-300 bg-primary-50 text-primary-800",
  warning: "border-warning-300 bg-warning-50 text-warning-800",
}

const KIND_ICON: Record<ToastKind, typeof Info> = {
  success: CheckCircle2,
  error: XCircle,
  info: Info,
  warning: TriangleAlert,
}

let toastCounter = 0
let pushToast: ((kind: ToastKind, message: string) => void) | null = null

/** API global untuk memunculkan toast dari mana saja */
export const toast = {
  success: (m: string) => pushToast?.("success", m),
  error: (m: string) => pushToast?.("error", m),
  info: (m: string) => pushToast?.("info", m),
  warning: (m: string) => pushToast?.("warning", m),
}

/** Provider toast — mount sekali di StateBuilder */
export function Toaster() {
  const [items, setItems] = useState<ToastItem[]>([])

  useEffect(() => {
    pushToast = (kind, message) => {
      const id = ++toastCounter
      setItems((prev) => [...prev, { id, kind, message }])
      setTimeout(() => {
        setItems((prev) => prev.filter((t) => t.id !== id))
      }, 3500)
    }
    return () => {
      pushToast = null
    }
  }, [])

  return (
    <ul className="pointer-events-none fixed right-4 top-4 z-[1000] flex w-80 flex-col gap-2">
      {items.map((t) => {
        const Icon = KIND_ICON[t.kind]
        return (
          <li
            key={t.id}
            className={`flex items-center gap-2 rounded-lg border px-3 py-2 text-sm shadow-lg ${KIND_STYLE[t.kind]}`}
          >
            <Icon className="h-4 w-4 shrink-0" />
            <span>{t.message}</span>
          </li>
        )
      })}
    </ul>
  )
}
