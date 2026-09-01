"use client"

import Link from "next/link"

import { AdminFlowGuide } from "$/components/admin-console"
import { Button } from "$/components/button"
import { ContentTitle } from "$/components/content-title"
import { LoadingSpinner } from "$/components/loading-spinner"
import { PanelCard } from "$/components/panel-card"
import { tenantConfig } from "$/configs/tenant"
import { useIntentsList } from "$/hooks/transactions/use-intent"
import { useAuthorization } from "$/providers/authorization-provider"
import { extractErrorMessage } from "$/utils/api-error"
import { ArrowRightIcon, MessageSquareIcon } from "lucide-react"
import { useSearchParams } from "next/navigation"

export default function IntentsPageContent() {
  const authorization = useAuthorization()
  const projectId = useSearchParams().get("projectId") || undefined
  const canRead = authorization.hasPermission("workflow:read")
  const intents = useIntentsList({
    projectId,
    enabled: authorization.status === "ready" && canRead,
  })

  if (!canRead) {
    return (
      <div className="space-y-6 p-8">
        <ContentTitle title="Intent catalog" />
        <div className="rounded-md bg-amber-50 px-4 py-3 text-sm text-amber-800">
          You are not authorized to view intents.
        </div>
      </div>
    )
  }

  if (intents.isLoading) return <LoadingSpinner />

  return (
    <div className="space-y-6 p-8">
      <ContentTitle title="Intent catalog" />
      <p className="max-w-3xl text-sm text-slate-600">
        Intents are the routing choices that connect natural-language requests
        to published workflows. This catalog is read-only.
      </p>
      <AdminFlowGuide currentStep="intent" projectId={projectId} />

      <PanelCard
        title="Current catalog scope"
        description={
          projectId
            ? "The catalog follows the project selected in the flow."
            : "The catalog follows the tenant context and the existing Default Project used by the console."
        }
      >
        <div className="grid gap-4 text-sm sm:grid-cols-2">
          <div>
            <p className="font-semibold text-slate-900">Tenant</p>
            <p className="mt-1 break-all font-mono text-xs text-slate-600">
              {tenantConfig.tenantId}
            </p>
          </div>
          <div>
            <p className="font-semibold text-slate-900">Project</p>
            <p className="mt-1 text-slate-600">
              {projectId ? "Selected Project" : "Default Project"}{" "}
              <span className="break-all font-mono text-xs text-slate-500">
                ({projectId ?? "automatic"})
              </span>
            </p>
          </div>
        </div>
      </PanelCard>

      <PanelCard
        title="Published intents"
        description="Only active-project mappings to published workflows appear here."
      >
        {intents.isError ? (
          <div className="flex items-center justify-between gap-4 rounded-md bg-red-50 px-4 py-3 text-sm text-red-700">
            <span>
              {extractErrorMessage(intents.error) ??
                "Intents could not be loaded."}
            </span>
            <Button intent="clean" onClick={() => void intents.refetch()}>
              Retry
            </Button>
          </div>
        ) : intents.data?.length ? (
          <div className="overflow-x-auto rounded-lg border border-slate-200">
            <table className="min-w-full divide-y divide-slate-200 text-left text-sm">
              <thead className="bg-slate-50 text-xs uppercase text-slate-500">
                <tr>
                  <th className="px-4 py-3">Intent</th>
                  <th className="px-4 py-3">Examples</th>
                  <th className="px-4 py-3">Mapped workflow</th>
                  <th className="px-4 py-3" />
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-200 bg-white">
                {intents.data.map((intent) => (
                  <tr key={intent.id}>
                    <td className="px-4 py-4 align-top">
                      <div className="flex items-start gap-2">
                        <MessageSquareIcon
                          className="mt-0.5 shrink-0 text-slate-500"
                          size={16}
                          aria-hidden="true"
                        />
                        <div>
                          <div className="font-mono text-xs font-semibold text-slate-900">
                            {intent.key}
                          </div>
                          <div className="mt-1 font-medium text-slate-900">
                            {intent.name}
                          </div>
                          <div className="mt-1 max-w-sm text-xs text-slate-500">
                            {intent.description || "No description"}
                          </div>
                        </div>
                      </div>
                    </td>
                    <td className="px-4 py-4 align-top">
                      {intent.examples.length ? (
                        <ul className="max-w-md space-y-1 text-slate-600">
                          {intent.examples.map((example) => (
                            <li key={example}>&ldquo;{example}&rdquo;</li>
                          ))}
                        </ul>
                      ) : (
                        <span className="text-slate-500">No examples</span>
                      )}
                    </td>
                    <td className="px-4 py-4 align-top">
                      <div className="font-mono text-xs text-slate-900">
                        {intent.workflowSlug}
                      </div>
                      <div className="mt-1 text-xs text-slate-500">
                        Published workflow
                      </div>
                    </td>
                    <td className="px-4 py-4 text-right align-top">
                      <Link
                        href={`/state-builder/${intent.workflowId}${
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
          </div>
        ) : (
          <div className="rounded-md border border-dashed border-slate-300 px-4 py-10 text-center">
            <p className="font-medium text-slate-900">
              No published intents yet
            </p>
            <p className="mt-1 text-sm text-slate-500">
              Publish an intent-to-workflow mapping for this project to make it
              available to LLM routing.
            </p>
          </div>
        )}
      </PanelCard>
    </div>
  )
}
