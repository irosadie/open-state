import { authConfig } from "$/configs/auth"
import { serverAuthConfig } from "$/configs/auth-server"
import { getToken } from "next-auth/jwt"
import { NextResponse } from "next/server"
import type { NextRequest } from "next/server"

const PUBLIC_PATHS = new Set([authConfig.loginPath, "/register"])

const isPublicPath = (pathname: string) => PUBLIC_PATHS.has(pathname)

export async function proxy(request: NextRequest) {
  const { pathname } = request.nextUrl
  const token = await getToken({
    req: request,
    secret: serverAuthConfig.secret,
    secureCookie: serverAuthConfig.secureCookies,
  })

  const isProtectedRoute = !isPublicPath(pathname)
  const callbackUrl =
    pathname.startsWith("/") && !isPublicPath(pathname)
      ? pathname
      : authConfig.defaultRedirectPath

  if (isProtectedRoute && !token) {
    const loginUrl = new URL(authConfig.loginPath, request.url)

    loginUrl.searchParams.set("callbackUrl", callbackUrl)

    return NextResponse.redirect(loginUrl)
  }

  return NextResponse.next()
}

export const config = {
  matcher: ["/((?!api|_next/static|_next/image|favicon.ico|.*\\..*).*)"],
}
