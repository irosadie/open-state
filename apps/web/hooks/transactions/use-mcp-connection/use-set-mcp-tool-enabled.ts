import { tenantConfig } from "$/configs/tenant"
import { apiRouters, queryKeys } from "$/constants"
import { axios } from "$/services/axios"
import type { ErrorResponse } from "$/types/generals"
import { mcpDiscoveredToolResponseSchema } from "@openstate/schemas"
import type { MCPDiscoveredToolResponse } from "@openstate/types"
import { pathVariable } from "@openstate/utils"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import type { AxiosError } from "axios"

type Variables = {
  projectId: string
  id: string
  toolName: string
  enabled: boolean
}

const setMCPToolEnabled = async ({
  projectId,
  id,
  toolName,
  enabled,
}: Variables) => {
  const result = await axios<MCPDiscoveredToolResponse>({
    method: "PATCH",
    url: pathVariable(apiRouters.mcpConnections.updateTool, {
      projectId,
      id,
      toolName: encodeURIComponent(toolName),
    }),
    headers: { "X-Tenant-ID": tenantConfig.tenantId },
    data: { enabled },
  })
  return mcpDiscoveredToolResponseSchema.parse(result)
}

const useSetMCPToolEnabled = () => {
  const queryClient = useQueryClient()
  return useMutation<
    MCPDiscoveredToolResponse,
    ErrorResponse<AxiosError>,
    Variables
  >({
    mutationKey: [queryKeys.mcpConnections.updateTool],
    mutationFn: setMCPToolEnabled,
    onSuccess: (_data, variables) => {
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

export default useSetMCPToolEnabled
export { useSetMCPToolEnabled }
