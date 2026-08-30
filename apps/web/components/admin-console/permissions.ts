import { hasPermission } from "$/utils/rbac"

export const hasAdminPermission = (
  requiredPermission: string,
  permissions: readonly string[],
) => hasPermission(requiredPermission, permissions)
