"use client"

import Link from "next/link"
import { useState } from "react"

import { PermissionGate } from "$/components/auth-guard/permission-gate"
import { Button } from "$/components/button"
import { ContentTitle } from "$/components/content-title"
import { LoadingSpinner } from "$/components/loading-spinner"
import { PanelCard } from "$/components/panel-card"
import { StatusBadge } from "$/components/status-badge"
import { useAdminInstanceCommand } from "$/hooks/transactions/use-admin"
import {
  useRuntimeDebugTrace,
  useRuntimeInstanceGet,
} from "$/hooks/transactions/use-runtime-inspector"
import { useAuthorization } from "$/providers/authorization-provider"
import { extractErrorMessage } from "$/utils/api-error"
import { ArrowLeftIcon } from "lucide-react"
import { RuntimeDebugView } from "./_components/runtime-debug-view"

const formatTime = (value: string) => {
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? value : date.toLocaleString()
}

export default function RuntimeInstanceDetailPageContent({
  id,
}: { id: string }) {
  const authorization = useAuthorization()
  const [commandResult, setCommandResult] = useState<string | null>(null)
  const [commandError, setCommandError] = useState<string | null>(null)
  const suspend = useAdminInstanceCommand("suspend")
  const resume = useAdminInstanceCommand("resume")
  const retry = useAdminInstanceCommand("retry")
  const detail = useRuntimeInstanceGet({
    id,
    enabled:
      authorization.status === "ready" &&
      authorization.hasPermission("instance:read"),
  })
  const debug = useRuntimeDebugTrace({
    id,
    enabled:
      authorization.status === "ready" &&
      authorization.hasPermission("debug:read"),
  })

  if (authorization.status === "loading" || detail.isLoading)
    return <LoadingSpinner />
  if (
    authorization.status === "ready" &&
    !authorization.hasPermission("instance:read")
  ) {
    return (
      <div
        className="space-y-6 p-8"
        data-testid="runtime-instance-access-denied"
      >
        <ContentTitle title="Runtime Instance" />
        <div className="rounded-md bg-amber-50 px-4 py-3 text-sm text-amber-800">
          You are not authorized to inspect runtime instances.
        </div>
      </div>
    )
  }
  if (detail.isError || !detail.data) {
    return (
      <div className="space-y-6 p-8" data-testid="runtime-instance-error">
        <ContentTitle title="Runtime Instance" />
        <div className="rounded-md bg-red-50 px-4 py-3 text-sm text-red-700">
          {extractErrorMessage(detail.error) ||
            "Runtime instance could not be loaded."}
        </div>
        <Link href="/admin/runtime-instances">
          <Button intent="secondary" leftIcon={<ArrowLeftIcon size={16} />}>
            Back to Runtime Inspector
          </Button>
        </Link>
      </div>
    )
  }

  const { summary, context, timeline } = detail.data
  return (
    <div className="space-y-6 p-8" data-testid="runtime-instance-detail">
      <ContentTitle
        title={summary.workflow.name || "Runtime Instance"}
        breadcrumbData={[
          { title: "Runtime Inspector", link: "/admin/runtime-instances" },
        ]}
      />
      <Link href="/admin/runtime-instances">
        <Button intent="clean" leftIcon={<ArrowLeftIcon size={16} />}>
          Back to instances
        </Button>
      </Link>

      <div className="grid gap-6 lg:grid-cols-3">
        <div data-testid="runtime-workflow-summary">
          <PanelCard title="Workflow">
            <dl className="space-y-3 text-sm">
              <div>
                <dt className="text-gray-500">Workflow</dt>
                <dd className="font-medium">
                  {summary.workflow.slug || summary.workflow.id}
                </dd>
              </div>
              <div>
                <dt className="text-gray-500">Pinned version</dt>
                <dd className="font-medium">
                  v{summary.workflow.version || "—"}
                </dd>
              </div>
              <div>
                <dt className="text-gray-500">Instance id</dt>
                <dd className="break-all font-mono text-xs">{summary.id}</dd>
              </div>
              <div>
                <dt className="text-gray-500">Correlation</dt>
                <dd className="break-all">{summary.correlationId || "—"}</dd>
              </div>
            </dl>
          </PanelCard>
        </div>
        <div data-testid="runtime-current-state">
          <PanelCard title="Current state">
            {summary.currentState ? (
              <div className="space-y-3 text-sm">
                <div className="text-xl font-semibold text-gray-900">
                  {summary.currentState.name}
                </div>
                <div className="text-gray-500">{summary.currentState.key}</div>
                <StatusBadge variant="info">
                  {summary.currentState.status}
                </StatusBadge>
                <div className="text-gray-500">
                  Entered {formatTime(summary.currentState.enteredAt)}
                </div>
              </div>
            ) : (
              <div className="text-sm text-gray-500">
                Current state is not recorded.
              </div>
            )}
          </PanelCard>
        </div>
        <PanelCard title="Lifecycle">
          <div className="space-y-3 text-sm">
            <div data-testid="runtime-lifecycle-status">
              <StatusBadge variant="primary">{summary.status}</StatusBadge>
            </div>
            <div className="text-gray-500">Last activity</div>
            <div className="font-medium">
              {formatTime(summary.lastActivityAt)}
            </div>
            {detail.data.auditCorrelationIds.length > 0 ? (
              <div className="break-all text-xs text-gray-500">
                Audit correlations: {detail.data.auditCorrelationIds.join(", ")}
              </div>
            ) : null}
          </div>
        </PanelCard>
      </div>

      <PermissionGate permission="instance:suspend">
        <PanelCard
          title="Lifecycle commands"
          description="Confirmed tenant-scoped runtime actions"
        >
          <div className="flex flex-wrap items-center gap-3">
            {summary.status === "RUNNING" ? (
              <Button
                data-testid="runtime-command-suspend"
                intent="secondary"
                loading={suspend.isPending}
                onClick={() => {
                  if (!window.confirm("Suspend this runtime instance?")) return
                  setCommandError(null)
                  setCommandResult(null)
                  suspend.mutate(
                    { id },
                    {
                      onSuccess: (result) =>
                        setCommandResult(`SUSPENDED: ${result.status}`),
                      onError: (error) =>
                        setCommandError(error.message || "Suspend failed."),
                    },
                  )
                }}
              >
                Suspend
              </Button>
            ) : null}
            {summary.status === "SUSPENDED" ? (
              <Button
                data-testid="runtime-command-resume"
                intent="secondary"
                loading={resume.isPending}
                onClick={() => {
                  if (!window.confirm("Resume this runtime instance?")) return
                  setCommandError(null)
                  setCommandResult(null)
                  resume.mutate(
                    { id },
                    {
                      onSuccess: (result) =>
                        setCommandResult(`RESUMED: ${result.status}`),
                      onError: (error) =>
                        setCommandError(error.message || "Resume failed."),
                    },
                  )
                }}
              >
                Resume
              </Button>
            ) : null}
            {summary.status === "FAILED" ? (
              <Button
                data-testid="runtime-command-retry"
                intent="secondary"
                loading={retry.isPending}
                onClick={() => {
                  if (!window.confirm("Retry this runtime instance?")) return
                  setCommandError(null)
                  setCommandResult(null)
                  retry.mutate(
                    { id },
                    {
                      onSuccess: (result) =>
                        setCommandResult(`RETRIED: ${result.status}`),
                      onError: (error) =>
                        setCommandError(error.message || "Retry failed."),
                    },
                  )
                }}
              >
                Retry
              </Button>
            ) : null}
          </div>
          {commandResult ? (
            <div
              className="mt-3 rounded-md bg-emerald-50 px-3 py-2 text-sm text-emerald-800"
              data-testid="runtime-command-result"
            >
              {commandResult}
            </div>
          ) : null}
          {commandError ? (
            <div
              className="mt-3 rounded-md bg-red-50 px-3 py-2 text-sm text-red-700"
              data-testid="runtime-command-error"
            >
              {commandError}
            </div>
          ) : null}
        </PanelCard>
      </PermissionGate>

      <div className="grid gap-6 lg:grid-cols-2">
        <div data-testid="runtime-context">
          <PanelCard
            title="Sanitized context"
            description="Available and missing context for the current state"
          >
            {context.redacted ? (
              <div className="mb-3 rounded bg-amber-50 px-3 py-2 text-xs text-amber-800">
                Sensitive values are redacted.
              </div>
            ) : null}
            <div className="space-y-2 text-sm">
              {Object.entries(context.available).length === 0 ? (
                <div className="text-gray-500">No available context.</div>
              ) : (
                Object.entries(context.available).map(([key, value]) => (
                  <div
                    key={key}
                    className="flex items-start justify-between gap-4 border-b border-gray-100 py-2"
                  >
                    <span className="font-mono text-xs text-gray-600">
                      {key}
                    </span>
                    <span className="max-w-[60%] break-words text-right text-gray-900">
                      {typeof value === "string"
                        ? value
                        : JSON.stringify(value)}
                    </span>
                  </div>
                ))
              )}
            </div>
            {context.missing.length > 0 ? (
              <div className="mt-4 text-sm text-amber-700">
                Missing: {context.missing.join(", ")}
              </div>
            ) : null}
          </PanelCard>
        </div>
        <div data-testid="runtime-timeline">
          <PanelCard
            title="Runtime timeline"
            description="Chronological state, event, and decision activity"
          >
            {timeline.length === 0 ? (
              <div className="text-sm text-gray-500">
                No timeline activity has been recorded.
              </div>
            ) : (
              <ol className="space-y-3">
                {timeline.map((entry) => (
                  <li
                    key={`${entry.kind}-${entry.id}`}
                    className="border-l-2 border-gray-200 pl-4 text-sm"
                    data-testid={`runtime-timeline-entry-${entry.kind.toLowerCase()}-${entry.id}`}
                  >
                    <div className="flex flex-wrap items-center gap-2">
                      <StatusBadge
                        variant={
                          entry.kind === "DECISION" ? "warning" : "neutral"
                        }
                      >
                        {entry.kind}
                      </StatusBadge>
                      <span className="font-medium text-gray-900">
                        {entry.label}
                      </span>
                      {entry.reasonCode ? (
                        <span className="font-mono text-xs text-gray-600">
                          {entry.reasonCode}
                        </span>
                      ) : null}
                    </div>
                    <div className="mt-1 text-xs text-gray-500">
                      {formatTime(entry.occurredAt)} · {entry.status}
                    </div>
                  </li>
                ))}
              </ol>
            )}
          </PanelCard>
        </div>
      </div>

      <PermissionGate
        action="debug:read"
        fallback={
          <PanelCard
            title="Debug View"
            description="OpenState-owned, sanitized per-turn evidence"
          >
            <div
              className="rounded-md bg-amber-50 px-4 py-3 text-sm text-amber-800"
              data-testid="runtime-debug-access-denied"
            >
              Debug evidence is not authorized for this tenant role. Runtime
              detail remains available.
            </div>
          </PanelCard>
        }
      >
        <RuntimeDebugView query={debug} />
      </PermissionGate>
    </div>
  )
}
