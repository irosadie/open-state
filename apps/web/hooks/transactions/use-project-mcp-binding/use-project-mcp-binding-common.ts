import { tenantConfig } from "$/configs/tenant"
import { apiRouters } from "$/constants"
import { axios } from "$/services/axios"
import {
  mcpToolOptionListResponseSchema,
  projectCapabilityMCPBindingListResponseSchema,
  projectCapabilityMCPBindingResponseSchema,
} from "@openstate/schemas"
import type {
  MCPToolOptionListResponse,
  ProjectCapabilityMCPBindingListResponse,
  ProjectCapabilityMCPBindingResponse,
} from "@openstate/types"
import { pathVariable } from "@openstate/utils"

const projectHeaders = { "X-Tenant-ID": tenantConfig.tenantId }

export const listProjectMCPToolOptions = async (
  projectId: string,
): Promise<MCPToolOptionListResponse> => {
  const result = await axios<MCPToolOptionListResponse>({
    method: "GET",
    url: pathVariable(apiRouters.projectMCPBindings.options, { projectId }),
    headers: projectHeaders,
  })
  return mcpToolOptionListResponseSchema.parse(result)
}

export const listProjectMCPBindings = async (
  projectId: string,
): Promise<ProjectCapabilityMCPBindingListResponse> => {
  const result = await axios<ProjectCapabilityMCPBindingListResponse>({
    method: "GET",
    url: pathVariable(apiRouters.projectMCPBindings.list, { projectId }),
    headers: projectHeaders,
  })
  return projectCapabilityMCPBindingListResponseSchema.parse(result)
}

export type ProjectMCPBindingMutation = {
  projectId: string
  capabilityId: string
  connectionId: string
  toolId: string
}

export const upsertProjectMCPBinding = async ({
  projectId,
  capabilityId,
  connectionId,
  toolId,
}: ProjectMCPBindingMutation): Promise<ProjectCapabilityMCPBindingResponse> => {
  const result = await axios<ProjectCapabilityMCPBindingResponse>({
    method: "PUT",
    url: pathVariable(apiRouters.projectMCPBindings.upsert, {
      projectId,
      capabilityId,
    }),
    headers: projectHeaders,
    data: { connectionId, toolId },
  })
  return projectCapabilityMCPBindingResponseSchema.parse(result)
}

export const deleteProjectMCPBinding = async ({
  projectId,
  capabilityId,
}: Pick<ProjectMCPBindingMutation, "projectId" | "capabilityId">) => {
  await axios<{ message: string }>({
    method: "DELETE",
    url: pathVariable(apiRouters.projectMCPBindings.delete, {
      projectId,
      capabilityId,
    }),
    headers: projectHeaders,
  })
}
