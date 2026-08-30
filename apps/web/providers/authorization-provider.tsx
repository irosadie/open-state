"use client"

import { useAuthCurrentUser } from "$/hooks/transactions/use-auth"
import { getApiErrorStatus } from "$/utils/auth-error"
import { hasPermission } from "$/utils/rbac"
import type { Permission } from "$/utils/rbac"
import { authorizationSnapshotSchema } from "@openstate/schemas"
import { useSession } from "next-auth/react"
import {
  type ReactNode,
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
} from "react"

export type AuthorizationStatus =
  | "loading"
  | "ready"
  | "unauthenticated"
  | "forbidden"
  | "error"

type AuthorizationContextValue = {
  status: AuthorizationStatus
  role: string | null
  permissions: readonly Permission[]
  hasPermission: (permission: Permission) => boolean
  refresh: () => Promise<void>
}

const AuthorizationContext = createContext<AuthorizationContextValue | null>(
  null,
)

export function AuthorizationProvider({ children }: { children: ReactNode }) {
  const { status: sessionStatus } = useSession()
  const currentUserQuery = useAuthCurrentUser({
    enabled: sessionStatus === "authenticated",
  })
  const queryErrorStatus = getApiErrorStatus(currentUserQuery.error)

  const status: AuthorizationStatus =
    sessionStatus === "loading" ||
    (sessionStatus === "authenticated" && currentUserQuery.isLoading)
      ? "loading"
      : sessionStatus === "unauthenticated"
        ? "unauthenticated"
        : queryErrorStatus === 403
          ? "forbidden"
          : currentUserQuery.isError
            ? "error"
            : "ready"

  const snapshot = useMemo(() => {
    const parsed = authorizationSnapshotSchema.safeParse(currentUserQuery.data)

    return parsed.success ? parsed.data : { role: null, permissions: [] }
  }, [currentUserQuery.data])
  const permissions = snapshot.permissions

  const checkPermission = useCallback(
    (permission: Permission) => hasPermission(permission, permissions),
    [permissions],
  )

  const refresh = useCallback(async () => {
    await currentUserQuery.refetch()
  }, [currentUserQuery.refetch])

  useEffect(() => {
    const handleForbidden = () => {
      void currentUserQuery.refetch()
    }

    window.addEventListener("openstate:forbidden", handleForbidden)

    return () =>
      window.removeEventListener("openstate:forbidden", handleForbidden)
  }, [currentUserQuery.refetch])

  const value = useMemo<AuthorizationContextValue>(
    () => ({
      status,
      role: snapshot.role ?? null,
      permissions,
      hasPermission: checkPermission,
      refresh,
    }),
    [checkPermission, permissions, refresh, snapshot.role, status],
  )

  return (
    <AuthorizationContext.Provider value={value}>
      {children}
    </AuthorizationContext.Provider>
  )
}

export function useAuthorization() {
  const context = useContext(AuthorizationContext)

  if (!context) {
    throw new Error(
      "useAuthorization must be used within AuthorizationProvider",
    )
  }

  return context
}
