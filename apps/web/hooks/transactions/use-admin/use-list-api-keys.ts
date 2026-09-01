import { tenantConfig } from "$/configs/tenant"
import { apiRouters, queryKeys } from "$/constants"
import { axios } from "$/services/axios"
import type { ErrorResponse } from "$/types/generals"
import { stateMCPAPIKeyListResponseSchema } from "@openstate/schemas"
import type { StateMCPAPIKeyListResponse } from "@openstate/types"
import { useQuery } from "@tanstack/react-query"
import type { AxiosError } from "axios"

const listAPIKeys = async () => {
  const result = await axios<StateMCPAPIKeyListResponse>({
    method: "GET",
    url: apiRouters.admin.apiKeys,
    headers: { "X-Tenant-ID": tenantConfig.tenantId },
  })

  return stateMCPAPIKeyListResponseSchema.parse(result)
}

const useListAPIKeys = (enabled = true) =>
  useQuery<StateMCPAPIKeyListResponse, ErrorResponse<AxiosError>>({
    queryKey: [queryKeys.admin.apiKeys, tenantConfig.tenantId],
    queryFn: listAPIKeys,
    enabled,
  })

export default useListAPIKeys
export { useListAPIKeys as useAdminAPIKeys }
