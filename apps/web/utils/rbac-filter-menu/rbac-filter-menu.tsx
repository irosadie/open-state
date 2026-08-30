import { canAccessRoute, getRoutePolicy, hasAnyPermission } from "$/utils/rbac"
import type { ReactNode } from "react"

type Permission = string

export type MenuProps = {
  key: string
  label?: ReactNode
  icon?: ReactNode
  href?: string
  isActive?: boolean
  hasPermissions?: Permission[]
  children?: MenuProps[]
  type?: string
  [key: string]: unknown
}

export type RbacProviderProps<T> = {
  user: T
  permissions: string[]
  children: ReactNode
}

export const rbacFilterMenu = (
  menu: MenuProps[],
  permissions: Permission[],
) => {
  const cleanInvalidParents = (items: MenuProps[]): MenuProps[] => {
    return items
      .map((item) => {
        if (item.children) {
          item.children = cleanInvalidParents(item.children)
        }

        const hasValidChildren = item.children && item.children.length > 0

        return ("isValid" in item && item.isValid) || hasValidChildren
          ? item
          : null
      })
      .filter((item) => item !== null)
      .map((v) => {
        if ("isValid" in v) {
          v.isValid = undefined
        }
        return v
      })
  }

  const recursiveFilter = (items: MenuProps[]) =>
    items
      .map((item) => {
        const newItem: MenuProps = { ...item }

        if (newItem.children?.length) {
          newItem.children = recursiveFilter(newItem.children)
        }

        const hasRoutePolicy = Boolean(
          newItem.href && getRoutePolicy(newItem.href),
        )
        const hasExplicitPermission = newItem.hasPermissions
          ? hasAnyPermission(newItem.hasPermissions, permissions)
          : true
        const hasRoutePermission = newItem.href
          ? !hasRoutePolicy || canAccessRoute(newItem.href, permissions)
          : true
        const isPermissionValid = hasExplicitPermission && hasRoutePermission
        const noHasPermission = !newItem.hasPermissions && !hasRoutePolicy

        if (isPermissionValid) {
          return { ...newItem, isValid: true }
        }
        if (noHasPermission) {
          return { ...newItem, isValid: false }
        }

        return null
      })
      .filter(Boolean) as MenuProps[]

  return cleanInvalidParents(recursiveFilter(menu))
}
