import { tenantConfig } from "$/configs/tenant"
import { apiRouters, queryKeys } from "$/constants"
import { axios } from "$/services/axios"
import type { ErrorResponse } from "$/types/generals"
import type {
  CapabilityProviderType,
  CapabilityResponse,
  CapabilityStatus,
} from "@openstate/types"
import { useQuery } from "@tanstack/react-query"
import type { AxiosError } from "axios"

type UseListCapabilitiesArgs = {
  providerType?: CapabilityProviderType
  status?: CapabilityStatus
  enabled?: boolean
}

type ListCapabilitiesQueryKey = [
  string,
  { providerType?: CapabilityProviderType; status?: CapabilityStatus },
]

const listCapabilities = async ({
  providerType,
  status,
}: {
  providerType?: CapabilityProviderType
  status?: CapabilityStatus
}) => {
  const result = await axios<CapabilityResponse[]>({
    method: "GET",
    url: apiRouters.capabilities.index,
    headers: { "X-Tenant-ID": tenantConfig.tenantId },
    params: {
      providerType: providerType || undefined,
      status: status || undefined,
    },
  })

  return result
}

const useListCapabilities = (args?: UseListCapabilitiesArgs) => {
  const { providerType, status, enabled = true } = args || {}

  const query = useQuery<
    CapabilityResponse[],
    ErrorResponse<AxiosError>,
    CapabilityResponse[],
    ListCapabilitiesQueryKey
  >({
    queryKey: [queryKeys.capabilities.list, { providerType, status }],
    queryFn: () => listCapabilities({ providerType, status }),
    enabled,
  })

  return {
    ...query,
  }
}

export default useListCapabilities
export { useListCapabilities as useCapabilitiesList }
