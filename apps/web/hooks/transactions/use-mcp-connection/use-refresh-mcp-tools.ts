import { tenantConfig } from "$/configs/tenant"
import { apiRouters, queryKeys } from "$/constants"
import { axios } from "$/services/axios"
import type { ErrorResponse } from "$/types/generals"
import { mcpToolCatalogResponseSchema } from "@openstate/schemas"
import type { MCPToolCatalogResponse } from "@openstate/types"
import { pathVariable } from "@openstate/utils"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import type { AxiosError } from "axios"

type Variables = { projectId: string; id: string }

const refreshMCPTools = async ({ projectId, id }: Variables) => {
  const result = await axios<MCPToolCatalogResponse>({
    method: "POST",
    url: pathVariable(apiRouters.mcpConnections.refreshTools, {
      projectId,
      id,
    }),
    headers: { "X-Tenant-ID": tenantConfig.tenantId },
  })
  return mcpToolCatalogResponseSchema.parse(result)
}

const useRefreshMCPTools = () => {
  const queryClient = useQueryClient()
  return useMutation<
    MCPToolCatalogResponse,
    ErrorResponse<AxiosError>,
    Variables
  >({
    mutationKey: [queryKeys.mcpConnections.refreshTools],
    mutationFn: refreshMCPTools,
    onSuccess: (_data, variables) => {
      void queryClient.invalidateQueries({
        queryKey: [
          queryKeys.mcpConnections.tools,
          variables.projectId,
          variables.id,
        ],
      })
    },
    onError: (_error, variables) => {
      void queryClient.invalidateQueries({
        queryKey: [
          queryKeys.mcpConnections.tools,
          variables.projectId,
          variables.id,
        ],
      })
    },
  })
}

export default useRefreshMCPTools
export { useRefreshMCPTools }
