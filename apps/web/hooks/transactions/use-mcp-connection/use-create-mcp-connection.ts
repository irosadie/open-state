import { tenantConfig } from "$/configs/tenant"
import { apiRouters, queryKeys } from "$/constants"
import { axios } from "$/services/axios"
import type { ErrorResponse } from "$/types/generals"
import type { CreateMCPConnectionSchemaProps } from "@openstate/schemas"
import { mcpConnectionResponseSchema } from "@openstate/schemas"
import type { MCPConnectionResponse } from "@openstate/types"
import { pathVariable } from "@openstate/utils"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import type { AxiosError } from "axios"

type Variables = { projectId: string; payload: CreateMCPConnectionSchemaProps }

const createMCPConnection = async ({ projectId, payload }: Variables) => {
  const result = await axios<MCPConnectionResponse>({
    method: "POST",
    url: pathVariable(apiRouters.mcpConnections.create, { projectId }),
    headers: { "X-Tenant-ID": tenantConfig.tenantId },
    data: payload,
  })
  return mcpConnectionResponseSchema.parse(result)
}

const useCreateMCPConnection = () => {
  const queryClient = useQueryClient()
  return useMutation<
    MCPConnectionResponse,
    ErrorResponse<AxiosError>,
    Variables
  >({
    mutationKey: [queryKeys.mcpConnections.create],
    mutationFn: createMCPConnection,
    onSuccess: (_data, variables) => {
      void queryClient.invalidateQueries({
        queryKey: [queryKeys.mcpConnections.list, variables.projectId],
      })
    },
  })
}

export default useCreateMCPConnection
