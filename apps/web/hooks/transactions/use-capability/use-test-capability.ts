import { tenantConfig } from "$/configs/tenant"
import { apiRouters, queryKeys } from "$/constants"
import { axios } from "$/services/axios"
import type { ErrorResponse } from "$/types/generals"
import type { TestInvocationSchemaProps } from "@openstate/schemas"
import type {
  CapabilityErrorResponse,
  CapabilityInvocationResultResponse,
} from "@openstate/types"
import { pathVariable } from "@openstate/utils"
import { useMutation } from "@tanstack/react-query"
import type { AxiosError } from "axios"

type TestCapabilityPayload = {
  capabilityId: string
  payload: TestInvocationSchemaProps
}

const testCapability = async ({
  capabilityId,
  payload,
}: TestCapabilityPayload) => {
  const result = await axios<CapabilityInvocationResultResponse>({
    method: "POST",
    url: pathVariable(apiRouters.capabilities.test, { id: capabilityId }),
    headers: { "X-Tenant-ID": tenantConfig.tenantId },
    data: payload,
  })

  return result
}

const useTestCapability = () => {
  const mutation = useMutation<
    CapabilityInvocationResultResponse,
    ErrorResponse<AxiosError> & { errors?: CapabilityErrorResponse },
    TestCapabilityPayload
  >({
    mutationKey: [queryKeys.capabilities.test],
    mutationFn: testCapability,
  })

  return {
    ...mutation,
  }
}

export default useTestCapability
export { useTestCapability as useCapabilitiesTest }
