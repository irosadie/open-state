"use client"

import Link from "next/link"
import { useState } from "react"

import { Button } from "$/components/button"
import { ContentTitle } from "$/components/content-title"
import { LoadingSpinner } from "$/components/loading-spinner"
import { PanelCard } from "$/components/panel-card"
import { StatusBadge } from "$/components/status-badge"
import {
  useAdminInstanceCommand,
  useAdminInstances,
} from "$/hooks/transactions/use-admin"
import { useAuthorization } from "$/providers/authorization-provider"
import { extractErrorMessage } from "$/utils/api-error"
import type { InstanceResponse } from "@openstate/types"

type Command = "suspend" | "resume" | "retry"
type Notice = { type: "success" | "error"; text: string }

const commandLabels: Record<Command, string> = {
  suspend: "Suspend",
  resume: "Resume",
  retry: "Retry",
}

const statusVariant = (status: string) => {
  if (["RUNNING", "COMPLETED"].includes(status)) return "success" as const
  if (["FAILED", "CANCELLED", "ABORTED"].includes(status))
    return "danger" as const
  if (["SUSPENDED", "WAITING"].includes(status)) return "warning" as const
  return "neutral" as const
}

const formatTime = (value: string) => {
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? value : date.toLocaleString()
}

export default function InstancesPageContent() {
  const authorization = useAuthorization()
  const canRead = authorization.hasPermission("instance:read")
  const instances = useAdminInstances(
    authorization.status === "ready" && canRead,
  )
  const suspend = useAdminInstanceCommand("suspend")
  const resume = useAdminInstanceCommand("resume")
  const retry = useAdminInstanceCommand("retry")
  const [notice, setNotice] = useState<Notice | null>(null)

  if (!canRead) {
    return (
      <div className="space-y-6 p-8">
        <ContentTitle title="Instance operations" />
        <div className="rounded-md bg-amber-50 px-4 py-3 text-sm text-amber-800">
          You are not authorized to view runtime instances.
        </div>
      </div>
    )
  }
  if (instances.isLoading) return <LoadingSpinner />

  const runCommand = (command: Command, instance: InstanceResponse) => {
    if (!window.confirm(`${commandLabels[command]} this instance?`)) return
    setNotice(null)
    const mutation =
      command === "suspend" ? suspend : command === "resume" ? resume : retry
    mutation.mutate(
      { id: instance.id },
      {
        onSuccess: () =>
          setNotice({
            type: "success",
            text: `${commandLabels[command]} request accepted.`,
          }),
        onError: (error) =>
          setNotice({
            type: "error",
            text:
              extractErrorMessage(error) ??
              `${commandLabels[command]} request was rejected.`,
          }),
      },
    )
  }

  return (
    <div className="space-y-6 p-8">
      <ContentTitle title="Instance operations" />
      {notice ? (
        <output
          className={`rounded-md px-4 py-3 text-sm ${notice.type === "success" ? "bg-green-50 text-green-700" : "bg-red-50 text-red-700"}`}
        >
          {notice.text}
        </output>
      ) : null}
      <PanelCard
        title="Runtime instances"
        description="Commands are tenant-scoped and validated against the current lifecycle state."
      >
        {instances.isError ? (
          <div className="rounded-md bg-red-50 px-4 py-3 text-sm text-red-700">
            {extractErrorMessage(instances.error) ??
              "Instances could not be loaded."}
          </div>
        ) : (
          <div className="overflow-x-auto rounded-lg border border-slate-200">
            <table className="min-w-full divide-y divide-slate-200 text-left text-sm">
              <thead className="bg-slate-50 text-xs uppercase text-slate-500">
                <tr>
                  <th className="px-4 py-3">Instance</th>
                  <th className="px-4 py-3">Workflow</th>
                  <th className="px-4 py-3">Status</th>
                  <th className="px-4 py-3">Updated</th>
                  <th className="px-4 py-3">Actions</th>
                  <th className="px-4 py-3" />
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-200 bg-white">
                {(instances.data ?? []).map((instance) => (
                  <tr key={instance.id}>
                    <td className="px-4 py-3">
                      <div className="font-mono text-xs text-slate-900">
                        {instance.id}
                      </div>
                      <div className="text-xs text-slate-500">
                        v{instance.version}
                      </div>
                    </td>
                    <td className="px-4 py-3 text-slate-700">
                      {instance.workflowId}
                    </td>
                    <td className="px-4 py-3">
                      <StatusBadge variant={statusVariant(instance.status)}>
                        {instance.status}
                      </StatusBadge>
                    </td>
                    <td className="px-4 py-3 text-slate-600">
                      {formatTime(instance.updatedAt)}
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex flex-wrap gap-2">
                        {authorization.hasPermission("instance:suspend") &&
                        instance.status !== "SUSPENDED" ? (
                          <Button
                            intent="clean"
                            onClick={() => runCommand("suspend", instance)}
                            loading={suspend.isPending}
                          >
                            Suspend
                          </Button>
                        ) : null}
                        {authorization.hasPermission("instance:resume") &&
                        instance.status === "SUSPENDED" ? (
                          <Button
                            intent="clean"
                            onClick={() => runCommand("resume", instance)}
                            loading={resume.isPending}
                          >
                            Resume
                          </Button>
                        ) : null}
                        {authorization.hasPermission("instance:retry") &&
                        instance.status === "FAILED" ? (
                          <Button
                            intent="clean"
                            onClick={() => runCommand("retry", instance)}
                            loading={retry.isPending}
                          >
                            Retry
                          </Button>
                        ) : null}
                      </div>
                    </td>
                    <td className="px-4 py-3 text-right">
                      <Link
                        className="text-sm font-medium text-slate-700 underline"
                        href={`/admin/runtime-instances/${instance.id}`}
                      >
                        Inspect
                      </Link>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
            {(instances.data ?? []).length === 0 ? (
              <p className="px-4 py-8 text-center text-sm text-slate-500">
                No runtime instances found.
              </p>
            ) : null}
          </div>
        )}
      </PanelCard>
    </div>
  )
}
