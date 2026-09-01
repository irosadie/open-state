import { z } from "zod"

export const mcpAPIKeyScopes = [
  "state:read",
  "state:write",
  "capability:invoke",
] as const

export const mcpAPIKeyScopeLabels = [
  { label: "Read state", value: "state:read" },
  { label: "Write state", value: "state:write" },
  { label: "Invoke capabilities", value: "capability:invoke" },
] as const

export const getMCPAPIKeyScopeLabel = (
  value: (typeof mcpAPIKeyScopes)[number],
) => mcpAPIKeyScopeLabels.find((item) => item.value === value)?.label ?? value

export const createStateMCPAPIKeySchema = z.object({
  name: z.string().trim().min(1, "API key name is required").max(255),
  projectIds: z
    .array(z.string().trim().min(1, "Project ID is required"))
    .min(1, "At least one project ID is required"),
  defaultProjectId: z.string().trim().min(1).optional(),
  scopes: z.array(z.enum(mcpAPIKeyScopes)).min(1, "Select at least one scope"),
  expiresAt: z.string().datetime({ offset: true }).optional(),
})

export type CreateStateMCPAPIKeySchemaProps = z.infer<
  typeof createStateMCPAPIKeySchema
>

const nullableDate = z.string().nullable()

export const stateMCPAPIKeyResponseSchema = z.object({
  id: z.string(),
  tenantId: z.string(),
  name: z.string(),
  prefix: z.string(),
  projectIds: z.array(z.string()),
  defaultProjectId: z.string().nullable(),
  scopes: z.array(z.enum(mcpAPIKeyScopes)),
  expiresAt: nullableDate,
  revokedAt: nullableDate,
  lastUsedAt: nullableDate,
  createdBy: z.string(),
  createdAt: z.string(),
})

export const stateMCPAPIKeyListResponseSchema = z.array(
  stateMCPAPIKeyResponseSchema,
)

export const createStateMCPAPIKeyResponseSchema = z.object({
  key: z.string().min(1),
  apiKey: stateMCPAPIKeyResponseSchema,
})

export type StateMCPAPIKeyResponseSchemaProps = z.infer<
  typeof stateMCPAPIKeyResponseSchema
>
