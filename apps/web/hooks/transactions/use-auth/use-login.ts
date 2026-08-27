import { authConfig } from "$/configs/auth"
import { queryKeys } from "$/constants"
import { type LoginProps, loginSchema } from "@openstate/schemas"
import { useMutation } from "@tanstack/react-query"
import { signIn } from "next-auth/react"

type LoginPayload = LoginProps & {
  callbackUrl?: string
}

type LoginResult = {
  redirectTo: string
}

const login = async ({ callbackUrl, ...credentials }: LoginPayload) => {
  const validatedCredentials = loginSchema.parse(credentials)
  const normalizedCallbackUrl = callbackUrl?.startsWith("/")
    ? callbackUrl
    : authConfig.defaultRedirectPath

  const result = await signIn("credentials", {
    ...validatedCredentials,
    redirect: false,
    callbackUrl: normalizedCallbackUrl,
  })

  if (!result) {
    throw new Error("Login failed")
  }

  if (result.error) {
    throw new Error(result.error)
  }

  return {
    redirectTo: result.url || normalizedCallbackUrl,
  }
}

const useLogin = () => {
  const mutation = useMutation<LoginResult, Error, LoginPayload, unknown>({
    mutationKey: [queryKeys.auth.login],
    mutationFn: login,
  })

  return {
    ...mutation,
  }
}

export default useLogin
export { useLogin as useAuthLogin }
export type { LoginPayload, LoginResult }
