import { z } from "zod"

export const tenantRoles = [
  "OWNER",
  "ADMIN",
  "EDITOR",
  "OPERATOR",
  "VIEWER",
] as const

export const tenantRoleLabels = [
  { label: "Owner", value: "OWNER" },
  { label: "Admin", value: "ADMIN" },
  { label: "Editor", value: "EDITOR" },
  { label: "Operator", value: "OPERATOR" },
  { label: "Viewer", value: "VIEWER" },
]

export const getTenantRoleLabel = (value: (typeof tenantRoles)[number]) => {
  return tenantRoleLabels.find((item) => item.value === value)?.label ?? value
}

export const updateTenantSchema = z.object({
  name: z.string().trim().min(1, "Tenant name is required").max(255),
  slug: z
    .string()
    .trim()
    .min(1, "Tenant slug is required")
    .max(255)
    .regex(/^[a-z0-9-]+$/, "Use lowercase letters, numbers, and hyphens"),
  description: z.string().trim().max(1000),
})

export type UpdateTenantSchemaProps = z.infer<typeof updateTenantSchema>

export const updateMembershipRoleSchema = z.object({
  role: z.enum(tenantRoles),
})

export type UpdateMembershipRoleSchemaProps = z.infer<
  typeof updateMembershipRoleSchema
>

export const eventBrowserQuerySchema = z.object({
  workflowInstanceId: z.string().optional(),
  type: z.string().optional(),
  source: z.string().optional(),
  correlationId: z.string().optional(),
  page: z.number().int().positive().optional(),
  pageSize: z.number().int().positive().max(100).optional(),
})

export type EventBrowserQuerySchemaProps = z.infer<
  typeof eventBrowserQuerySchema
>

const nullableString = z.string().nullable().optional()
const jsonObject = z.record(z.string(), z.unknown())

export const tenantResponseSchema = z.object({
  id: z.string(),
  name: z.string(),
  slug: z.string(),
  description: z.string(),
  createdAt: z.string(),
  updatedAt: z.string(),
})

export const tenantMembershipResponseSchema = z.object({
  roleAssignmentId: z.string(),
  userId: z.string(),
  tenantId: z.string(),
  role: z.enum(tenantRoles),
  email: z.string(),
  name: z.string(),
  status: z.string(),
  photo: nullableString,
  createdAt: z.string(),
  updatedAt: z.string(),
})

export const tenantMembershipPageSchema = z.object({
  data: z.array(tenantMembershipResponseSchema),
  page: z.number(),
  pageSize: z.number(),
  total: z.number(),
  hasNext: z.boolean(),
})

export const instanceResponseSchema = z.object({
  id: z.string(),
  tenantId: z.string(),
  workflowId: z.string(),
  workflowVersionId: z.string(),
  correlationKey: nullableString,
  status: z.string(),
  version: z.number(),
  currentStateInstanceId: nullableString,
  startedAt: nullableString,
  completedAt: nullableString,
  expiresAt: nullableString,
  createdAt: z.string(),
  updatedAt: z.string(),
})

export const eventResponseSchema = z.object({
  id: z.string(),
  tenantId: z.string(),
  eventId: z.string(),
  type: z.string(),
  source: z.string(),
  aggregateId: nullableString,
  workflowInstanceId: nullableString,
  sequence: z.number(),
  timestamp: z.string(),
  payload: jsonObject,
  correlationId: nullableString,
  causationId: nullableString,
  createdAt: z.string(),
})

export const eventPageSchema = z.object({
  data: z.array(eventResponseSchema),
  page: z.number(),
  pageSize: z.number(),
  total: z.number(),
  hasNext: z.boolean(),
})
