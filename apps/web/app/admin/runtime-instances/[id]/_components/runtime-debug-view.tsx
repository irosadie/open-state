"use client"

import { Button } from "$/components/button"
import { PanelCard } from "$/components/panel-card"
import { StatusBadge } from "$/components/status-badge"
import { extractErrorMessage } from "$/utils/api-error"
import type { RuntimeTraceResponse } from "@openstate/types"
import { BugIcon } from "lucide-react"

type RuntimeDebugQuery = {
  data?: RuntimeTraceResponse
  error: unknown
  isError: boolean
  isForbidden: boolean
  isLoading: boolean
  refetch: () => unknown
}

const formatTime = (value: string) => {
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? value : date.toLocaleString()
}

export function RuntimeDebugView({ query }: { query: RuntimeDebugQuery }) {
  return (
    <div data-testid="runtime-debug-view">
      <PanelCard
        title="Debug View"
        description="OpenState-owned, sanitized per-turn evidence"
      >
        {query.isForbidden ? (
          <div className="rounded-md bg-amber-50 px-4 py-3 text-sm text-amber-800">
            Debug evidence is not authorized for your role. Runtime detail
            remains available.
          </div>
        ) : query.isLoading ? (
          <div className="animate-pulse text-sm text-gray-500">
            Loading trace data…
          </div>
        ) : query.isError ? (
          <div className="rounded-md bg-red-50 px-4 py-3 text-sm text-red-700">
            {extractErrorMessage(query.error) ||
              "Debug evidence could not be loaded."}{" "}
            <Button intent="clean" onClick={() => void query.refetch()}>
              Retry
            </Button>
          </div>
        ) : !query.data?.available ? (
          <div className="flex items-center gap-3 text-sm text-gray-500">
            <BugIcon size={18} /> No trace has been recorded for this instance.
          </div>
        ) : (
          <div className="space-y-3">
            {query.data.data.map((entry) => (
              <div
                key={entry.id}
                className="rounded-lg border border-gray-200 p-4"
                data-testid={`runtime-debug-entry-${entry.id}`}
              >
                <div className="flex flex-wrap items-center gap-2">
                  <StatusBadge
                    variant={
                      entry.status === "FAILED"
                        ? "danger"
                        : entry.status === "SUCCEEDED"
                          ? "success"
                          : "neutral"
                    }
                  >
                    {entry.status}
                  </StatusBadge>
                  <span className="font-medium text-gray-900">
                    {entry.stage}
                  </span>
                  {entry.source === "EXTERNAL_PROVIDER" ? (
                    <StatusBadge variant="info">
                      External provider metadata
                    </StatusBadge>
                  ) : (
                    <StatusBadge variant="neutral">OpenState</StatusBadge>
                  )}
                </div>
                <div className="mt-2 grid gap-2 text-xs text-gray-600 sm:grid-cols-2">
                  <div>Occurred: {formatTime(entry.occurredAt)}</div>
                  <div>
                    Duration:{" "}
                    {entry.durationMs == null ? "—" : `${entry.durationMs} ms`}
                  </div>
                  <div>Correlation: {entry.correlationId || "—"}</div>
                  <div>Reason: {entry.reasonCode || "—"}</div>
                  {entry.providerAlias ? (
                    <div>Provider alias: {entry.providerAlias}</div>
                  ) : null}
                  {entry.providerReference ? (
                    <div>Operation reference: {entry.providerReference}</div>
                  ) : null}
                </div>
                {entry.summary ? (
                  <p className="mt-3 text-sm text-gray-700">{entry.summary}</p>
                ) : null}
                {Object.keys(entry.attributes).length > 0 ? (
                  <pre className="mt-3 overflow-x-auto rounded bg-gray-50 p-3 text-xs text-gray-700">
                    {JSON.stringify(entry.attributes, null, 2)}
                  </pre>
                ) : null}
              </div>
            ))}
          </div>
        )}
      </PanelCard>
    </div>
  )
}
