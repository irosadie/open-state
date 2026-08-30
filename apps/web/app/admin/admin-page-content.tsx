"use client"

import Link from "next/link"

import { AdminFlowGuide, hasAdminPermission } from "$/components/admin-console"
import { ContentTitle } from "$/components/content-title"
import { PanelCard } from "$/components/panel-card"
import { useAuthorization } from "$/providers/authorization-provider"
import {
  ActivityIcon,
  BlocksIcon,
  ClipboardListIcon,
  CogIcon,
  ListTreeIcon,
  UsersIcon,
} from "lucide-react"

const cards = [
  {
    href: "/admin/tenant",
    permission: "tenant:read",
    title: "Tenant settings",
    description: "Review the current tenant profile and configuration.",
    icon: CogIcon,
  },
  {
    href: "/admin/members",
    permission: "user:read",
    title: "Members & roles",
    description: "Review tenant membership and role assignments.",
    icon: UsersIcon,
  },
  {
    href: "/admin/workflows",
    permission: "workflow:read",
    title: "Workflow inventory",
    description: "Open workflows in the Builder lifecycle experience.",
    icon: ListTreeIcon,
  },
  {
    href: "/admin/instances",
    permission: "instance:read",
    title: "Instance operations",
    description:
      "Manage eligible runtime instances and open inspection details.",
    icon: ActivityIcon,
  },
  {
    href: "/admin/audit",
    permission: "audit:read",
    title: "Audit log",
    description: "Review the existing tenant-scoped audit trail.",
    icon: ClipboardListIcon,
  },
  {
    href: "/admin/capabilities",
    permission: "capability:read",
    title: "Capabilities",
    description: "Open the established capability administration surface.",
    icon: BlocksIcon,
  },
] as const

export default function AdminPageContent() {
  const authorization = useAuthorization()
  const visibleCards = cards.filter((card) =>
    hasAdminPermission(card.permission, authorization.permissions),
  )

  return (
    <div className="space-y-6 p-8">
      <ContentTitle title="Admin Console" />
      <p className="max-w-2xl text-sm text-slate-600">
        Manage the current tenant and move into the product areas that own
        workflow authoring, runtime inspection, audit, and capabilities.
      </p>
      <AdminFlowGuide currentStep="tenant" />
      <div className="grid gap-6 md:grid-cols-2 xl:grid-cols-3">
        {visibleCards.map((card) => {
          const Icon = card.icon
          return (
            <Link key={card.href} href={card.href} className="group">
              <PanelCard className="h-full transition-shadow group-hover:shadow-md">
                <Icon
                  className="mb-5 text-slate-700"
                  size={24}
                  aria-hidden="true"
                />
                <h2 className="text-lg font-semibold text-slate-950">
                  {card.title}
                </h2>
                <p className="mt-2 text-sm text-slate-600">
                  {card.description}
                </p>
              </PanelCard>
            </Link>
          )
        })}
      </div>
    </div>
  )
}
