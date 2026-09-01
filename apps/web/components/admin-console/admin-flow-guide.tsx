"use client"

import {
  Building2Icon,
  FolderIcon,
  ListTreeIcon,
  type LucideIcon,
  MessageSquareIcon,
  PaintbrushIcon,
} from "lucide-react"
import Link from "next/link"

import { PanelCard } from "$/components/panel-card"
import { tenantConfig } from "$/configs/tenant"

export type AdminFlowStep =
  | "tenant"
  | "project"
  | "intent"
  | "workflow"
  | "builder"

type AdminFlowGuideProps = {
  currentStep?: AdminFlowStep
  projectId?: string
  projectName?: string
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
    description: "Choose the business area for the flow",
    href: "/admin/projects",
    icon: FolderIcon,
  },
  {
    key: "intent",
    label: "Intent",
    description: "Choose what the user wants to do",
    href: "/admin/intents",
    icon: MessageSquareIcon,
  },
  {
    key: "workflow",
    label: "Workflow",
    description: "Map the request to a business flow",
    href: "/admin/workflows",
    icon: ListTreeIcon,
  },
  {
    key: "builder",
    label: "State",
    description: "Design states and transitions",
    href: "/state-builder",
    icon: PaintbrushIcon,
  },
] as const

export function AdminFlowGuide({
  currentStep,
  projectId,
  projectName,
}: AdminFlowGuideProps) {
  const getStepHref = (step: FlowStep) => {
    if (
      !step.href ||
      !projectId ||
      (step.key !== "project" &&
        step.key !== "intent" &&
        step.key !== "workflow" &&
        step.key !== "builder")
    ) {
      return step.href
    }

    return `${step.href}?projectId=${encodeURIComponent(projectId)}`
  }

  return (
    <PanelCard
      title="How the workspace fits together"
      description="Follow this path to turn a tenant setup into a published workflow."
    >
      <ol
        aria-label="Admin Console setup path"
        className="grid gap-3 md:grid-cols-5"
      >
        {flowSteps.map((step, index) => {
          const Icon = step.icon
          const isCurrent = step.key === currentStep
          const href = getStepHref(step)
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
              {href ? (
                <Link
                  href={href}
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
            {projectName ??
              (projectId ? "Selected Project" : "Default Project")}{" "}
            <span className="text-blue-700">
              {projectId ? `(${projectId})` : "(automatic)"}
            </span>
          </p>
        </div>
      </div>

      <p className="mt-4 text-xs text-slate-500">
        Choose a project to scope its Intents, Workflows, and States. If no
        project is selected, the console uses Default Project automatically;
        each published Intent maps user language to a Workflow.
      </p>
    </PanelCard>
  )
}
