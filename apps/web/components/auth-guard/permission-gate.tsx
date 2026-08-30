"use client"

import { useAuthorization } from "$/providers/authorization-provider"
import { canAccessAction } from "$/utils/rbac"
import type { Permission } from "$/utils/rbac"
import type { ReactNode } from "react"

type PermissionGateProps = {
  permission?: Permission
  action?: string
  children: ReactNode
  fallback?: ReactNode
}

export function PermissionGate({
  permission,
  action,
  children,
  fallback = null,
}: PermissionGateProps) {
  const authorization = useAuthorization()
  const isAllowed = permission
    ? authorization.hasPermission(permission)
    : action
      ? canAccessAction(action, authorization.permissions)
      : false

  if (authorization.status !== "ready" || !isAllowed) {
    return fallback
  }

  return children
}
