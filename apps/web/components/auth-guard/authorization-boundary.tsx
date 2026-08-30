"use client"

import { AccessDenied } from "$/components/auth-guard/access-denied"
import { useAuthorization } from "$/providers/authorization-provider"
import { canAccessRoute, isPublicRoute } from "$/utils/rbac"
import { usePathname, useRouter } from "next/navigation"
import { type ReactNode, useEffect } from "react"

const loginPath = "/login"

function LoadingBoundary() {
  return (
    <main className="flex min-h-screen items-center justify-center px-6 py-12 text-sm text-slate-600">
      Checking access…
    </main>
  )
}

export function AuthorizationBoundary({ children }: { children: ReactNode }) {
  const pathname = usePathname() || "/"
  const router = useRouter()
  const authorization = useAuthorization()
  const isPublic = isPublicRoute(pathname)

  useEffect(() => {
    if (!isPublic && authorization.status === "unauthenticated") {
      const loginUrl = `${loginPath}?callbackUrl=${encodeURIComponent(pathname)}`
      router.replace(loginUrl)
    }
  }, [authorization.status, isPublic, pathname, router])

  if (isPublic) return children
  if (authorization.status === "loading") return <LoadingBoundary />
  if (authorization.status === "unauthenticated") return <LoadingBoundary />
  if (authorization.status === "forbidden") {
    return (
      <AccessDenied
        title="Access denied"
        description="Your current tenant role cannot be verified for this page."
        onRetry={() => void authorization.refresh()}
      />
    )
  }
  if (authorization.status === "error") {
    return (
      <AccessDenied
        title="Unable to verify access"
        description="We could not load your current tenant permissions."
        onRetry={() => void authorization.refresh()}
      />
    )
  }
  if (!canAccessRoute(pathname, authorization.permissions)) {
    return (
      <AccessDenied description="Your role cannot access this application route." />
    )
  }

  return children
}
