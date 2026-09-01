"use client"

import Link from "next/link"

import { AdminFlowGuide } from "$/components/admin-console"
import { Button } from "$/components/button"
import { ContentTitle } from "$/components/content-title"
import { LoadingSpinner } from "$/components/loading-spinner"
import { PanelCard } from "$/components/panel-card"
import { useWorkflowsList } from "$/hooks/transactions/use-workflow"
import { useAuthorization } from "$/providers/authorization-provider"
import { extractErrorMessage } from "$/utils/api-error"
import { ArrowRightIcon, PlusIcon } from "lucide-react"
import { useSearchParams } from "next/navigation"
import { useState } from "react"
import { CreateWorkflowDialog } from "./_components/create-workflow-dialog"

export default function WorkflowsPageContent() {
  const authorization = useAuthorization()
  const projectId = useSearchParams().get("projectId") || undefined
  const canRead = authorization.hasPermission("workflow:read")
  const canCreate = authorization.hasPermission("workflow:create")
  const [isCreateOpen, setIsCreateOpen] = useState(false)
  const workflows = useWorkflowsList({
    projectId,
    enabled: authorization.status === "ready" && canRead,
  })

  if (!canRead) {
    return (
      <div className="space-y-6 p-8">
        <ContentTitle title="Workflow inventory" />
        <div className="rounded-md bg-amber-50 px-4 py-3 text-sm text-amber-800">
          You are not authorized to view workflows.
        </div>
      </div>
    )
  }
  if (workflows.isLoading) return <LoadingSpinner />

  return (
    <div className="space-y-6 p-8">
      <div className="flex items-center justify-between">
        <ContentTitle title="Workflow inventory" />
        {canCreate ? (
          <Button
            leftIcon={<PlusIcon size={16} />}
            onClick={() => setIsCreateOpen(true)}
          >
            New Workflow
          </Button>
        ) : null}
      </div>
      <AdminFlowGuide currentStep="workflow" projectId={projectId} />
      <PanelCard
        title={`Workflows in ${projectId ? "Selected Project" : "Default Project"}`}
        description={
          projectId
            ? `These workflows belong to project ${projectId}. Review Intent mappings, then open a workflow in Builder for authoring.`
            : "These workflows belong to the current tenant's Default Project. Review Intent mappings, then open a workflow in Builder for authoring."
        }
      >
        {workflows.isError ? (
          <div className="rounded-md bg-red-50 px-4 py-3 text-sm text-red-700">
            {extractErrorMessage(workflows.error) ??
              "Workflows could not be loaded."}
            <Button intent="clean" onClick={() => void workflows.refetch()}>
              Retry
            </Button>
          </div>
        ) : (
          <div className="overflow-x-auto rounded-lg border border-slate-200">
            <table className="min-w-full divide-y divide-slate-200 text-left text-sm">
              <thead className="bg-slate-50 text-xs uppercase text-slate-500">
                <tr>
                  <th className="px-4 py-3">Workflow</th>
                  <th className="px-4 py-3">Status</th>
                  <th className="px-4 py-3">Current version</th>
                  <th className="px-4 py-3" />
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-200 bg-white">
                {(workflows.data ?? []).map((workflow) => (
                  <tr key={workflow.id}>
                    <td className="px-4 py-3">
                      <div className="font-medium text-slate-900">
                        {workflow.name || workflow.slug}
                      </div>
                      <div className="text-xs text-slate-500">
                        {workflow.slug}
                      </div>
                    </td>
                    <td className="px-4 py-3 text-slate-600">
                      {workflow.status}
                    </td>
                    <td className="px-4 py-3 text-slate-600">
                      v{workflow.currentVersion || workflow.version}
                    </td>
                    <td className="px-4 py-3 text-right">
                      <Link
                        href={`/state-builder/${workflow.id}${
                          projectId
                            ? `?projectId=${encodeURIComponent(projectId)}`
                            : ""
                        }`}
                      >
                        <Button
                          intent="clean"
                          rightIcon={<ArrowRightIcon size={16} />}
                        >
                          Open Builder
                        </Button>
                      </Link>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
            {(workflows.data ?? []).length === 0 ? (
              <p className="px-4 py-8 text-center text-sm text-slate-500">
                No workflows found for this tenant.
              </p>
            ) : null}
          </div>
        )}
      </PanelCard>

      <CreateWorkflowDialog
        open={isCreateOpen}
        projectId={projectId}
        onCancel={() => setIsCreateOpen(false)}
      />
    </div>
  )
}
