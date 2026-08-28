import { tenantConfig } from "$/configs/tenant"
import { apiRouters, queryKeys } from "$/constants"
import { axios } from "$/services/axios"
import type { ErrorResponse } from "$/types/generals"
import type { AuditPageResponse } from "@openstate/types"
import { useQuery } from "@tanstack/react-query"
import type { AxiosError } from "axios"

type UseListAuditArgs = {
  action?: string
  resourceType?: string
  resourceId?: string
  actor?: string
  page?: number
  pageSize?: number
  enabled?: boolean
}

type AuditQueryKey = [
  string,
  {
    action?: string
    resourceType?: string
    resourceId?: string
    actor?: string
    page?: number
    pageSize?: number
  },
]

const listAudit = async (args: UseListAuditArgs) => {
  const result = await axios<AuditPageResponse>({
    method: "GET",
    url: apiRouters.audit.index,
    headers: { "X-Tenant-ID": tenantConfig.tenantId },
    params: {
      action: args.action || undefined,
      resourceType: args.resourceType || undefined,
      resourceId: args.resourceId || undefined,
      actor: args.actor || undefined,
      page: args.page || undefined,
      pageSize: args.pageSize || undefined,
    },
  })

  return result
}

const useListAudit = (args?: UseListAuditArgs) => {
  const {
    action,
    resourceType,
    resourceId,
    actor,
    page = 1,
    pageSize = 20,
    enabled = true,
  } = args || {}

  const query = useQuery<
    AuditPageResponse,
    ErrorResponse<AxiosError>,
    AuditPageResponse,
    AuditQueryKey
  >({
    queryKey: [
      queryKeys.audit.list,
      { action, resourceType, resourceId, actor, page, pageSize },
    ],
    queryFn: () =>
      listAudit({ action, resourceType, resourceId, actor, page, pageSize }),
    enabled,
  })

  return {
    ...query,
  }
}

export default useListAudit
export { useListAudit as useAuditList }
