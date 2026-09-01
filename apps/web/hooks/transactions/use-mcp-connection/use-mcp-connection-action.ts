import { tenantConfig } from "$/configs/tenant"
import { apiRouters, queryKeys } from "$/constants"
import { axios } from "$/services/axios"
import type { ErrorResponse } from "$/types/generals"
import { mcpConnectionResponseSchema } from "@openstate/schemas"
import type { MCPConnectionResponse } from "@openstate/types"
import { pathVariable } from "@openstate/utils"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import type { AxiosError } from "axios"

type Action = "enable" | "disable" | "test" | "delete"
type Variables = { projectId: string; id: string }

const useMCPConnectionAction = (action: Action) => {
  const queryClient = useQueryClient()
  return useMutation<
    MCPConnectionResponse,
    ErrorResponse<AxiosError>,
    Variables
  >({
    mutationKey: [queryKeys.mcpConnections[action]],
    mutationFn: async ({ projectId, id }) => {
      const router = apiRouters.mcpConnections[action]
      const result = await axios<MCPConnectionResponse>({
        method: action === "delete" ? "DELETE" : "POST",
        url: pathVariable(router, { projectId, id }),
        headers: { "X-Tenant-ID": tenantConfig.tenantId },
      })
      if (action === "delete") return { id, projectId } as MCPConnectionResponse
      return mcpConnectionResponseSchema.parse(result)
    },
    onSuccess: (_data, variables) => {
      void queryClient.invalidateQueries({
        queryKey: [queryKeys.mcpConnections.list, variables.projectId],
      })
    },
  })
}

export const useEnableMCPConnection = () => useMCPConnectionAction("enable")
export const useDisableMCPConnection = () => useMCPConnectionAction("disable")
export const useTestMCPConnection = () => useMCPConnectionAction("test")
export const useDeleteMCPConnection = () => useMCPConnectionAction("delete")
