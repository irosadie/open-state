"use client"

import { PermissionGate } from "$/components/auth-guard/permission-gate"
import { getSimulationOutcomeLabel } from "@openstate/schemas"
import type {
  SimulationResultResponse,
  SimulationStepResponse,
} from "@openstate/types"
import {
  AlertTriangle,
  CheckCircle2,
  CircleDot,
  FlaskConical,
  Play,
  Plus,
  RotateCcw,
  X,
} from "lucide-react"
import type { SimulationDraftEvent } from "./state-builder.store"

type SimulationRunInput = {
  initialContext: Record<string, unknown>
  events: Array<{ type: string; payload: Record<string, unknown> }>
}

interface SimulationPanelProps {
  initialContextText: string
  events: SimulationDraftEvent[]
  result: SimulationResultResponse | null
  error: string | null
  isRunning: boolean
  stale: boolean
  selectedSequence: number | null
  onInitialContextChange: (value: string) => void
  onAddEvent: () => void
  onUpdateEvent: (
    id: string,
    patch: Partial<Pick<SimulationDraftEvent, "type" | "payloadText">>,
  ) => void
  onRemoveEvent: (id: string) => void
  onRun: (input: SimulationRunInput) => void
  onReset: () => void
  onSelectStep: (sequence: number | null) => void
  onClose: () => void
}

function parseJsonObject(value: string, label: string) {
  try {
    const parsed: unknown = JSON.parse(value)
    if (
      parsed === null ||
      typeof parsed !== "object" ||
      Array.isArray(parsed)
    ) {
      return { value: null, error: `${label} harus berupa object JSON.` }
    }
    return { value: parsed as Record<string, unknown>, error: null }
  } catch {
    return { value: null, error: `${label} bukan JSON yang valid.` }
  }
}

function stateLabel(step: SimulationStepResponse) {
  if (step.sequence === 0) return step.stateBefore.name
  return `${step.stateBefore.name} → ${step.stateAfter?.name ?? "ditolak"}`
}

function TraceStep({
  step,
  selected,
  onSelect,
}: {
  step: SimulationStepResponse
  selected: boolean
  onSelect: () => void
}) {
  const rejected = step.outcome === "REJECTED"
  return (
    <button
      type="button"
      onClick={onSelect}
      className={`w-full rounded-md border p-2 text-left transition-colors ${
        selected
          ? "border-primary-400 bg-primary-50"
          : rejected
            ? "border-danger-200 bg-danger-50/50 hover:bg-danger-50"
            : "border-slate-200 bg-white hover:bg-slate-50"
      }`}
    >
      <div className="flex items-start gap-2">
        {rejected ? (
          <AlertTriangle className="mt-0.5 h-4 w-4 shrink-0 text-danger-600" />
        ) : (
          <CheckCircle2 className="mt-0.5 h-4 w-4 shrink-0 text-success-600" />
        )}
        <div className="min-w-0 flex-1">
          <div className="flex items-center justify-between gap-2">
            <span className="text-xs font-semibold text-slate-700">
              {step.sequence === 0 ? "Entry" : `Event ${step.sequence}`}
            </span>
            <span
              className={`text-[10px] font-semibold uppercase ${
                rejected ? "text-danger-700" : "text-success-700"
              }`}
            >
              {getSimulationOutcomeLabel(step.outcome)}
            </span>
          </div>
          {step.eventType ? (
            <p className="mt-0.5 truncate text-xs text-slate-500">
              {step.eventType}
            </p>
          ) : null}
          <p className="mt-1 text-xs text-slate-600">{stateLabel(step)}</p>
          {step.selectedTransitionId ? (
            <p className="mt-1 text-[11px] text-primary-700">
              Transition: {step.selectedTransitionId}
            </p>
          ) : null}
          {step.errorMessage ? (
            <p className="mt-1 text-[11px] text-danger-700">
              {step.errorMessage}
            </p>
          ) : null}
          {step.candidates.length > 0 ? (
            <div className="mt-1 space-y-0.5">
              <p className="text-[11px] text-slate-400">
                Guards:{" "}
                {
                  step.candidates.filter((candidate) => candidate.guardPassed)
                    .length
                }
                /{step.candidates.length} passed
              </p>
              {step.candidates.map((candidate) => (
                <p
                  key={candidate.transitionId}
                  className="truncate text-[10px] text-slate-500"
                >
                  {candidate.guardPassed ? "✓" : "✗"} {candidate.transitionId} ·
                  priority {candidate.priority}
                  {candidate.guardError ? ` · ${candidate.guardError}` : ""}
                </p>
              ))}
            </div>
          ) : null}
          {step.capabilityRequests.length > 0 ? (
            <div className="mt-1 flex flex-wrap gap-1">
              {step.capabilityRequests.map((request) => (
                <span
                  key={request.name}
                  className="rounded bg-violet-50 px-1.5 py-0.5 text-[10px] font-medium text-violet-700"
                >
                  Mock · {request.name}
                </span>
              ))}
            </div>
          ) : null}
          <div className="mt-1 text-[10px] text-slate-400">
            <p>Context setelah step</p>
            <pre className="mt-1 max-h-20 overflow-auto whitespace-pre-wrap break-words font-mono text-[10px] text-slate-500">
              {JSON.stringify(step.context, null, 2)}
            </pre>
          </div>
        </div>
      </div>
    </button>
  )
}

