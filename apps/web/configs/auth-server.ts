const isProduction = process.env.NODE_ENV === "production"

export const serverAuthConfig = {
  appBaseUrl:
    process.env.NEXTAUTH_URL ||
    process.env.NEXT_PUBLIC_APP_URL ||
    "http://localhost:3020",
  backendApiBaseUrl:
    process.env.API_URL ||
    process.env.INTERNAL_API_URL ||
    process.env.NEXT_PUBLIC_API_URL ||
    "http://localhost:8020",
  secret:
    process.env.NEXTAUTH_SECRET ||
    process.env.AUTH_SECRET ||
    process.env.JWT_SECRET ||
    "development-only-change-me",
  secureCookies: isProduction,
  sessionCookieName: isProduction
    ? "__Secure-next-auth.session-token"
    : "next-auth.session-token",
} as const
