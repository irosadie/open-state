"use client"

import Link from "next/link"
import { useState } from "react"

import { Button } from "$/components/button"
import { ContentTitle } from "$/components/content-title"
import { EmptyState } from "$/components/empty-state"
import { LoadingSpinner } from "$/components/loading-spinner"
import { PanelCard } from "$/components/panel-card"
import { StatusBadge } from "$/components/status-badge"
import { useRuntimeInstancesList } from "$/hooks/transactions/use-runtime-inspector"
import { useAuthorization } from "$/providers/authorization-provider"
import { getApiErrorStatus } from "$/utils/auth-error"
import { runtimeInstanceStatuses } from "@openstate/schemas"
import { ActivityIcon, ArrowRightIcon } from "lucide-react"

const formatTime = (value: string) => {
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? value : date.toLocaleString()
}

const statusVariant = (status: string) => {
  if (["RUNNING", "COMPLETED"].includes(status)) return "success" as const
  if (["FAILED", "CANCELLED", "ABORTED"].includes(status))
    return "danger" as const
  if (["WAITING", "SUSPENDED"].includes(status)) return "warning" as const
  return "neutral" as const
}

export default function RuntimeInstancesPageContent() {
  const authorization = useAuthorization()
  const [status, setStatus] = useState<string>("")
  const [workflowId, setWorkflowId] = useState("")
  const [correlationKey, setCorrelationKey] = useState("")
  const [page, setPage] = useState(1)
  const query = useRuntimeInstancesList({
    status: status || undefined,
    workflowId: workflowId || undefined,
    correlationKey: correlationKey || undefined,
    page,
    pageSize: 20,
    enabled:
      authorization.status === "ready" &&
      authorization.hasPermission("instance:read"),
  })

  if (
    authorization.status === "ready" &&
    !authorization.hasPermission("instance:read")
  ) {
    return (
      <div
        className="space-y-6 p-8"
        data-testid="runtime-inspector-access-denied"
      >
        <ContentTitle title="Runtime Inspector" />
        <div className="rounded-md bg-amber-50 px-4 py-3 text-sm text-amber-800">
          You are not authorized to inspect runtime instances.
        </div>
      </div>
    )
  }

  if (query.isLoading || authorization.status === "loading")
    return <LoadingSpinner />

  if (query.isError) {
    const forbidden = getApiErrorStatus(query.error) === 403
    return (
      <div className="space-y-6 p-8" data-testid="runtime-inspector-error">
        <ContentTitle title="Runtime Inspector" />
        <div className="rounded-md bg-red-50 px-4 py-3 text-sm text-red-700">
          {forbidden
            ? "Runtime instance access is not authorized."
            : "Runtime instances could not be loaded."}
          {!forbidden ? (
            <Button intent="clean" onClick={() => void query.refetch()}>
              Retry
            </Button>
          ) : null}
        </div>
      </div>
    )
  }

  const instances = query.data?.data ?? []
  return (
    <div className="space-y-6 p-8" data-testid="runtime-inspector-root">
      <ContentTitle title="Runtime Inspector" />
      <PanelCard
        title="Runtime instances"
        description="Tenant-scoped persisted workflow execution"
      >
        <div className="mb-6 flex flex-wrap items-end gap-4">
          <label className="flex min-w-48 flex-col gap-1 text-sm text-gray-600">
            Lifecycle status
            <select
              data-testid="runtime-status-filter"
              value={status}
              onChange={(event) => {
                setStatus(event.target.value)
                setPage(1)
              }}
              className="rounded-md border border-gray-300 bg-white px-3 py-2 text-sm text-gray-900"
            >
              <option value="">All statuses</option>
              {runtimeInstanceStatuses.map((value) => (
                <option key={value} value={value}>
                  {value}
                </option>
              ))}
            </select>
          </label>
          <label className="flex min-w-56 flex-1 flex-col gap-1 text-sm text-gray-600">
            Workflow id
            <input
              data-testid="runtime-workflow-filter"
              value={workflowId}
              onChange={(event) => {
                setWorkflowId(event.target.value)
                setPage(1)
              }}
              placeholder="Search workflow id"
              className="rounded-md border border-gray-300 px-3 py-2 text-sm text-gray-900"
            />
          </label>
          <label className="flex min-w-64 flex-1 flex-col gap-1 text-sm text-gray-600">
            Correlation key
            <input
              data-testid="runtime-correlation-filter"
              value={correlationKey}
              onChange={(event) => {
                setCorrelationKey(event.target.value)
                setPage(1)
              }}
              placeholder="Search correlation key"
              className="rounded-md border border-gray-300 px-3 py-2 text-sm text-gray-900"
            />
          </label>
        </div>

        {instances.length === 0 ? (
          <EmptyState
            icon={ActivityIcon}
            title="No runtime instances"
            description="No persisted workflow instances match the current filters."
          />
        ) : (
          <div className="overflow-x-auto rounded-lg border border-gray-200">
            <table className="min-w-full divide-y divide-gray-200 text-left text-sm">
              <thead className="bg-gray-50 text-xs uppercase text-gray-500">
                <tr>
                  <th className="px-4 py-3">Workflow / version</th>
                  <th className="px-4 py-3">Current state</th>
                  <th className="px-4 py-3">Lifecycle</th>
                  <th className="px-4 py-3">Last activity</th>
                  <th className="px-4 py-3">Correlation</th>
                  <th className="px-4 py-3" />
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200 bg-white">
                {instances.map((instance) => (
                  <tr
                    key={instance.id}
                    className="hover:bg-gray-50"
                    data-testid={`runtime-instance-row-${instance.id}`}
                  >
                    <td className="px-4 py-3">
                      <div className="font-medium text-gray-900">
                        {instance.workflow.name ||
                          instance.workflow.slug ||
                          instance.workflow.id}
                      </div>
                      <div className="text-xs text-gray-500">
                        v{instance.workflow.version || "—"}
                      </div>
                    </td>
                    <td className="px-4 py-3 text-gray-700">
                      {instance.currentState?.name ||
                        instance.currentState?.key ||
                        "Not recorded"}
                    </td>
                    <td className="px-4 py-3">
                      <StatusBadge variant={statusVariant(instance.status)}>
                        {instance.status}
                      </StatusBadge>
                    </td>
                    <td className="px-4 py-3 text-gray-600">
                      {formatTime(instance.lastActivityAt)}
                    </td>
                    <td className="max-w-48 truncate px-4 py-3 text-gray-600">
                      {instance.correlationId || "—"}
                    </td>
                    <td className="px-4 py-3 text-right">
                      <Link href={`/admin/runtime-instances/${instance.id}`}>
                        <Button
                          data-testid={`runtime-inspect-${instance.id}`}
                          intent="clean"
                          rightIcon={<ArrowRightIcon size={16} />}
                        >
                          Inspect
                        </Button>
                      </Link>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}

        <div className="mt-6 flex items-center justify-end gap-3 text-sm text-gray-600">
          <Button
            intent="clean"
            disabled={page <= 1}
            onClick={() => setPage((value) => Math.max(1, value - 1))}
          >
            Previous
          </Button>
          <span>Page {page}</span>
          <Button
            intent="clean"
            disabled={!query.data?.hasNext}
            onClick={() => setPage((value) => value + 1)}
          >
            Next
          </Button>
        </div>
      </PanelCard>
    </div>
  )
}
