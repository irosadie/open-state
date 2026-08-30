import { tenantConfig } from "$/configs/tenant"
import { apiRouters, queryKeys } from "$/constants"
import { axios } from "$/services/axios"
import type { ErrorResponse } from "$/types/generals"
import { runtimeTraceResponseSchema } from "@openstate/schemas"
import type { RuntimeTraceResponse } from "@openstate/types"
import { pathVariable } from "@openstate/utils"
import { useQuery } from "@tanstack/react-query"
import type { AxiosError } from "axios"

type UseGetRuntimeDebugTraceArgs = {
  id: string
  turnId?: string
  enabled?: boolean
}

type RuntimeDebugTraceError = ErrorResponse<AxiosError> & { status?: number }

const getRuntimeDebugTrace = async (id: string, turnId?: string) => {
  const result = await axios<RuntimeTraceResponse>({
    method: "GET",
    url: pathVariable(apiRouters.runtimeInstances.debugTrace, { id }),
    headers: { "X-Tenant-ID": tenantConfig.tenantId },
    params: { turnId: turnId || undefined },
  })
  return runtimeTraceResponseSchema.parse(result)
}

const getErrorStatus = (error: unknown) => {
  if (!error || typeof error !== "object") return undefined
  const value = error as RuntimeDebugTraceError & {
    response?: { status?: number }
  }
  return value.status ?? value.response?.status
}

const useGetRuntimeDebugTrace = ({
  id,
  turnId,
  enabled = true,
}: UseGetRuntimeDebugTraceArgs) => {
  const query = useQuery<
    RuntimeTraceResponse,
    RuntimeDebugTraceError,
    RuntimeTraceResponse,
    [string, string, string | undefined]
  >({
    queryKey: [queryKeys.runtimeInstances.debugTrace, id, turnId],
    queryFn: () => getRuntimeDebugTrace(id, turnId),
    enabled: enabled && !!id,
  })

  return { ...query, isForbidden: getErrorStatus(query.error) === 403 }
}

export default useGetRuntimeDebugTrace
export { useGetRuntimeDebugTrace as useRuntimeDebugTrace }