export function SimulationPanel({
  initialContextText,
  events,
  result,
  error,
  isRunning,
  stale,
  selectedSequence,
  onInitialContextChange,
  onAddEvent,
  onUpdateEvent,
  onRemoveEvent,
  onRun,
  onReset,
  onSelectStep,
  onClose,
}: SimulationPanelProps) {
  const context = parseJsonObject(initialContextText, "Initial context")
  const eventErrors = events.map((event) => ({
    id: event.id,
    typeError: event.type.trim() ? null : "Event type wajib diisi.",
    payload: parseJsonObject(
      event.payloadText,
      `Payload ${event.type || "event"}`,
    ),
  }))
  const hasErrors =
    Boolean(context.error) ||
    eventErrors.some((item) => item.typeError || item.payload.error)

  const handleRun = () => {
    if (hasErrors || !context.value) return
    onRun({
      initialContext: context.value,
      events: events.map((event) => ({
        type: event.type.trim(),
        payload:
          eventErrors.find((item) => item.id === event.id)?.payload.value ?? {},
      })),
    })
  }

  return (
    <aside className="absolute right-3 top-3 z-30 flex max-h-[calc(100%-1.5rem)] w-[min(390px,calc(100%-1.5rem))] flex-col overflow-hidden rounded-lg border border-slate-200 bg-white shadow-xl">
      <div className="flex items-center justify-between border-b border-slate-200 px-4 py-3">
        <div className="flex items-center gap-2">
          <FlaskConical className="h-4 w-4 text-violet-600" />
          <div>
            <h3 className="text-sm font-semibold text-slate-800">
              Simulation sandbox
            </h3>
            <p className="text-[11px] text-slate-400">
              Tidak menyimpan atau menjalankan provider live
            </p>
          </div>
        </div>
        <button
          type="button"
          onClick={onClose}
          className="rounded p-1 text-slate-400 hover:bg-slate-100 hover:text-slate-600"
          aria-label="Tutup simulation"
        >
          <X className="h-4 w-4" />
        </button>
      </div>

      <div className="sb-scroll space-y-4 overflow-y-auto p-4">
        <label className="block text-xs font-medium text-slate-600">
          Initial context (JSON object)
          <textarea
            value={initialContextText}
            onChange={(event) => onInitialContextChange(event.target.value)}
            rows={4}
            spellCheck={false}
            className={`mt-1 w-full rounded-md border px-2 py-1.5 font-mono text-xs outline-none focus:ring-1 ${context.error ? "border-danger-300 focus:border-danger-400 focus:ring-danger-200" : "border-slate-200 focus:border-primary-400 focus:ring-primary-200"}`}
          />
          {context.error ? (
            <span className="mt-1 block text-[11px] text-danger-600">
              {context.error}
            </span>
          ) : null}
        </label>

        <section className="space-y-2">
          <div className="flex items-center justify-between">
            <p className="text-xs font-medium text-slate-600">
              Events ({events.length}/100)
            </p>
            <button
              type="button"
              onClick={onAddEvent}
              disabled={events.length >= 100}
              className="flex items-center gap-1 rounded border border-slate-200 px-2 py-1 text-[11px] font-medium text-slate-600 hover:bg-slate-50 disabled:opacity-40"
            >
              <Plus className="h-3.5 w-3.5" /> Tambah event
            </button>
          </div>
          {events.length === 0 ? (
            <p className="rounded bg-slate-50 px-2 py-2 text-[11px] text-slate-500">
              Belum ada event. Kamu bisa menjalankan entry state saja.
            </p>
          ) : null}
          {events.map((event, index) => {
            const itemError = eventErrors.find((item) => item.id === event.id)
            return (
              <div
                key={event.id}
                className="rounded-md border border-slate-200 p-2"
              >
                <div className="mb-1 flex items-center justify-between gap-2">
                  <span className="text-[11px] font-semibold text-slate-500">
                    #{index + 1}
                  </span>
                  <button
                    type="button"
                    onClick={() => onRemoveEvent(event.id)}
                    className="rounded p-1 text-slate-400 hover:bg-danger-50 hover:text-danger-600"
                    aria-label={`Hapus event ${index + 1}`}
                  >
                    <X className="h-3.5 w-3.5" />
                  </button>
                </div>
                <input
                  value={event.type}
                  onChange={(input) =>
                    onUpdateEvent(event.id, { type: input.target.value })
                  }
                  placeholder="event.type"
                  className={`w-full rounded border px-2 py-1 text-xs outline-none focus:ring-1 ${itemError?.typeError ? "border-danger-300 focus:ring-danger-200" : "border-slate-200 focus:border-primary-400 focus:ring-primary-200"}`}
                />
                {itemError?.typeError ? (
                  <p className="mt-1 text-[11px] text-danger-600">
                    {itemError.typeError}
                  </p>
                ) : null}
                <textarea
                  value={event.payloadText}
                  onChange={(input) =>
                    onUpdateEvent(event.id, { payloadText: input.target.value })
                  }
                  rows={3}
                  spellCheck={false}
                  placeholder={'{"key": "value"}'}
                  className={`mt-1 w-full rounded border px-2 py-1 font-mono text-xs outline-none focus:ring-1 ${itemError?.payload.error ? "border-danger-300 focus:ring-danger-200" : "border-slate-200 focus:border-primary-400 focus:ring-primary-200"}`}
                />
                {itemError?.payload.error ? (
                  <p className="mt-1 text-[11px] text-danger-600">
                    {itemError.payload.error}
                  </p>
                ) : null}
              </div>
            )
          })}
        </section>

        {error ? (
          <p className="rounded-md bg-danger-50 px-2 py-2 text-xs text-danger-700">
            {error}
          </p>
        ) : null}
        {stale && result ? (
          <p className="flex items-start gap-1.5 rounded-md bg-warning-50 px-2 py-2 text-xs text-warning-700">
            <AlertTriangle className="mt-0.5 h-3.5 w-3.5 shrink-0" />
            Hasil ini stale karena workflow berubah setelah simulation.
          </p>
        ) : null}

        <div className="flex gap-2">
          <PermissionGate action="workflow:simulate">
            <button
              type="button"
              onClick={handleRun}
              disabled={isRunning || hasErrors}
              className="flex flex-1 items-center justify-center gap-1.5 rounded-md bg-violet-600 px-3 py-2 text-xs font-semibold text-white hover:bg-violet-700 disabled:cursor-not-allowed disabled:opacity-50"
            >
              <Play className="h-3.5 w-3.5" />
              {isRunning ? "Menjalankan…" : "Run simulation"}
            </button>
          </PermissionGate>
          <button
            type="button"
            onClick={onReset}
            className="flex items-center gap-1 rounded-md border border-slate-200 px-3 py-2 text-xs font-medium text-slate-600 hover:bg-slate-50"
          >
            <RotateCcw className="h-3.5 w-3.5" /> Reset
          </button>
        </div>

        {result ? (
          <section className="space-y-2 border-t border-slate-200 pt-3">
            <div className="flex items-center justify-between">
              <p className="text-xs font-semibold text-slate-700">Trace</p>
              <span
                className={`rounded px-1.5 py-0.5 text-[10px] font-semibold ${result.finalStatus === "COMPLETED" ? "bg-success-50 text-success-700" : "bg-slate-100 text-slate-600"}`}
              >
                {result.finalStatus}
              </span>
            </div>
            <div className="space-y-1.5">
              {result.steps.map((step) => (
                <TraceStep
                  key={step.sequence}
                  step={step}
                  selected={selectedSequence === step.sequence}
                  onSelect={() => onSelectStep(step.sequence)}
                />
              ))}
            </div>
            <div className="rounded-md bg-slate-50 p-2">
              <p className="mb-1 flex items-center gap-1 text-[11px] font-semibold text-slate-600">
                <CircleDot className="h-3 w-3" /> Final context
              </p>
              <pre className="max-h-28 overflow-auto whitespace-pre-wrap break-words font-mono text-[10px] text-slate-500">
                {JSON.stringify(result.finalContext, null, 2)}
              </pre>
            </div>
          </section>
        ) : null}
      </div>
    </aside>
  )
}

export type { SimulationRunInput }
