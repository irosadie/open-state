import { tenantConfig } from "$/configs/tenant"
import { apiRouters, queryKeys } from "$/constants"
import { axios } from "$/services/axios"
import type { ErrorResponse } from "$/types/generals"
import { instanceResponseSchema } from "@openstate/schemas"
import type { InstanceResponse } from "@openstate/types"
import { useQuery } from "@tanstack/react-query"
import type { AxiosError } from "axios"

const listInstances = async () => {
  const result = await axios<InstanceResponse[]>({
    method: "GET",
    url: apiRouters.admin.instances,
    headers: { "X-Tenant-ID": tenantConfig.tenantId },
  })
  return instanceResponseSchema.array().parse(result)
}

const useListInstances = (enabled = true) => {
  return useQuery<InstanceResponse[], ErrorResponse<AxiosError>>({
    queryKey: [queryKeys.admin.instances],
    queryFn: listInstances,
    enabled,
  })
}

export default useListInstances
export { useListInstances as useAdminInstances }
