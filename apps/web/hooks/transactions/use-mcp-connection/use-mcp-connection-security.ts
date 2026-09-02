import { tenantConfig } from "$/configs/tenant"
import { apiRouters, queryKeys } from "$/constants"
import { axios } from "$/services/axios"
import type { ErrorResponse } from "$/types/generals"
import {
  mcpConnectionResponseSchema,
  mcpOAuthStartResponseSchema,
  mcpOAuthStatusResponseSchema,
} from "@openstate/schemas"
import type {
  MCPConnectionResponse,
  MCPOAuthStartResponse,
  MCPOAuthStatusResponse,
} from "@openstate/types"
import { pathVariable } from "@openstate/utils"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import type { AxiosError } from "axios"

type Variables = { projectId: string; id: string }
type RotateVariables = Variables & { credentialValue: string }
type MutationError = ErrorResponse<AxiosError>

const invalidateConnections = (
  queryClient: ReturnType<typeof useQueryClient>,
  projectId: string,
) => {
  void queryClient.invalidateQueries({
    queryKey: [queryKeys.mcpConnections.list, projectId],
  })
}

export const useRotateMCPCredential = () => {
  const queryClient = useQueryClient()
  return useMutation<MCPConnectionResponse, MutationError, RotateVariables>({
    mutationKey: [queryKeys.mcpConnections.rotateCredential],
    mutationFn: async ({ projectId, id, credentialValue }) => {
      const result = await axios<MCPConnectionResponse>({
        method: "POST",
        url: pathVariable(apiRouters.mcpConnections.rotateCredential, {
          projectId,
          id,
        }),
        headers: { "X-Tenant-ID": tenantConfig.tenantId },
        data: { credentialValue },
      })
      return mcpConnectionResponseSchema.parse(result)
    },
    onSuccess: (_data, variables) =>
      invalidateConnections(queryClient, variables.projectId),
  })
}

export const useRevokeMCPCredential = () => {
  const queryClient = useQueryClient()
  return useMutation<MCPConnectionResponse, MutationError, Variables>({
    mutationKey: [queryKeys.mcpConnections.revokeCredential],
    mutationFn: async ({ projectId, id }) => {
      const result = await axios<MCPConnectionResponse>({
        method: "POST",
        url: pathVariable(apiRouters.mcpConnections.revokeCredential, {
          projectId,
          id,
        }),
        headers: { "X-Tenant-ID": tenantConfig.tenantId },
      })
      return mcpConnectionResponseSchema.parse(result)
    },
    onSuccess: (_data, variables) =>
      invalidateConnections(queryClient, variables.projectId),
  })
}

export const useStartMCPOAuth = () =>
  useMutation<MCPOAuthStartResponse, MutationError, Variables>({
    mutationKey: [queryKeys.mcpConnections.oauthStart],
    mutationFn: async ({ projectId, id }) => {
      const result = await axios<MCPOAuthStartResponse>({
        method: "POST",
        url: pathVariable(apiRouters.mcpConnections.oauthStart, {
          projectId,
          id,
        }),
        headers: { "X-Tenant-ID": tenantConfig.tenantId },
      })
      return mcpOAuthStartResponseSchema.parse(result)
    },
  })

export const useDisconnectMCPOAuth = () => {
  const queryClient = useQueryClient()
  return useMutation<MCPConnectionResponse, MutationError, Variables>({
    mutationKey: [queryKeys.mcpConnections.oauthDisconnect],
    mutationFn: async ({ projectId, id }) => {
      const result = await axios<MCPConnectionResponse>({
        method: "POST",
        url: pathVariable(apiRouters.mcpConnections.oauthDisconnect, {
          projectId,
          id,
        }),
        headers: { "X-Tenant-ID": tenantConfig.tenantId },
      })
      return mcpConnectionResponseSchema.parse(result)
    },
    onSuccess: (_data, variables) =>
      invalidateConnections(queryClient, variables.projectId),
  })
}

export const useMCPOAuthStatus = (
  projectId: string | undefined,
  id: string | undefined,
  enabled = true,
) =>
  useQuery<MCPOAuthStatusResponse, MutationError>({
    queryKey: [queryKeys.mcpConnections.oauthStatus, projectId, id],
    queryFn: async () => {
      const result = await axios<MCPOAuthStatusResponse>({
        method: "GET",
        url: pathVariable(apiRouters.mcpConnections.oauthStatus, {
          projectId,
          id,
        }),
        headers: { "X-Tenant-ID": tenantConfig.tenantId },
      })
      return mcpOAuthStatusResponseSchema.parse(result)
    },
    enabled: enabled && !!projectId && !!id,
  })
