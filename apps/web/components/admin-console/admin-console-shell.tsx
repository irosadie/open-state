"use client"

import {
  ActivityIcon,
  BlocksIcon,
  BookOpenIcon,
  ClipboardListIcon,
  CogIcon,
  FileClockIcon,
  LayoutDashboardIcon,
  ListTreeIcon,
  UsersIcon,
} from "lucide-react"
import Link from "next/link"
import { usePathname } from "next/navigation"
import type { ReactNode } from "react"

import { UserMenu } from "$/components/user-menu"
import { useAuthorization } from "$/providers/authorization-provider"
import { canAccessRoute } from "$/utils/rbac"
import { hasAdminPermission } from "./permissions"

type AdminNavItem = {
  href: string
  label: string
  permission?: string
  icon: typeof LayoutDashboardIcon
}

const navItems: readonly AdminNavItem[] = [
  { href: "/admin", label: "Overview", icon: LayoutDashboardIcon },
  {
    href: "/admin/tenant",
    label: "Tenant settings",
    permission: "tenant:read",
    icon: CogIcon,
  },
  {
    href: "/admin/members",
    label: "Members & roles",
    permission: "user:read",
    icon: UsersIcon,
  },
  {
    href: "/admin/workflows",
    label: "Workflows",
    permission: "workflow:read",
    icon: ListTreeIcon,
  },
  {
    href: "/admin/instances",
    label: "Instances",
    permission: "instance:read",
    icon: ActivityIcon,
  },
  {
    href: "/admin/runtime-instances",
    label: "Runtime Inspector",
    permission: "instance:read",
    icon: BookOpenIcon,
  },
  {
    href: "/admin/events",
    label: "Events",
    permission: "instance:read",
    icon: FileClockIcon,
  },
  {
    href: "/admin/audit",
    label: "Audit log",
    permission: "audit:read",
    icon: ClipboardListIcon,
  },
  {
    href: "/admin/capabilities",
    label: "Capabilities",
    permission: "capability:read",
    icon: BlocksIcon,
  },
]

export function AdminConsoleShell({ children }: { children: ReactNode }) {
  const pathname = usePathname()
  const authorization = useAuthorization()
  const permissions = authorization.permissions

  const visibleItems = navItems.filter(
    (item) =>
      canAccessRoute(item.href, permissions) &&
      (!item.permission || hasAdminPermission(item.permission, permissions)),
  )

  return (
    <div className="min-h-screen bg-slate-50 lg:flex">
      <aside className="border-b border-slate-200 bg-white lg:min-h-screen lg:w-64 lg:border-b-0 lg:border-r">
        <div className="flex items-center justify-between px-6 py-5 lg:block">
          <Link href="/admin" className="text-lg font-semibold text-slate-950">
            Admin Console
          </Link>
          <span className="text-xs text-slate-500 lg:mt-1 lg:block">
            Tenant operations
          </span>
        </div>
        <nav
          aria-label="Admin Console"
          className="overflow-x-auto px-3 pb-4 lg:overflow-visible"
        >
          <ul className="flex min-w-max gap-1 lg:block lg:min-w-0">
            {visibleItems.map((item) => {
              const Icon = item.icon
              const active =
                pathname === item.href ||
                (item.href !== "/admin" && pathname.startsWith(`${item.href}/`))

              return (
                <li key={item.href}>
                  <Link
                    href={item.href}
                    aria-current={active ? "page" : undefined}
                    className={`flex items-center gap-3 rounded-md px-3 py-2 text-sm transition-colors lg:mb-1 ${
                      active
                        ? "bg-slate-900 text-white"
                        : "text-slate-600 hover:bg-slate-100 hover:text-slate-950"
                    }`}
                  >
                    <Icon size={16} aria-hidden="true" />
                    {item.label}
                  </Link>
                </li>
              )
            })}
          </ul>
        </nav>
        <div className="px-3 pb-4 mt-auto">
          <UserMenu />
        </div>
      </aside>
      <main className="min-w-0 flex-1">{children}</main>
    </div>
  )
}
