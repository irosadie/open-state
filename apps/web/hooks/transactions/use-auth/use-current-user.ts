import { apiRouters, queryKeys } from "$/constants"
import { axios } from "$/services/axios"
import type { ErrorResponse } from "$/types/generals"
import { useQuery } from "@tanstack/react-query"
import type { AuthCurrentUserResponse } from "@vibecoding-starter/types"
import type { AxiosError } from "axios"

type UseCurrentUserArgs = {
  enabled?: boolean
}

const getCurrentUser = async () => {
  const result = await axios<AuthCurrentUserResponse>({
    method: "GET",
    url: apiRouters.auth.me,
  })

  return result
}

const useCurrentUser = (args?: UseCurrentUserArgs) => {
  const { enabled = true } = args || {}

  const query = useQuery<
    AuthCurrentUserResponse,
    ErrorResponse<AxiosError>,
    AuthCurrentUserResponse,
    [string]
  >({
    queryKey: [queryKeys.auth.me],
    queryFn: getCurrentUser,
    enabled,
  })

  return {
    ...query,
  }
}

export default useCurrentUser
export { useCurrentUser as useAuthCurrentUser }
