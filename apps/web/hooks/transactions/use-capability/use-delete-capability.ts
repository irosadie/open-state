import { tenantConfig } from "$/configs/tenant"
import { apiRouters, queryKeys } from "$/constants"
import { axios } from "$/services/axios"
import type { ErrorResponse } from "$/types/generals"
import { pathVariable } from "@openstate/utils"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import type { AxiosError } from "axios"

type DeleteCapabilityResponse = {
  message: string
}

const deleteCapability = async (id: string) => {
  const result = await axios<DeleteCapabilityResponse>({
    method: "DELETE",
    url: pathVariable(apiRouters.capabilities.delete, { id }),
    headers: { "X-Tenant-ID": tenantConfig.tenantId },
  })

  return result
}

const useDeleteCapability = () => {
  const queryClient = useQueryClient()

  const mutation = useMutation<
    DeleteCapabilityResponse,
    ErrorResponse<AxiosError>,
    string
  >({
    mutationKey: [queryKeys.capabilities.delete],
    mutationFn: deleteCapability,
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

export default useDeleteCapability
export { useDeleteCapability as useCapabilitiesDelete }
