import { tenantConfig } from "$/configs/tenant"
import { apiRouters, queryKeys } from "$/constants"
import { axios } from "$/services/axios"
import type { ErrorResponse } from "$/types/generals"
import type { CapabilityResponse } from "@openstate/types"
import { pathVariable } from "@openstate/utils"
import { useQuery } from "@tanstack/react-query"
import type { AxiosError } from "axios"

type UseGetCapabilityArgs = {
  id: string
  enabled?: boolean
}

const getCapability = async (id: string) => {
  const result = await axios<CapabilityResponse>({
    method: "GET",
    url: pathVariable(apiRouters.capabilities.show, { id }),
    headers: { "X-Tenant-ID": tenantConfig.tenantId },
  })

  return result
}

const useGetCapability = (args: UseGetCapabilityArgs) => {
  const { id, enabled = true } = args

  const query = useQuery<
    CapabilityResponse,
    ErrorResponse<AxiosError>,
    CapabilityResponse,
    [string, string]
  >({
    queryKey: [queryKeys.capabilities.get, id],
    queryFn: () => getCapability(id),
    enabled: enabled && !!id,
  })

  return {
    ...query,
  }
}

export default useGetCapability
export { useGetCapability as useCapabilitiesGet }
