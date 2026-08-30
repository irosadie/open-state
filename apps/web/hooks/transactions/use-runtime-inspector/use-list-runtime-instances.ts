import { tenantConfig } from "$/configs/tenant"
import { apiRouters, queryKeys } from "$/constants"
import { axios } from "$/services/axios"
import type { ErrorResponse } from "$/types/generals"
import { runtimeInstanceListResponseSchema } from "@openstate/schemas"
import type { RuntimeInstanceListResponse } from "@openstate/types"
import { useQuery } from "@tanstack/react-query"
import type { AxiosError } from "axios"

export type UseListRuntimeInstancesArgs = {
  status?: string
  workflowId?: string
  correlationKey?: string
  page?: number
  pageSize?: number
  enabled?: boolean
}

type RuntimeInstanceListQueryKey = [
  string,
  {
    status?: string
    workflowId?: string
    correlationKey?: string
    page: number
    pageSize: number
  },
]

const listRuntimeInstances = async (args: UseListRuntimeInstancesArgs) => {
  const result = await axios<RuntimeInstanceListResponse>({
    method: "GET",
    url: apiRouters.runtimeInstances.index,
    headers: { "X-Tenant-ID": tenantConfig.tenantId },
    params: {
      status: args.status || undefined,
      workflowId: args.workflowId || undefined,
      correlationKey: args.correlationKey || undefined,
      page: args.page || undefined,
      pageSize: args.pageSize || undefined,
    },
  })
  return runtimeInstanceListResponseSchema.parse(result)
}

const useListRuntimeInstances = (args?: UseListRuntimeInstancesArgs) => {
  const {
    status,
    workflowId,
    correlationKey,
    page = 1,
    pageSize = 20,
    enabled = true,
  } = args || {}

  return useQuery<
    RuntimeInstanceListResponse,
    ErrorResponse<AxiosError>,
    RuntimeInstanceListResponse,
    RuntimeInstanceListQueryKey
  >({
    queryKey: [
      queryKeys.runtimeInstances.list,
      { status, workflowId, correlationKey, page, pageSize },
    ],
    queryFn: () =>
      listRuntimeInstances({
        status,
        workflowId,
        correlationKey,
        page,
        pageSize,
      }),
    enabled,
  })
}

export default useListRuntimeInstances
export { useListRuntimeInstances as useRuntimeInstancesList }
