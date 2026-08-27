import { authConfig } from "$/configs/auth"
import { env } from "$/configs/env"
import type { ErrorResponse } from "$/types/generals"
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
    // Session expired — redirect to login
    if (error.response?.status === 401 && typeof window !== "undefined") {
      window.location.href = authConfig.loginPath

      return new Promise(() => {})
    }

    throw (error.response?.data ?? error) as ErrorResponse
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
