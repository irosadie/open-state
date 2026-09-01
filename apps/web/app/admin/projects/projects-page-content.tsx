"use client"

import Link from "next/link"
import { useSearchParams } from "next/navigation"

import { AdminFlowGuide } from "$/components/admin-console"
import { Button } from "$/components/button"
import { ContentTitle } from "$/components/content-title"
import { LoadingSpinner } from "$/components/loading-spinner"
import { PanelCard } from "$/components/panel-card"
import { tenantConfig } from "$/configs/tenant"
import { useProjectsList } from "$/hooks/transactions/use-project"
import { useAuthorization } from "$/providers/authorization-provider"
import { extractErrorMessage } from "$/utils/api-error"
import { ArrowRightIcon, FolderIcon } from "lucide-react"

const scopedHref = (path: string, projectId: string) =>
  `${path}?projectId=${encodeURIComponent(projectId)}`

export default function ProjectsPageContent() {
  const authorization = useAuthorization()
  const canRead = authorization.hasPermission("workflow:read")
  const selectedProjectId = useSearchParams().get("projectId") || undefined
  const projects = useProjectsList({
    enabled: authorization.status === "ready" && canRead,
  })
  const selectedProject = projects.data?.find(
    (project) => project.id === selectedProjectId,
  )

  if (!canRead) {
    return (
      <div className="space-y-6 p-8">
        <ContentTitle title="Project inventory" />
        <div className="rounded-md bg-amber-50 px-4 py-3 text-sm text-amber-800">
          You are not authorized to view projects.
        </div>
      </div>
    )
  }

  if (projects.isLoading) return <LoadingSpinner />

  return (
    <div className="space-y-6 p-8">
      <ContentTitle title="Project inventory" />
      <p className="max-w-3xl text-sm text-slate-600">
        Choose the project that scopes its Intents, Workflows, and States. A
        project selection is preserved in the URL so the flow stays explicit.
      </p>
      <AdminFlowGuide
        currentStep="project"
        projectId={selectedProject?.id ?? selectedProjectId}
        projectName={selectedProject?.name}
      />

      <PanelCard
        title="Current tenant"
        description="Only projects owned by this tenant are available."
      >
        <p className="break-all font-mono text-xs text-slate-600">
          {tenantConfig.tenantId}
        </p>
      </PanelCard>

      <PanelCard
        title="Projects"
        description="Select a project to continue to its intent and workflow scope."
      >
        {projects.isError ? (
          <div className="flex items-center justify-between gap-4 rounded-md bg-red-50 px-4 py-3 text-sm text-red-700">
            <span>
              {extractErrorMessage(projects.error) ??
                "Projects could not be loaded."}
            </span>
            <Button intent="clean" onClick={() => void projects.refetch()}>
              Retry
            </Button>
          </div>
        ) : projects.data?.length ? (
          <div className="grid gap-4 md:grid-cols-2">
            {projects.data.map((project) => (
              <article
                key={project.id}
                className={`rounded-lg border p-5 transition-shadow hover:shadow-md ${
                  project.id === selectedProjectId
                    ? "border-slate-900 bg-slate-50"
                    : "border-slate-200"
                }`}
              >
                <div className="flex items-start gap-3">
                  <FolderIcon
                    className="mt-0.5 shrink-0 text-slate-500"
                    size={20}
                    aria-hidden="true"
                  />
                  <div className="min-w-0">
                    <h3 className="font-semibold text-slate-950">
                      {project.name}
                    </h3>
                    <p className="mt-1 text-sm text-slate-600">
                      {project.slug}
                    </p>
                    <p className="mt-2 break-all font-mono text-xs text-slate-500">
                      {project.id}
                    </p>
                    <span className="mt-3 inline-flex rounded-full bg-emerald-100 px-2 py-1 text-xs font-medium text-emerald-700">
                      {project.status}
                    </span>
                  </div>
                </div>
                <div className="mt-5 flex flex-wrap gap-2">
                  <Link href={scopedHref("/admin/intents", project.id)}>
                    <Button
                      intent="primary"
                      rightIcon={<ArrowRightIcon size={16} />}
                    >
                      Use project
                    </Button>
                  </Link>
                  <Link href={scopedHref("/admin/workflows", project.id)}>
                    <Button intent="clean">Open Workflows</Button>
                  </Link>
                  <Link href={scopedHref("/admin/mcp", project.id)}>
                    <Button intent="clean">MCP Connections</Button>
                  </Link>
                </div>
              </article>
            ))}
          </div>
        ) : (
          <div className="rounded-md border border-dashed border-slate-300 px-4 py-10 text-center">
            <p className="font-medium text-slate-900">No projects found</p>
            <p className="mt-1 text-sm text-slate-500">
              This tenant does not have a project available for the flow yet.
            </p>
          </div>
        )}
      </PanelCard>
    </div>
  )
}
