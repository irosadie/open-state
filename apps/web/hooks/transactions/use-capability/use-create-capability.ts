import { tenantConfig } from "$/configs/tenant"
import { apiRouters, queryKeys } from "$/constants"
import { axios } from "$/services/axios"
import type { ErrorResponse } from "$/types/generals"
import type { CreateCapabilitySchemaProps } from "@openstate/schemas"
import type { CapabilityResponse } from "@openstate/types"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import type { AxiosError } from "axios"

const createCapability = async (payload: CreateCapabilitySchemaProps) => {
  const result = await axios<CapabilityResponse>({
    method: "POST",
    url: apiRouters.capabilities.index,
    headers: { "X-Tenant-ID": tenantConfig.tenantId },
    data: payload,
  })

  return result
}

const useCreateCapability = () => {
  const queryClient = useQueryClient()

  const mutation = useMutation<
    CapabilityResponse,
    ErrorResponse<AxiosError>,
    CreateCapabilitySchemaProps
  >({
    mutationKey: [queryKeys.capabilities.create],
    mutationFn: createCapability,
    onSuccess: () => {
      void queryClient.invalidateQueries({
        queryKey: [queryKeys.capabilities.list],
      })
    },
  })

  return {
    ...mutation,
  }
}

export default useCreateCapability
export { useCreateCapability as useCapabilitiesCreate }
