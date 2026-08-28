import { tenantConfig } from "$/configs/tenant"
import { apiRouters, queryKeys } from "$/constants"
import { axios } from "$/services/axios"
import type { ErrorResponse } from "$/types/generals"
import { pathVariable } from "@openstate/utils"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import type { AxiosError } from "axios"

type DeleteBindingResponse = {
  message: string
}

type DeleteBindingPayload = {
  bindingId: string
  capabilityId: string
}

const deleteBinding = async ({ bindingId }: DeleteBindingPayload) => {
  const result = await axios<DeleteBindingResponse>({
    method: "DELETE",
    url: pathVariable(apiRouters.bindings.delete, { id: bindingId }),
    headers: { "X-Tenant-ID": tenantConfig.tenantId },
  })

  return result
}

const useDeleteBinding = () => {
  const queryClient = useQueryClient()

  const mutation = useMutation<
    DeleteBindingResponse,
    ErrorResponse<AxiosError>,
    DeleteBindingPayload
  >({
    mutationKey: [queryKeys.bindings.delete],
    mutationFn: deleteBinding,
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

export default useDeleteBinding
export { useDeleteBinding as useCapabilitiesDeleteBinding }
