import { tenantConfig } from "$/configs/tenant"
import { axios } from "$/services/axios"
import type { ErrorResponse } from "$/types/generals"
import {
  mcpConnectionListResponseSchema,
  mcpConnectionResponseSchema,
  mcpDiscoveredToolResponseSchema,
  mcpToolCatalogResponseSchema,
} from "@openstate/schemas"
import type {
  MCPConnectionListResponse,
  MCPConnectionResponse,
  MCPDiscoveredToolResponse,
  MCPToolCatalogResponse,
} from "@openstate/types"
import type { AxiosError } from "axios"

export type MCPConnectionMutation = {
  projectId: string
  id?: string
}

export type MCPConnectionError = ErrorResponse<AxiosError>
export type MCPToolCatalogError = ErrorResponse<AxiosError>

export const listMCPConnections = async (projectId: string) => {
  const result = await axios<MCPConnectionListResponse>({
    method: "GET",
    url: `/projects/${encodeURIComponent(projectId)}/mcp-connections`,
    headers: { "X-Tenant-ID": tenantConfig.tenantId },
  })
  return mcpConnectionListResponseSchema.parse(result)
}

export const parseMCPConnection = (result: unknown): MCPConnectionResponse =>
  mcpConnectionResponseSchema.parse(result)

export const listMCPTools = async (
  projectId: string,
  connectionId: string,
): Promise<MCPToolCatalogResponse> => {
  const result = await axios<MCPToolCatalogResponse>({
    method: "GET",
    url: `/projects/${encodeURIComponent(projectId)}/mcp-connections/${encodeURIComponent(connectionId)}/tools`,
    headers: { "X-Tenant-ID": tenantConfig.tenantId },
  })
  return mcpToolCatalogResponseSchema.parse(result)
}

export const parseMCPToolCatalog = (result: unknown): MCPToolCatalogResponse =>
  mcpToolCatalogResponseSchema.parse(result)

export const parseMCPDiscoveredTool = (
  result: unknown,
): MCPDiscoveredToolResponse => mcpDiscoveredToolResponseSchema.parse(result)
