import { apiRouters, queryKeys } from "$/constants"
import { axios } from "$/services/axios"
import { useMutation } from "@tanstack/react-query"
import { signOut } from "next-auth/react"

type LogoutResponse = {
  success: boolean
}

const logout = async () => {
  try {
    await axios<LogoutResponse>({
      method: "POST",
      url: apiRouters.auth.logout,
    })
  } catch {
    // Keep local logout working even while backend logout endpoint is unavailable.
  }

  await signOut({
    redirect: false,
  })
}

const useLogout = () => {
  const mutation = useMutation<void, Error, void, unknown>({
    mutationKey: [queryKeys.auth.logout],
    mutationFn: logout,
  })

  return {
    ...mutation,
  }
}

export default useLogout
export { useLogout as useAuthLogout }
