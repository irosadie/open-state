import { tenantConfig } from "$/configs/tenant"
import { apiRouters, queryKeys } from "$/constants"
import { axios } from "$/services/axios"
import type { ErrorResponse } from "$/types/generals"
import { stateMCPAPIKeyResponseSchema } from "@openstate/schemas"
import type { StateMCPAPIKeyResponse } from "@openstate/types"
import { pathVariable } from "@openstate/utils"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import type { AxiosError } from "axios"

const revokeAPIKey = async (id: string) => {
  const result = await axios<StateMCPAPIKeyResponse>({
    method: "POST",
    url: pathVariable(apiRouters.admin.revokeAPIKey, { id }),
    headers: { "X-Tenant-ID": tenantConfig.tenantId },
  })

  return stateMCPAPIKeyResponseSchema.parse(result)
}

const useRevokeAPIKey = () => {
  const queryClient = useQueryClient()

  return useMutation<StateMCPAPIKeyResponse, ErrorResponse<AxiosError>, string>(
    {
      mutationKey: [queryKeys.admin.revokeAPIKey],
      mutationFn: revokeAPIKey,
      onSuccess: () => {
        void queryClient.invalidateQueries({
          queryKey: [queryKeys.admin.apiKeys],
        })
      },
    },
  )
}

export default useRevokeAPIKey
export { useRevokeAPIKey as useAdminAPIKeyRevoke }
