import { hasAnyPermission, hasPermission } from "./permissions"
import type { Permission } from "./permissions"

export type RoutePolicy = {
  id: string
  pattern: RegExp
  requiredPermissions?: readonly Permission[]
  anyOfPermissions?: readonly Permission[]
  landing?: boolean
}

export type ActionPolicy = {
  id: string
  requiredPermissions: readonly Permission[]
}

export const actionPolicies: readonly ActionPolicy[] = [
  { id: "tenant:read", requiredPermissions: ["tenant:read"] },
  { id: "user:read", requiredPermissions: ["user:read"] },
  { id: "workflow:read", requiredPermissions: ["workflow:read"] },
  { id: "instance:read", requiredPermissions: ["instance:read"] },
  { id: "event:read", requiredPermissions: ["instance:read"] },
  { id: "capability:read", requiredPermissions: ["capability:read"] },
  { id: "binding:read", requiredPermissions: ["binding:read"] },
  { id: "capability:create", requiredPermissions: ["capability:create"] },
  { id: "capability:update", requiredPermissions: ["capability:update"] },
  { id: "capability:delete", requiredPermissions: ["capability:delete"] },
  { id: "capability:invoke", requiredPermissions: ["capability:invoke"] },
  { id: "binding:create", requiredPermissions: ["binding:create"] },
  { id: "binding:delete", requiredPermissions: ["binding:delete"] },
  { id: "workflow:create", requiredPermissions: ["workflow:create"] },
  { id: "workflow:update", requiredPermissions: ["workflow:update"] },
  { id: "workflow:publish", requiredPermissions: ["workflow:publish"] },
  { id: "workflow:simulate", requiredPermissions: ["workflow:simulate"] },
  { id: "instance:suspend", requiredPermissions: ["instance:suspend"] },
  { id: "instance:resume", requiredPermissions: ["instance:resume"] },
  { id: "instance:retry", requiredPermissions: ["instance:retry"] },
  { id: "debug:read", requiredPermissions: ["debug:read"] },
  { id: "tenant:update", requiredPermissions: ["tenant:update"] },
  { id: "user:update", requiredPermissions: ["user:update"] },
  { id: "user:delete", requiredPermissions: ["user:delete"] },
  { id: "audit:read", requiredPermissions: ["audit:read"] },
]

const readPermissions = [
  "workflow:read",
  "instance:read",
  "capability:read",
  "audit:read",
] as const

export const routePolicies: readonly RoutePolicy[] = [
  {
    id: "capabilities",
    pattern: /^\/admin\/capabilities(?:\/.*)?$/,
    requiredPermissions: ["capability:read"],
    landing: true,
  },
  {
    id: "audit",
    pattern: /^\/admin\/audit(?:\/.*)?$/,
    requiredPermissions: ["audit:read"],
    landing: true,
  },
  {
    id: "state-builder",
    pattern: /^\/state-builder(?:\/.*)?$/,
    requiredPermissions: ["workflow:read"],
    landing: true,
  },
  {
    id: "admin-tenant",
    pattern: /^\/admin\/tenant(?:\/.*)?$/,
    requiredPermissions: ["tenant:read"],
    landing: true,
  },
  {
    id: "admin-members",
    pattern: /^\/admin\/members(?:\/.*)?$/,
    requiredPermissions: ["user:read"],
    landing: true,
  },
  {
    id: "admin-workflows",
    pattern: /^\/admin\/workflows(?:\/.*)?$/,
    requiredPermissions: ["workflow:read"],
    landing: true,
  },
  {
    id: "admin-runtime-instances",
    pattern: /^\/admin\/runtime-instances(?:\/.*)?$/,
    requiredPermissions: ["instance:read"],
    landing: true,
  },
  {
    id: "admin-instances",
    pattern: /^\/admin\/instances(?:\/.*)?$/,
    requiredPermissions: ["instance:read"],
    landing: true,
  },
  {
    id: "admin-events",
    pattern: /^\/admin\/events(?:\/.*)?$/,
    requiredPermissions: ["instance:read"],
    landing: true,
  },
  {
    id: "admin",
    pattern: /^\/admin$/,
    anyOfPermissions: [
      "tenant:read",
      "user:read",
      "workflow:read",
      "instance:read",
      "capability:read",
      "audit:read",
    ],
    landing: true,
  },
  {
    id: "home",
    pattern: /^\/$/,
    anyOfPermissions: readPermissions,
    landing: true,
  },
]

const publicRoutes = new Set(["/login", "/register"])

const pathnameOnly = (value: string) => value.split(/[?#]/, 1)[0] || "/"

export const isPublicRoute = (pathname: string) =>
  publicRoutes.has(pathnameOnly(pathname))

export const getRoutePolicy = (pathname: string) => {
  const normalizedPath = pathnameOnly(pathname)

  return routePolicies.find((policy) => policy.pattern.test(normalizedPath))
}

export const getActionPolicy = (id: string) =>
  actionPolicies.find((policy) => policy.id === id)

export const canAccessAction = (
  id: string,
  grantedPermissions: readonly Permission[],
) => {
  const policy = getActionPolicy(id)

  return (
    policy?.requiredPermissions.every((permission) =>
      hasPermission(permission, grantedPermissions),
    ) ?? false
  )
}

export const canAccessRoute = (
  pathname: string,
  grantedPermissions: readonly Permission[],
) => {
  if (isPublicRoute(pathname)) return true

  const policy = getRoutePolicy(pathname)

  if (!policy) return false

  if (policy.requiredPermissions) {
    return policy.requiredPermissions.every((permission) =>
      hasPermission(permission, grantedPermissions),
    )
  }

  return policy.anyOfPermissions
    ? hasAnyPermission(policy.anyOfPermissions, grantedPermissions)
    : false
}

export const sanitizeCallbackPath = (value?: string | null) => {
  if (!value?.startsWith("/") || value.startsWith("//")) return null

  const path = pathnameOnly(value)

  if (isPublicRoute(path)) return null

  return value
}

const fallbackPath = (policy: RoutePolicy) => {
  const paths: Record<string, string> = {
    capabilities: "/admin/capabilities",
    audit: "/admin/audit",
    "state-builder": "/state-builder",
    "admin-tenant": "/admin/tenant",
    "admin-members": "/admin/members",
    "admin-workflows": "/admin/workflows",
    "admin-runtime-instances": "/admin/runtime-instances",
    "admin-instances": "/admin/instances",
    "admin-events": "/admin/events",
    admin: "/admin",
    home: "/",
  }

  return paths[policy.id] ?? "/"
}

export const resolveAuthorizedPath = (
  callbackPath: string | null | undefined,
  grantedPermissions: readonly Permission[],
) => {
  const safeCallback = sanitizeCallbackPath(callbackPath)

  if (safeCallback && canAccessRoute(safeCallback, grantedPermissions)) {
    return safeCallback
  }

  for (const policy of routePolicies) {
    if (!policy.landing) continue

    const candidate = fallbackPath(policy)

    if (canAccessRoute(candidate, grantedPermissions)) {
      return candidate
    }
  }

  return null
}
