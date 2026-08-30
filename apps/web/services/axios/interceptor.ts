import { authConfig } from "$/configs/auth"
import { env } from "$/configs/env"
import type { ErrorResponse } from "$/types/generals"
import { getAuthRecoveryAction } from "$/utils/auth-error"
import axiosClient, { type AxiosRequestConfig } from "axios"

/**
 * Creates an initial 'axios' instance with custom settings.
 */

const apiBaseUrl = env.apiBaseUrl

const instance = axiosClient.create({
  baseURL: `${apiBaseUrl}`,
  headers: {
    "Content-Type": "application/json",
  },
})

instance.interceptors.request.use(
  async (config) => {
    /**
     * the bearer token is no needed for currently, but will needed if we use external API
     * const session = await getSession()
     * if (session && session.accessToken) {
     *   config.headers.Authorization = `Bearer ${session.accessToken}`
     * }
     */
    return config
  },
  (error) => {
    throw error
  },
)

/**
 * Handle all responses. It is possible to add handlers
 * for requests, but it is omitted here for brevity.
 */
instance.interceptors.response.use(
  async (res) => {
    const { meta, data } = res.data

    if (meta && data) {
      return {
        list: data,
        meta: {
          pagination: meta,
          cursor: null,
        },
      }
    }

    if (data) {
      return data
    }

    return res.data
  },
  (error) => {
    const recoveryAction = getAuthRecoveryAction(error.response?.status)

    // Session expired — redirect to login, but skip if already there
    if (
      recoveryAction === "login" &&
      typeof window !== "undefined" &&
      window.location.pathname !== authConfig.loginPath
    ) {
      const loginUrl = new URL(authConfig.loginPath, window.location.origin)

      loginUrl.searchParams.set(
        "callbackUrl",
        window.location.pathname + window.location.search,
      )

      window.location.href = loginUrl.toString()

      return new Promise(() => {})
    }

    if (recoveryAction === "forbidden" && typeof window !== "undefined") {
      window.dispatchEvent(new Event("openstate:forbidden"))
    }

    const responseData: unknown = error.response?.data ?? error

    if (responseData && typeof responseData === "object") {
      throw {
        ...responseData,
        status: error.response?.status,
      } as ErrorResponse
    }

    throw responseData as ErrorResponse
  },
)
/**
 * Replaces main `axios` instance with the custom-one.
 *
 * @param cfg - Axios configuration object.
 * @returns A promise object of a response of the HTTP request with the 'data' object already
 * destructured.
 */
const axios = <T>(cfg: AxiosRequestConfig) => instance.request<unknown, T>(cfg)

export default axios
