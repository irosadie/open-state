"use client"

import {
  Building2Icon,
  FolderIcon,
  ListTreeIcon,
  type LucideIcon,
  PaintbrushIcon,
} from "lucide-react"
import Link from "next/link"

import { PanelCard } from "$/components/panel-card"
import { tenantConfig } from "$/configs/tenant"

export type AdminFlowStep = "tenant" | "project" | "workflow" | "builder"

type AdminFlowGuideProps = {
  currentStep?: AdminFlowStep
}

type FlowStep = {
  key: AdminFlowStep
  label: string
  description: string
  href?: string
  icon: LucideIcon
}

const flowSteps: readonly FlowStep[] = [
  {
    key: "tenant",
    label: "Tenant",
    description: "Organization profile and access",
    href: "/admin/tenant",
    icon: Building2Icon,
  },
  {
    key: "project",
    label: "Project",
    description: "Default Project is used automatically",
    icon: FolderIcon,
  },
  {
    key: "workflow",
    label: "Workflow",
    description: "Create the business flow",
    href: "/admin/workflows",
    icon: ListTreeIcon,
  },
  {
    key: "builder",
    label: "Builder",
    description: "Design states and transitions",
    href: "/state-builder",
    icon: PaintbrushIcon,
  },
] as const

export function AdminFlowGuide({ currentStep }: AdminFlowGuideProps) {
  return (
    <PanelCard
      title="How the workspace fits together"
      description="Follow this path to turn a tenant setup into a published workflow."
    >
      <ol
        aria-label="Admin Console setup path"
        className="grid gap-3 md:grid-cols-4"
      >
        {flowSteps.map((step, index) => {
          const Icon = step.icon
          const isCurrent = step.key === currentStep
          const content = (
            <div
              className={`h-full rounded-lg border p-4 transition-colors ${
                isCurrent
                  ? "border-slate-900 bg-slate-900 text-white"
                  : "border-slate-200 bg-slate-50 text-slate-700"
              }`}
            >
              <div className="flex items-center gap-2">
                <span
                  className={`flex h-7 w-7 items-center justify-center rounded-full text-xs font-semibold ${
                    isCurrent
                      ? "bg-white text-slate-900"
                      : "bg-white text-slate-600"
                  }`}
                >
                  {index + 1}
                </span>
                <Icon size={16} aria-hidden="true" />
              </div>
              <p className="mt-3 font-semibold">{step.label}</p>
              <p
                className={`mt-1 text-xs ${
                  isCurrent ? "text-slate-300" : "text-slate-500"
                }`}
              >
                {step.description}
              </p>
            </div>
          )

          return (
            <li key={step.key}>
              {step.href ? (
                <Link
                  href={step.href}
                  className="block h-full rounded-lg focus:outline-none focus:ring-2 focus:ring-slate-500"
                >
                  {content}
                </Link>
              ) : (
                content
              )}
            </li>
          )
        })}
      </ol>

      <div className="mt-5 grid gap-3 rounded-lg border border-blue-100 bg-blue-50 p-4 text-sm text-blue-950 sm:grid-cols-2">
        <div>
          <p className="font-semibold">Current tenant</p>
          <p className="mt-1 break-all font-mono text-xs text-blue-800">
            {tenantConfig.tenantId}
          </p>
        </div>
        <div>
          <p className="font-semibold">Current project</p>
          <p className="mt-1 text-blue-800">
            Default Project <span className="text-blue-700">(automatic)</span>
          </p>
        </div>
      </div>

      <p className="mt-4 text-xs text-slate-500">
        Project settings and switching are not available yet. New workflows
        created from this console are placed in Default Project.
      </p>
    </PanelCard>
  )
}
