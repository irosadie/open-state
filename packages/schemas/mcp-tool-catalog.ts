import { z } from "zod"

export const mcpDiscoveryRunStatuses = ["succeeded", "failed"] as const
export const mcpDiscoveredToolAvailabilities = ["present", "removed"] as const
export const mcpDiscoveredToolDriftStatuses = [
  "new",
  "unchanged",
  "changed",
  "removed",
] as const

const jsonObjectSchema = z.record(z.string(), z.unknown())

export const mcpDiscoveryRunResponseSchema = z.object({
  id: z.string(),
  tenantId: z.string(),
  projectId: z.string(),
  connectionId: z.string(),
  status: z.enum(mcpDiscoveryRunStatuses),
  toolCount: z.number(),
  catalogFingerprint: z.string().nullable(),
  errorCode: z.string().nullable(),
  startedAt: z.string(),
  completedAt: z.string(),
  createdBy: z.string(),
})

export const mcpDiscoveredToolResponseSchema = z.object({
  id: z.string(),
  tenantId: z.string(),
  projectId: z.string(),
  connectionId: z.string(),
  name: z.string(),
  title: z.string().nullable(),
  description: z.string(),
  inputSchema: jsonObjectSchema,
  annotations: jsonObjectSchema,
  fingerprint: z.string(),
  enabled: z.boolean(),
  availability: z.enum(mcpDiscoveredToolAvailabilities),
  driftStatus: z.enum(mcpDiscoveredToolDriftStatuses),
  firstSeenAt: z.string(),
  lastSeenAt: z.string(),
  removedAt: z.string().nullable(),
  discoveryRunId: z.string().nullable(),
  createdAt: z.string(),
  updatedAt: z.string(),
})

export const mcpToolCatalogResponseSchema = z.object({
  connectionId: z.string(),
  tools: z.array(mcpDiscoveredToolResponseSchema),
  latestRun: mcpDiscoveryRunResponseSchema.nullable(),
  lastSuccessfulRun: mcpDiscoveryRunResponseSchema.nullable(),
})

export const setMCPToolEnabledSchema = z.object({
  enabled: z.boolean(),
})

export type SetMCPToolEnabledSchemaProps = z.infer<
  typeof setMCPToolEnabledSchema
>
