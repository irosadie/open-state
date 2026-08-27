import { apiRouters, queryKeys } from "$/constants"
import { axios } from "$/services/axios"
import type { ErrorResponse } from "$/types/generals"
import { useMutation } from "@tanstack/react-query"
import {
  type RegisterPayloadProps,
  registerPayloadSchema,
} from "@vibecoding-starter/schemas"
import type { AuthRegisterResponse } from "@vibecoding-starter/types"
import type { AxiosError } from "axios"

const register = async (payload: RegisterPayloadProps) => {
  const validatedPayload = registerPayloadSchema.parse(payload)

  const result = await axios<AuthRegisterResponse>({
    method: "POST",
    url: apiRouters.auth.register,
    data: validatedPayload,
  })

  return result
}

const useRegister = () => {
  const mutation = useMutation<
    AuthRegisterResponse,
    ErrorResponse<AxiosError>,
    RegisterPayloadProps,
    unknown
  >({
    mutationKey: [queryKeys.auth.register],
    mutationFn: register,
  })

  return {
    ...mutation,
  }
}

export default useRegister
export { useRegister as useAuthRegister }
