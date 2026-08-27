import { authConfig } from "$/configs/auth"
import { serverAuthConfig } from "$/configs/auth-server"
import { getToken } from "next-auth/jwt"
import { NextResponse } from "next/server"
import type { NextRequest } from "next/server"

const PROTECTED_PATHS = [authConfig.defaultRedirectPath]

export async function proxy(request: NextRequest) {
  const { pathname } = request.nextUrl
  const token = await getToken({
    req: request,
    secret: serverAuthConfig.secret,
    secureCookie: serverAuthConfig.secureCookies,
  })

  const isProtectedRoute = PROTECTED_PATHS.some((path) =>
    pathname.startsWith(path),
  )
  const isLoginPage = pathname === authConfig.loginPath
  const callbackUrl =
    pathname.startsWith("/") && pathname !== authConfig.loginPath
      ? pathname
      : authConfig.defaultRedirectPath

  if (isProtectedRoute && !token) {
    const loginUrl = new URL(authConfig.loginPath, request.url)

    loginUrl.searchParams.set("callbackUrl", callbackUrl)

    return NextResponse.redirect(loginUrl)
  }

  if (isLoginPage && token) {
    return NextResponse.redirect(
      new URL(authConfig.defaultRedirectPath, request.url),
    )
  }

  return NextResponse.next()
}

export const config = {
  matcher: ["/login"],
}
