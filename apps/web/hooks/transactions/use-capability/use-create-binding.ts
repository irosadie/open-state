import { tenantConfig } from "$/configs/tenant"
import { apiRouters, queryKeys } from "$/constants"
import { axios } from "$/services/axios"
import type { ErrorResponse } from "$/types/generals"
import type { BindingSchemaProps } from "@openstate/schemas"
import type { CapabilityBindingResponse } from "@openstate/types"
import { pathVariable } from "@openstate/utils"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import type { AxiosError } from "axios"

type CreateBindingPayload = {
  capabilityId: string
  payload: BindingSchemaProps
}

const createBinding = async ({
  capabilityId,
  payload,
}: CreateBindingPayload) => {
  const result = await axios<CapabilityBindingResponse>({
    method: "POST",
    url: pathVariable(apiRouters.capabilities.bindings, { id: capabilityId }),
    headers: { "X-Tenant-ID": tenantConfig.tenantId },
    data: payload,
  })

  return result
}

const useCreateBinding = () => {
  const queryClient = useQueryClient()

  const mutation = useMutation<
    CapabilityBindingResponse,
    ErrorResponse<AxiosError>,
    CreateBindingPayload
  >({
    mutationKey: [queryKeys.capabilities.create],
    mutationFn: createBinding,
    onSuccess: (_data, { capabilityId }) => {
      void queryClient.invalidateQueries({
        queryKey: [queryKeys.capabilities.bindings, capabilityId],
      })
    },
  })

  return {
    ...mutation,
  }
}

export default useCreateBinding
export { useCreateBinding as useCapabilitiesCreateBinding }
