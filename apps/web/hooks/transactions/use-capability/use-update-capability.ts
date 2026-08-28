import { tenantConfig } from "$/configs/tenant"
import { apiRouters, queryKeys } from "$/constants"
import { axios } from "$/services/axios"
import type { ErrorResponse } from "$/types/generals"
import type { UpdateCapabilitySchemaProps } from "@openstate/schemas"
import type { CapabilityResponse } from "@openstate/types"
import { pathVariable } from "@openstate/utils"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import type { AxiosError } from "axios"

type UpdateCapabilityPayload = {
  id: string
  payload: UpdateCapabilitySchemaProps
}

const updateCapability = async ({ id, payload }: UpdateCapabilityPayload) => {
  const result = await axios<CapabilityResponse>({
    method: "PATCH",
    url: pathVariable(apiRouters.capabilities.update, { id }),
    headers: { "X-Tenant-ID": tenantConfig.tenantId },
    data: payload,
  })

  return result
}

const useUpdateCapability = () => {
  const queryClient = useQueryClient()

  const mutation = useMutation<
    CapabilityResponse,
    ErrorResponse<AxiosError>,
    UpdateCapabilityPayload
  >({
    mutationKey: [queryKeys.capabilities.update],
    mutationFn: updateCapability,
    onSuccess: (_data, { id }) => {
      void queryClient.invalidateQueries({
        queryKey: [queryKeys.capabilities.list],
      })
      void queryClient.invalidateQueries({
        queryKey: [queryKeys.capabilities.get, id],
      })
    },
  })

  return {
    ...mutation,
  }
}

export default useUpdateCapability
export { useUpdateCapability as useCapabilitiesUpdate }
