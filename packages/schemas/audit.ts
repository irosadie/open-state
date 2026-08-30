import { z } from "zod"

// Audit action (PRD 50). A typed subset of the backend AuditAction set; unknown
// actions from the API are preserved as strings.
export const auditActions = [
  "workflow.published",
  "state.entered",
  "transition.executed",
  "guard.failed",
  "capability.invoked",
  "capability.denied",
  "workflow.suspended",
  "workflow.resumed",
  "human_handoff.created",
  "rbac.role_assigned",
  "rbac.role_updated",
  "rbac.role_removed",
  "authorization.denied",
  "tenant.updated",
  "workflow.retried",
  "binding.created",
  "binding.deleted",
] as const

export const auditActionLabels = [
  { label: "Workflow published", value: "workflow.published" },
  { label: "State entered", value: "state.entered" },
  { label: "Transition executed", value: "transition.executed" },
  { label: "Guard failed", value: "guard.failed" },
  { label: "Capability invoked", value: "capability.invoked" },
  { label: "Capability denied", value: "capability.denied" },
  { label: "Workflow suspended", value: "workflow.suspended" },
  { label: "Workflow resumed", value: "workflow.resumed" },
  { label: "Human handoff created", value: "human_handoff.created" },
  { label: "Role assigned", value: "rbac.role_assigned" },
  { label: "Role updated", value: "rbac.role_updated" },
  { label: "Role removed", value: "rbac.role_removed" },
  { label: "Authorization denied", value: "authorization.denied" },
  { label: "Tenant updated", value: "tenant.updated" },
  { label: "Workflow retried", value: "workflow.retried" },
  { label: "Binding created", value: "binding.created" },
  { label: "Binding deleted", value: "binding.deleted" },
]

export const getAuditActionLabel = (value: string) => {
  return auditActionLabels.find((l) => l.value === value)?.label ?? value
}

// Audit query filters (PRD 50). All optional.
export const auditQuerySchema = z.object({
  action: z.string().optional(),
  resourceType: z.string().optional(),
  resourceId: z.string().optional(),
  actor: z.string().optional(),
  page: z.number().int().positive().optional(),
  pageSize: z.number().int().positive().max(100).optional(),
})

export type AuditQuerySchemaProps = z.infer<typeof auditQuerySchema>
