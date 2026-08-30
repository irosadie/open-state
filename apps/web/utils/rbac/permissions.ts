export type Permission = string

const isResourceWildcard = (permission: string) => permission.endsWith(":*")

export const hasPermission = (
  requiredPermission: Permission,
  grantedPermissions: readonly Permission[],
) => {
  if (!requiredPermission) return false

  return grantedPermissions.some(
    (grantedPermission) =>
      grantedPermission === requiredPermission ||
      (isResourceWildcard(grantedPermission) &&
        requiredPermission.startsWith(`${grantedPermission.slice(0, -1)}`)),
  )
}

export const hasAnyPermission = (
  requiredPermissions: readonly Permission[],
  grantedPermissions: readonly Permission[],
) =>
  requiredPermissions.some((permission) =>
    hasPermission(permission, grantedPermissions),
  )

export const hasAllPermissions = (
  requiredPermissions: readonly Permission[],
  grantedPermissions: readonly Permission[],
) =>
  requiredPermissions.every((permission) =>
    hasPermission(permission, grantedPermissions),
  )
