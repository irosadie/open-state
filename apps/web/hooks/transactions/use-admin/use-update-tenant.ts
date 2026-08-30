import { tenantConfig } from "$/configs/tenant"
import { apiRouters, queryKeys } from "$/constants"
import { axios } from "$/services/axios"
import type { ErrorResponse } from "$/types/generals"
import type { UpdateTenantSchemaProps } from "@openstate/schemas"
import { tenantResponseSchema } from "@openstate/schemas"
import type { TenantResponse } from "@openstate/types"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import type { AxiosError } from "axios"

const updateTenant = async (payload: UpdateTenantSchemaProps) => {
  const result = await axios<TenantResponse>({
    method: "PATCH",
    url: apiRouters.admin.tenant,
    headers: { "X-Tenant-ID": tenantConfig.tenantId },
    data: payload,
  })
  return tenantResponseSchema.parse(result)
}

const useUpdateTenant = () => {
  const queryClient = useQueryClient()
  return useMutation<
    TenantResponse,
    ErrorResponse<AxiosError>,
    UpdateTenantSchemaProps
  >({
    mutationKey: [queryKeys.admin.updateTenant],
    mutationFn: updateTenant,
    onSuccess: () => {
      void queryClient.invalidateQueries({
        queryKey: [queryKeys.admin.tenant],
      })
    },
  })
}

export default useUpdateTenant
export { useUpdateTenant as useAdminTenantUpdate }
