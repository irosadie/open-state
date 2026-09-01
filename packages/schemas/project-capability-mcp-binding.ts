import { z } from "zod"

export const projectCapabilityMCPBindingHealthStatuses = [
  "ACTIVE",
  "MISSING_MAPPING",
  "CONNECTION_DISABLED",
  "TOOL_DISABLED",
  "TOOL_REMOVED",
  "STALE",
] as const

const jsonObjectSchema = z.record(z.string(), z.unknown())

export const mcpToolOptionResponseSchema = z.object({
  connectionId: z.string(),
  connectionName: z.string(),
  connectionAlias: z.string(),
  connectionStatus: z.enum(["enabled", "disabled"]),
  toolId: z.string(),
  toolName: z.string(),
  toolTitle: z.string().nullable(),
  toolDescription: z.string(),
  inputSchema: jsonObjectSchema,
  toolFingerprint: z.string(),
})

export const mcpToolOptionListResponseSchema = z.array(
  mcpToolOptionResponseSchema,
)

export const projectCapabilityMCPBindingResponseSchema = z.object({
  id: z.string().optional(),
  tenantId: z.string(),
  projectId: z.string(),
  capabilityId: z.string(),
  capabilityName: z.string(),
  capabilityDescription: z.string().nullable(),
  connectionId: z.string().optional(),
  connectionName: z.string().optional(),
  connectionAlias: z.string().optional(),
  connectionStatus: z.enum(["enabled", "disabled"]).optional(),
  toolId: z.string().optional(),
  toolName: z.string().optional(),
  toolTitle: z.string().nullable().optional(),
  toolDescription: z.string().optional(),
  boundToolFingerprint: z.string().optional(),
  currentToolFingerprint: z.string().optional(),
  toolEnabled: z.boolean().optional(),
  toolAvailability: z.enum(["present", "removed"]).optional(),
  toolDriftStatus: z
    .enum(["new", "unchanged", "changed", "removed"])
    .optional(),
  health: z.enum(projectCapabilityMCPBindingHealthStatuses),
  healthReason: z.string(),
  createdAt: z.string().optional(),
  updatedAt: z.string().optional(),
})

export const projectCapabilityMCPBindingListResponseSchema = z.array(
  projectCapabilityMCPBindingResponseSchema,
)

export const upsertProjectCapabilityMCPBindingSchema = z.object({
  connectionId: z.string().min(1),
  toolId: z.string().min(1),
})

export type UpsertProjectCapabilityMCPBindingSchemaProps = z.infer<
  typeof upsertProjectCapabilityMCPBindingSchema
>
