import { tenantConfig } from "$/configs/tenant"
import { apiRouters, queryKeys } from "$/constants"
import { axios } from "$/services/axios"
import type { ErrorResponse } from "$/types/generals"
import type { AuthCurrentUserResponse } from "@openstate/types"
import { useQuery } from "@tanstack/react-query"
import type { AxiosError } from "axios"

type UseCurrentUserArgs = {
  enabled?: boolean
}

const getCurrentUser = async () => {
  const result = await axios<AuthCurrentUserResponse>({
    method: "GET",
    url: apiRouters.auth.me,
    headers: { "X-Tenant-ID": tenantConfig.tenantId },
  })

  return result
}

const useCurrentUser = (args?: UseCurrentUserArgs) => {
  const { enabled = true } = args || {}

  const query = useQuery<
    AuthCurrentUserResponse,
    ErrorResponse<AxiosError>,
    AuthCurrentUserResponse,
    [string, string]
  >({
    queryKey: [queryKeys.auth.me, tenantConfig.tenantId],
    queryFn: getCurrentUser,
    enabled,
  })

  return {
    ...query,
  }
}

export default useCurrentUser
export { useCurrentUser as useAuthCurrentUser }
