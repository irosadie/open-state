export const authConfig = {
  loginPath: "/login",
  defaultRedirectPath: "/",
  authApiBasePath: "/api/auth",
  proxyApiBasePath: process.env.NEXT_PUBLIC_API_PROXY_BASE_URL ?? "/api/proxy",
  backendLoginPath: process.env.AUTH_LOGIN_PATH ?? "/auth/login",
  backendRefreshPath: process.env.AUTH_REFRESH_PATH ?? "/auth/refresh",
  backendLogoutPath: process.env.AUTH_LOGOUT_PATH ?? "/auth/logout",
  sessionMaxAgeSeconds: 6 * 60 * 60,
} as const

export function getRoleRedirectPath(_role: string): string {
  return authConfig.defaultRedirectPath
}
