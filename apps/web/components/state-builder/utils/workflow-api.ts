/**
 * Thin API service for the State Builder's workflow-definition operations
 * (PRD 146). The Zustand store uses these directly (a store is not a JSX
 * component, so this is not the forbidden "axios in components" case); the
 * React-query hooks in `apps/web/hooks/transactions/use-workflow/` are the
 * React-facing API for data-driven views.
 */
import { tenantConfig } from "$/configs/tenant"
import { apiRouters } from "$/constants"
import { axios } from "$/services/axios"
import type {
  WorkflowResponse,
  WorkflowVersionResponse,
} from "@openstate/types"
import { pathVariable } from "@openstate/utils"

/**
 * Create a workflow definition root (status=DRAFT). Returns the persisted
 * workflow with its authoritative id + optimistic version.
 */
export async function createWorkflowApi(input: {
  slug: string
  name: string
  description?: string
  projectId?: string
}): Promise<WorkflowResponse> {
  return axios<WorkflowResponse>({
    method: "POST",
    url: apiRouters.workflows.create,
    headers: { "X-Tenant-ID": tenantConfig.tenantId },
    data: input,
  })
}

/**
 * Bump the optimistic version of a DRAFT workflow root (PRD §31).
 */
export async function updateWorkflowApi(input: {
  id: string
  version: number
  projectId?: string
}): Promise<WorkflowResponse> {
  const headers: Record<string, string> = {
    "X-Tenant-ID": tenantConfig.tenantId,
  }
  if (input.projectId) headers["X-Project-ID"] = input.projectId

  return axios<WorkflowResponse>({
    method: "PATCH",
    url: pathVariable(apiRouters.workflows.update, { id: input.id }),
    headers,
    data: { version: input.version },
  })
}

/**
 * Publish a workflow definition to an immutable, current version (PRD §3.3, §9).
 * The `definition` is the full WorkflowDefinition envelope (PRD §68).
 */
export async function publishWorkflowApi(input: {
  id: string
  version: number
  definition: unknown
  projectId?: string
}): Promise<WorkflowVersionResponse> {
  const headers: Record<string, string> = {
    "X-Tenant-ID": tenantConfig.tenantId,
  }
  if (input.projectId) headers["X-Project-ID"] = input.projectId

  return axios<WorkflowVersionResponse>({
    method: "POST",
    url: pathVariable(apiRouters.workflows.publish, { id: input.id }),
    headers,
    data: { version: input.version, definition: input.definition },
  })
}
