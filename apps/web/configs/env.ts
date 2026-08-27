import { authConfig } from "./auth"

export const env = {
  appBaseUrl:
    process.env.NEXT_PUBLIC_APP_URL ||
    process.env.NEXTAUTH_URL ||
    "http://localhost:3020",
  apiBaseUrl: authConfig.proxyApiBasePath,
  authBaseUrl: authConfig.authApiBasePath,
  defaultAuthRedirectPath: authConfig.defaultRedirectPath,
}
