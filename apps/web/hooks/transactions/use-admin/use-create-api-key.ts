import { tenantConfig } from "$/configs/tenant"
import { apiRouters, queryKeys } from "$/constants"
import { axios } from "$/services/axios"
import type { ErrorResponse } from "$/types/generals"
import {
  type CreateStateMCPAPIKeySchemaProps,
  createStateMCPAPIKeyResponseSchema,
} from "@openstate/schemas"
import type { CreateStateMCPAPIKeyResponse } from "@openstate/types"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import type { AxiosError } from "axios"

const createAPIKey = async (payload: CreateStateMCPAPIKeySchemaProps) => {
  const result = await axios<CreateStateMCPAPIKeyResponse>({
    method: "POST",
    url: apiRouters.admin.apiKeys,
    headers: { "X-Tenant-ID": tenantConfig.tenantId },
    data: payload,
  })

  return createStateMCPAPIKeyResponseSchema.parse(result)
}

const useCreateAPIKey = () => {
  const queryClient = useQueryClient()

  return useMutation<
    CreateStateMCPAPIKeyResponse,
    ErrorResponse<AxiosError>,
    CreateStateMCPAPIKeySchemaProps
  >({
    mutationKey: [queryKeys.admin.createAPIKey],
    mutationFn: createAPIKey,
    onSuccess: () => {
      void queryClient.invalidateQueries({
        queryKey: [queryKeys.admin.apiKeys],
      })
    },
  })
}

export default useCreateAPIKey
export { useCreateAPIKey as useAdminAPIKeyCreate }
