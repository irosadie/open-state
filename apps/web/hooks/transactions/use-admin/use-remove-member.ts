import { tenantConfig } from "$/configs/tenant"
import { apiRouters, queryKeys } from "$/constants"
import { axios } from "$/services/axios"
import type { ErrorResponse } from "$/types/generals"
import { pathVariable } from "@openstate/utils"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import type { AxiosError } from "axios"

const removeMember = async (userId: string) => {
  await axios<{ message: string }>({
    method: "DELETE",
    url: pathVariable(apiRouters.admin.member, { userId }),
    headers: { "X-Tenant-ID": tenantConfig.tenantId },
  })
  return userId
}

const useRemoveMember = () => {
  const queryClient = useQueryClient()
  return useMutation<string, ErrorResponse<AxiosError>, string>({
    mutationKey: [queryKeys.admin.removeMember],
    mutationFn: removeMember,
    onSuccess: () => {
      void queryClient.invalidateQueries({
        queryKey: [queryKeys.admin.members],
      })
      void queryClient.invalidateQueries({ queryKey: [queryKeys.auth.me] })
    },
  })
}

export default useRemoveMember
export { useRemoveMember as useAdminMemberRemove }
