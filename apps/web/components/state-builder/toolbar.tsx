"use client"

import {
  CheckCircle2,
  Download,
  FilePlus2,
  LayoutGrid,
  Redo2,
  Save,
  ShieldCheck,
  Undo2,
  Upload,
} from "lucide-react"
import { useMemo } from "react"
import type { WorkflowValidationResult } from "./types/workflow"

interface ToolbarStats {
  states: number
  decisions: number
  events: number
  transitions: number
  total: number
}

interface ExampleWorkflow {
  slug: string
  name: string
}

interface ToolbarProps {
  validation: WorkflowValidationResult | null
  onValidate: () => void
  onAutoLayout: () => void
  onSave: () => void
  onExport: () => void
  onImport: () => void
  onNew: () => void
  onReset: (slug: string) => void
  examples: ExampleWorkflow[]
  onUndo: () => void
  onRedo: () => void
  canUndo: boolean
  canRedo: boolean
  isSaving: boolean
  lastSavedAt: string | null
  stats: ToolbarStats
  onNewStart: () => void
  hasNode: (kind: string) => boolean
}

export function Toolbar({
  validation,
  onValidate,
  onAutoLayout,
  onSave,
  onExport,
  onImport,
  onNew,
  onReset,
  onUndo,
  onRedo,
  canUndo,
  canRedo,
  isSaving,
  lastSavedAt,
  stats,
  examples,
  onNewStart,
  hasNode,
}: ToolbarProps) {
  const errorCount = useMemo(
    () => validation?.issues.filter((i) => i.severity === "error").length ?? 0,
    [validation],
  )
  const warningCount = useMemo(
    () =>
      validation?.issues.filter((i) => i.severity === "warning").length ?? 0,
    [validation],
  )

  const statusColor = !validation
    ? "bg-slate-100 text-slate-500"
    : errorCount === 0
      ? "bg-success-50 text-success-600"
      : "bg-danger-50 text-danger-600"

  const saveLabel = isSaving
    ? "Menyimpan…"
    : lastSavedAt
      ? `Tersimpan ${new Date(lastSavedAt).toLocaleTimeString()}`
      : "Belum tersimpan"

  return (
    <div className="flex items-center justify-between gap-2 border-b border-slate-200 bg-white px-3 py-2">
      {/* Kiri: judul + status */}
      <div className="flex items-center gap-2">
        <span className="text-sm font-semibold text-slate-700">
          State Builder
        </span>
        {validation ? (
          <span
            className={`rounded-md px-2 py-0.5 text-xs font-medium ${statusColor}`}
          >
            {errorCount === 0
              ? `${warningCount > 0 ? `${warningCount} warnings` : "Valid"}`
              : `${errorCount} errors`}
          </span>
        ) : null}
        <span className="hidden items-center gap-1 text-xs text-slate-400 laptop:flex">
          {stats.states} state · {stats.decisions} decision · {stats.events}{" "}
          event · {stats.transitions} transisi
        </span>
      </div>

      {/* Kanan: aksi */}
      <div className="flex items-center gap-1.5">
        <span className="mr-1 hidden text-[11px] text-slate-400 tablet:inline">
          {saveLabel}
        </span>

        {!hasNode("START") ? (
          <button
            type="button"
            onClick={onNewStart}
            title="Tambah node START"
            className="flex items-center gap-1.5 rounded-md border border-primary-200 bg-primary-50 px-3 py-1.5 text-sm font-medium text-primary-700 hover:bg-primary-100"
          >
            <CheckCircle2 className="h-4 w-4" />
            Tambah Start
          </button>
        ) : null}

        <button
          type="button"
          onClick={onUndo}
          disabled={!canUndo}
          title="Undo (Ctrl+Z)"
          className="flex items-center gap-1 rounded-md border border-slate-200 px-2 py-1.5 text-sm text-slate-600 hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-40"
        >
          <Undo2 className="h-4 w-4" />
        </button>
        <button
          type="button"
          onClick={onRedo}
          disabled={!canRedo}
          title="Redo (Ctrl+Y)"
          className="flex items-center gap-1 rounded-md border border-slate-200 px-2 py-1.5 text-sm text-slate-600 hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-40"
        >
          <Redo2 className="h-4 w-4" />
        </button>

        <button
          type="button"
          onClick={onAutoLayout}
          className="flex items-center gap-1.5 rounded-md border border-slate-200 px-3 py-1.5 text-sm font-medium text-slate-600 hover:bg-slate-50"
        >
          <LayoutGrid className="h-4 w-4" />
          <span className="hidden laptop:inline">Auto Layout</span>
        </button>

        <button
          type="button"
          onClick={onImport}
          title="Import JSON"
          className="flex items-center gap-1.5 rounded-md border border-slate-200 px-3 py-1.5 text-sm font-medium text-slate-600 hover:bg-slate-50"
        >
          <Upload className="h-4 w-4" />
          <span className="hidden laptop:inline">Import</span>
        </button>
        <button
          type="button"
          onClick={onExport}
          className="flex items-center gap-1.5 rounded-md border border-slate-200 px-3 py-1.5 text-sm font-medium text-slate-600 hover:bg-slate-50"
        >
          <Download className="h-4 w-4" />
          <span className="hidden laptop:inline">Export</span>
        </button>

        <button
          type="button"
          onClick={onNew}
          title="Workflow baru"
          className="flex items-center gap-1.5 rounded-md border border-slate-200 px-2 py-1.5 text-sm text-slate-600 hover:bg-slate-50"
        >
          <FilePlus2 className="h-4 w-4" />
        </button>
        <select
          value=""
          onChange={(e) => {
            if (e.target.value) onReset(e.target.value)
          }}
          title="Muat contoh workflow"
          className="rounded-md border border-slate-200 px-2 py-1.5 text-sm text-slate-600 hover:bg-slate-50 focus:outline-none"
        >
          <option value="" disabled>
            Contoh…
          </option>
          {examples.map((ex) => (
            <option key={ex.slug} value={ex.slug}>
              {ex.name}
            </option>
          ))}
        </select>

        <button
          type="button"
          onClick={onValidate}
          className="flex items-center gap-1.5 rounded-md border border-slate-200 px-3 py-1.5 text-sm font-medium text-slate-600 hover:bg-slate-50"
        >
          <ShieldCheck className="h-4 w-4" />
          <span className="hidden laptop:inline">Validate</span>
        </button>

        <button
          type="button"
          onClick={onSave}
          className="flex items-center gap-1.5 rounded-md bg-primary-500 px-3 py-1.5 text-sm font-medium text-white hover:bg-primary-600"
        >
          <Save className="h-4 w-4" />
          Save
        </button>
      </div>
    </div>
  )
}
