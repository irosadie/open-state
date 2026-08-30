import { tenantConfig } from "$/configs/tenant"
import { apiRouters, queryKeys } from "$/constants"
import { axios } from "$/services/axios"
import type { ErrorResponse } from "$/types/generals"
import { tenantMembershipPageSchema } from "@openstate/schemas"
import type { TenantMembershipPageResponse } from "@openstate/types"
import { useQuery } from "@tanstack/react-query"
import type { AxiosError } from "axios"

type UseListMembersArgs = {
  search?: string
  page?: number
  pageSize?: number
  enabled?: boolean
}

const listMembers = async (args: UseListMembersArgs) => {
  const result = await axios<TenantMembershipPageResponse>({
    method: "GET",
    url: apiRouters.admin.members,
    headers: { "X-Tenant-ID": tenantConfig.tenantId },
    params: {
      search: args.search?.trim() || undefined,
      page: args.page,
      pageSize: args.pageSize,
    },
  })
  return tenantMembershipPageSchema.parse(result)
}

const useListMembers = (args?: UseListMembersArgs) => {
  const { search, page = 1, pageSize = 20, enabled = true } = args || {}
  return useQuery<TenantMembershipPageResponse, ErrorResponse<AxiosError>>({
    queryKey: [queryKeys.admin.members, { search, page, pageSize }],
    queryFn: () => listMembers({ search, page, pageSize }),
    enabled,
  })
}

export default useListMembers
export { useListMembers as useAdminMembers }
