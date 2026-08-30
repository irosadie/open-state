import { tenantConfig } from "$/configs/tenant"
import { apiRouters, queryKeys } from "$/constants"
import { axios } from "$/services/axios"
import type { ErrorResponse } from "$/types/generals"
import { tenantMembershipResponseSchema } from "@openstate/schemas"
import type { TenantMembershipResponse, TenantRole } from "@openstate/types"
import { pathVariable } from "@openstate/utils"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import type { AxiosError } from "axios"

type UpdateMemberRoleArgs = { userId: string; role: TenantRole }

const updateMemberRole = async ({ userId, role }: UpdateMemberRoleArgs) => {
  const result = await axios<TenantMembershipResponse>({
    method: "PUT",
    url: pathVariable(apiRouters.admin.memberRole, { userId }),
    headers: { "X-Tenant-ID": tenantConfig.tenantId },
    data: { role },
  })
  return tenantMembershipResponseSchema.parse(result)
}

const useUpdateMemberRole = () => {
  const queryClient = useQueryClient()
  return useMutation<
    TenantMembershipResponse,
    ErrorResponse<AxiosError>,
    UpdateMemberRoleArgs
  >({
    mutationKey: [queryKeys.admin.updateMemberRole],
    mutationFn: updateMemberRole,
    onSuccess: () => {
      void queryClient.invalidateQueries({
        queryKey: [queryKeys.admin.members],
      })
      void queryClient.invalidateQueries({ queryKey: [queryKeys.auth.me] })
    },
  })
}

export default useUpdateMemberRole
export { useUpdateMemberRole as useAdminMemberRoleUpdate }
