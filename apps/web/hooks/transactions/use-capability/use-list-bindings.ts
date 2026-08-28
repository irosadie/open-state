import { tenantConfig } from "$/configs/tenant"
import { apiRouters, queryKeys } from "$/constants"
import { axios } from "$/services/axios"
import type { ErrorResponse } from "$/types/generals"
import type { CapabilityBindingResponse } from "@openstate/types"
import { pathVariable } from "@openstate/utils"
import { useQuery } from "@tanstack/react-query"
import type { AxiosError } from "axios"

type UseListBindingsArgs = {
  capabilityId: string
  enabled?: boolean
}

const listBindings = async (capabilityId: string) => {
  const result = await axios<CapabilityBindingResponse[]>({
    method: "GET",
    url: pathVariable(apiRouters.capabilities.bindings, { id: capabilityId }),
    headers: { "X-Tenant-ID": tenantConfig.tenantId },
  })

  return result
}

const useListBindings = (args: UseListBindingsArgs) => {
  const { capabilityId, enabled = true } = args

  const query = useQuery<
    CapabilityBindingResponse[],
    ErrorResponse<AxiosError>,
    CapabilityBindingResponse[],
    [string, string]
  >({
    queryKey: [queryKeys.capabilities.bindings, capabilityId],
    queryFn: () => listBindings(capabilityId),
    enabled: enabled && !!capabilityId,
  })

  return {
    ...query,
  }
}

export default useListBindings
export { useListBindings as useCapabilitiesListBindings }
