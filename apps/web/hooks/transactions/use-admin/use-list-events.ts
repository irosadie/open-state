import { tenantConfig } from "$/configs/tenant"
import { apiRouters, queryKeys } from "$/constants"
import { axios } from "$/services/axios"
import type { ErrorResponse } from "$/types/generals"
import { eventPageSchema } from "@openstate/schemas"
import type { EventPageResponse } from "@openstate/types"
import { useQuery } from "@tanstack/react-query"
import type { AxiosError } from "axios"

type ListEventsArgs = {
  workflowInstanceId?: string
  type?: string
  source?: string
  correlationId?: string
  page?: number
  pageSize?: number
  enabled?: boolean
}

const listEvents = async (args: ListEventsArgs) => {
  const result = await axios<EventPageResponse>({
    method: "GET",
    url: apiRouters.admin.events,
    headers: { "X-Tenant-ID": tenantConfig.tenantId },
    params: {
      workflowInstanceId: args.workflowInstanceId || undefined,
      type: args.type || undefined,
      source: args.source || undefined,
      correlationId: args.correlationId || undefined,
      page: args.page,
      pageSize: args.pageSize,
    },
  })
  return eventPageSchema.parse(result)
}

const useListEvents = (args?: ListEventsArgs) => {
  const {
    workflowInstanceId,
    type,
    source,
    correlationId,
    page = 1,
    pageSize = 20,
    enabled = true,
  } = args || {}
  return useQuery<EventPageResponse, ErrorResponse<AxiosError>>({
    queryKey: [
      queryKeys.admin.events,
      { workflowInstanceId, type, source, correlationId, page, pageSize },
    ],
    queryFn: () =>
      listEvents({
        workflowInstanceId,
        type,
        source,
        correlationId,
        page,
        pageSize,
      }),
    enabled,
  })
}

export default useListEvents
export { useListEvents as useAdminEvents }
