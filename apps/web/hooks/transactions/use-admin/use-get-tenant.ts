import { tenantConfig } from "$/configs/tenant"
import { apiRouters, queryKeys } from "$/constants"
import { axios } from "$/services/axios"
import type { ErrorResponse } from "$/types/generals"
import { tenantResponseSchema } from "@openstate/schemas"
import type { TenantResponse } from "@openstate/types"
import { useQuery } from "@tanstack/react-query"
import type { AxiosError } from "axios"

const getTenant = async () => {
  const result = await axios<TenantResponse>({
    method: "GET",
    url: apiRouters.admin.tenant,
    headers: { "X-Tenant-ID": tenantConfig.tenantId },
  })
  return tenantResponseSchema.parse(result)
}

const useGetTenant = (enabled = true) => {
  return useQuery<TenantResponse, ErrorResponse<AxiosError>>({
    queryKey: [queryKeys.admin.tenant],
    queryFn: getTenant,
    enabled,
  })
}

export default useGetTenant
export { useGetTenant as useAdminTenant }
