import { authConfig } from "$/configs/auth"
import { serverAuthConfig } from "$/configs/auth-server"
import type { AuthTokensResponse } from "@vibecoding-starter/types"
import axios, { type AxiosError } from "axios"
import { encode, getToken } from "next-auth/jwt"
import { NextResponse } from "next/server"
import type { NextRequest } from "next/server"

type ProxyPayload<T> = T | { data: T }

type SessionToken = {
  accessToken?: string
  refreshToken?: string
  accessTokenExpires?: number
  [key: string]: unknown
}

const publicProxyPaths = new Set([
  authConfig.backendLoginPath.replace(/^\//, ""),
  authConfig.backendRefreshPath.replace(/^\//, ""),
])

const methodsWithBody = new Set(["POST", "PUT", "PATCH", "DELETE"])

const unwrapData = <T>(payload: ProxyPayload<T>): T => {
  if (
    typeof payload === "object" &&
    payload !== null &&
    "data" in payload &&
    payload.data
  ) {
    return payload.data as T
  }

  return payload as T
}

const isPublicProxyPath = (path: string) => publicProxyPaths.has(path)

const getTargetUrl = (path: string, search: string) => {
  // Backend routes diawali /api/... (misal /api/auth/register).
  // Path yang masuk ke proxy berbentuk "auth/register" tanpa prefix /api,
  // sehingga kita tambahkan prefix /api di sini.
  const cleanPath = path.replace(/^\/+/, "")
  const targetUrl = new URL(
    `/api/${cleanPath}`,
    serverAuthConfig.backendApiBaseUrl,
  )

  targetUrl.search = search

  return targetUrl.toString()
}

const setSessionCookie = async (
  response: NextResponse,
  token: SessionToken,
  tokens: AuthTokensResponse,
) => {
  const encodedToken = await encode({
    token: {
      ...token,
      accessToken: tokens.accessToken,
      refreshToken: tokens.refreshToken,
      accessTokenExpires: Date.now() + tokens.expiresIn * 1000,
    },
    secret: serverAuthConfig.secret,
    maxAge: authConfig.sessionMaxAgeSeconds,
  })

  response.cookies.set(serverAuthConfig.sessionCookieName, encodedToken, {
    httpOnly: true,
    sameSite: "lax",
    secure: serverAuthConfig.secureCookies,
    path: "/",
    maxAge: authConfig.sessionMaxAgeSeconds,
  })
}

const clearSessionCookie = (response: NextResponse) => {
  response.cookies.set(serverAuthConfig.sessionCookieName, "", {
    httpOnly: true,
    sameSite: "lax",
    secure: serverAuthConfig.secureCookies,
    path: "/",
    maxAge: 0,
  })
}

const refreshAccessToken = async (refreshToken: string) => {
  try {
    const { data } = await axios.post<ProxyPayload<AuthTokensResponse>>(
      getTargetUrl(authConfig.backendRefreshPath.replace(/^\//, ""), ""),
      undefined,
      {
        headers: {
          Authorization: `Bearer ${refreshToken}`,
        },
      },
    )

    return unwrapData(data)
  } catch {
    return null
  }
}

const readRequestBody = async (request: NextRequest) => {
  if (!methodsWithBody.has(request.method)) {
    return undefined
  }

  const contentType = request.headers.get("content-type") || ""

  if (contentType.includes("multipart/form-data")) {
    return request.formData()
  }

  if (contentType.includes("application/json")) {
    const textBody = await request.text()

    if (!textBody) {
      return undefined
    }

    return JSON.parse(textBody) as unknown
  }

  const textBody = await request.text()

  return textBody || undefined
}

const buildRequestHeaders = (
  request: NextRequest,
  accessToken?: string,
): Record<string, string> => {
  const headers: Record<string, string> = {}
  const contentType = request.headers.get("content-type")

  if (contentType && !contentType.includes("multipart/form-data")) {
    headers["Content-Type"] = contentType
  }

  if (accessToken) {
    headers.Authorization = `Bearer ${accessToken}`
  }

  return headers
}

const buildErrorResponse = (error: unknown, shouldClearSession = false) => {
  const axiosError = error as AxiosError
  const status = axiosError.response?.status || 500
  const data = axiosError.response?.data || {
    message: "Proxy request failed",
  }
  const response = NextResponse.json(data, { status })

  if (shouldClearSession) {
    clearSessionCookie(response)
  }

  return response
}

const proxyRequest = async (request: NextRequest, pathSegments: string[]) => {
  const normalizedPath = pathSegments.join("/")
  const token = (await getToken({
    req: request,
    secret: serverAuthConfig.secret,
    secureCookie: serverAuthConfig.secureCookies,
  })) as SessionToken | null
  const isPublicPath = isPublicProxyPath(normalizedPath)
  const initialAccessToken = isPublicPath ? undefined : token?.accessToken
  const body = await readRequestBody(request)
  const targetUrl = getTargetUrl(normalizedPath, request.nextUrl.search)

  let nextAccessToken = initialAccessToken
  let refreshedTokens: AuthTokensResponse | null = null

  if (
    !isPublicPath &&
    token?.refreshToken &&
    token.accessTokenExpires &&
    Date.now() >= token.accessTokenExpires - 60_000
  ) {
    refreshedTokens = await refreshAccessToken(token.refreshToken)

    if (refreshedTokens) {
      nextAccessToken = refreshedTokens.accessToken
    }
  }

  try {
    const { data, status } = await axios({
      method: request.method,
      url: targetUrl,
      data: body,
      headers: buildRequestHeaders(request, nextAccessToken),
    })
    const response = NextResponse.json(data, { status })

    if (token && refreshedTokens) {
      await setSessionCookie(response, token, refreshedTokens)
    }

    return response
  } catch (error) {
    const axiosError = error as AxiosError

    if (
      axiosError.response?.status === 401 &&
      !isPublicPath &&
      token?.refreshToken
    ) {
      const fallbackTokens =
        refreshedTokens || (await refreshAccessToken(token.refreshToken))

      if (!fallbackTokens) {
        return buildErrorResponse(error, true)
      }

      try {
        const retryResponse = await axios({
          method: request.method,
          url: targetUrl,
          data: body,
          headers: buildRequestHeaders(request, fallbackTokens.accessToken),
        })
        const response = NextResponse.json(retryResponse.data, {
          status: retryResponse.status,
        })

        await setSessionCookie(response, token, fallbackTokens)

        return response
      } catch (retryError) {
        const retryAxiosError = retryError as AxiosError
        const shouldClearSession = retryAxiosError.response?.status === 401

        return buildErrorResponse(retryError, shouldClearSession)
      }
    }

    return buildErrorResponse(error)
  }
}

export const runtime = "nodejs"

type RouteContext = {
  params: Promise<{ path: string[] }>
}

export async function GET(request: NextRequest, context: RouteContext) {
  const { path } = await context.params

  return proxyRequest(request, path)
}

export async function POST(request: NextRequest, context: RouteContext) {
  const { path } = await context.params

  return proxyRequest(request, path)
}

export async function PUT(request: NextRequest, context: RouteContext) {
  const { path } = await context.params

  return proxyRequest(request, path)
}

export async function PATCH(request: NextRequest, context: RouteContext) {
  const { path } = await context.params

  return proxyRequest(request, path)
}

export async function DELETE(request: NextRequest, context: RouteContext) {
  const { path } = await context.params

  return proxyRequest(request, path)
}
