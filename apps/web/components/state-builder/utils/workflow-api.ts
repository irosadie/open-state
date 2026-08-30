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
  WorkflowDiffResponse,
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
  definition: unknown
}): Promise<WorkflowResponse> {
  const headers: Record<string, string> = {
    "X-Tenant-ID": tenantConfig.tenantId,
  }
  if (input.projectId) headers["X-Project-ID"] = input.projectId
  return axios<WorkflowResponse>({
    method: "POST",
    url: apiRouters.workflows.create,
    headers,
    data: input,
  })
}

/**
 * Persist the complete editable draft under optimistic concurrency (PRD §31).
 */
export async function updateWorkflowApi(input: {
  id: string
  version: number
  name: string
  description?: string
  definition: unknown
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
    data: {
      version: input.version,
      name: input.name,
      description: input.description,
      definition: input.definition,
    },
  })
}

export async function getWorkflowApi(input: {
  id: string
  projectId?: string
}): Promise<WorkflowResponse> {
  const headers: Record<string, string> = {
    "X-Tenant-ID": tenantConfig.tenantId,
  }
  if (input.projectId) headers["X-Project-ID"] = input.projectId
  return axios<WorkflowResponse>({
    method: "GET",
    url: pathVariable(apiRouters.workflows.show, { id: input.id }),
    headers,
  })
}

/**
 * Publish the current server-side draft to an immutable, current version.
 */
export async function publishWorkflowApi(input: {
  id: string
  version: number
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
    data: { version: input.version },
  })
}

export async function listWorkflowVersionsApi(input: {
  id: string
  projectId?: string
}): Promise<WorkflowVersionResponse[]> {
  const headers: Record<string, string> = {
    "X-Tenant-ID": tenantConfig.tenantId,
  }
  if (input.projectId) headers["X-Project-ID"] = input.projectId
  return axios<WorkflowVersionResponse[]>({
    method: "GET",
    url: pathVariable(apiRouters.workflows.versions, { id: input.id }),
    headers,
  })
}

export async function compareWorkflowVersionsApi(input: {
  id: string
  baseVersion: number
  targetVersion: number
  projectId?: string
}): Promise<WorkflowDiffResponse> {
  const headers: Record<string, string> = {
    "X-Tenant-ID": tenantConfig.tenantId,
  }
  if (input.projectId) headers["X-Project-ID"] = input.projectId
  return axios<WorkflowDiffResponse>({
    method: "GET",
    url: pathVariable(apiRouters.workflows.compare, { id: input.id }),
    headers,
    params: {
      baseVersion: input.baseVersion,
      targetVersion: input.targetVersion,
    },
  })
}
