import { tenantConfig } from "$/configs/tenant"
import { apiRouters, queryKeys } from "$/constants"
import { axios } from "$/services/axios"
import type { ErrorResponse } from "$/types/generals"
import { eventResponseSchema } from "@openstate/schemas"
import type { EventResponse } from "@openstate/types"
import { pathVariable } from "@openstate/utils"
import { useQuery } from "@tanstack/react-query"
import type { AxiosError } from "axios"

const getEvent = async (eventId: string) => {
  const result = await axios<EventResponse>({
    method: "GET",
    url: pathVariable(apiRouters.admin.event, { eventId }),
    headers: { "X-Tenant-ID": tenantConfig.tenantId },
  })
  return eventResponseSchema.parse(result)
}

const useGetEvent = (eventId: string, enabled = true) => {
  return useQuery<EventResponse, ErrorResponse<AxiosError>>({
    queryKey: [queryKeys.admin.event, eventId],
    queryFn: () => getEvent(eventId),
    enabled: enabled && !!eventId,
  })
}

export default useGetEvent
export { useGetEvent as useAdminEvent }
