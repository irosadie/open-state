import { tenantConfig } from "$/configs/tenant"
import { apiRouters, queryKeys } from "$/constants"
import { axios } from "$/services/axios"
import type { ErrorResponse } from "$/types/generals"
import { intentListResponseSchema } from "@openstate/schemas"
import type { IntentListResponse } from "@openstate/types"
import { useQuery } from "@tanstack/react-query"
import type { AxiosError } from "axios"

type UseListIntentsArgs = {
  projectId?: string
  enabled?: boolean
}

type ListIntentsQueryKey = [string, { projectId?: string }]

const listIntents = async (projectId?: string) => {
  const result = await axios<IntentListResponse>({
    method: "GET",
    url: apiRouters.intents.index,
    headers: { "X-Tenant-ID": tenantConfig.tenantId },
    params: { projectId: projectId || undefined },
  })

  return intentListResponseSchema.parse(result)
}

const useListIntents = (args?: UseListIntentsArgs) => {
  const { projectId, enabled = true } = args || {}

  return useQuery<
    IntentListResponse,
    ErrorResponse<AxiosError>,
    IntentListResponse,
    ListIntentsQueryKey
  >({
    queryKey: [queryKeys.intents.list, { projectId }],
    queryFn: () => listIntents(projectId),
    enabled,
  })
}

export default useListIntents
export { useListIntents as useIntentsList }
