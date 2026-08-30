import { tenantConfig } from "$/configs/tenant"
import { apiRouters, queryKeys } from "$/constants"
import { axios } from "$/services/axios"
import type { ErrorResponse } from "$/types/generals"
import { runtimeInstanceDetailResponseSchema } from "@openstate/schemas"
import type { RuntimeInstanceDetailResponse } from "@openstate/types"
import { pathVariable } from "@openstate/utils"
import { useQuery } from "@tanstack/react-query"
import type { AxiosError } from "axios"

type UseGetRuntimeInstanceArgs = {
  id: string
  enabled?: boolean
}

const getRuntimeInstance = async (id: string) => {
  const result = await axios<RuntimeInstanceDetailResponse>({
    method: "GET",
    url: pathVariable(apiRouters.runtimeInstances.show, { id }),
    headers: { "X-Tenant-ID": tenantConfig.tenantId },
  })
  return runtimeInstanceDetailResponseSchema.parse(result)
}

const useGetRuntimeInstance = ({
  id,
  enabled = true,
}: UseGetRuntimeInstanceArgs) =>
  useQuery<
    RuntimeInstanceDetailResponse,
    ErrorResponse<AxiosError>,
    RuntimeInstanceDetailResponse,
    [string, string]
  >({
    queryKey: [queryKeys.runtimeInstances.get, id],
    queryFn: () => getRuntimeInstance(id),
    enabled: enabled && !!id,
  })

export default useGetRuntimeInstance
export { useGetRuntimeInstance as useRuntimeInstanceGet }
