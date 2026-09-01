import { z } from "zod"

export const mcpConnectionTransports = [
  "streamable_http",
  "sse",
  "stdio",
] as const
export const mcpConnectionAuthTypes = ["none", "bearer", "oauth"] as const
export const mcpConnectionStatuses = ["enabled", "disabled"] as const
export const mcpConnectionTestStatuses = [
  "never",
  "ready",
  "failed",
  "disabled",
] as const
export const mcpConnectionCredentialStatuses = [
  "configured",
  "missing",
  "action_required",
] as const

const baseMCPConnectionSchema = z.object({
  name: z.string().trim().min(1, "Name is required").max(255),
  alias: z
    .string()
    .trim()
    .regex(
      /^[a-z0-9][a-z0-9._-]{1,127}$/,
      "Use 2-128 lowercase letters, numbers, dots, hyphens, or underscores",
    ),
  transport: z.enum(mcpConnectionTransports),
  endpoint: z.string().trim().optional(),
  stdioProfile: z.string().trim().optional(),
  stdioArgs: z.array(z.string()).max(64).default([]),
  authType: z.enum(mcpConnectionAuthTypes),
  credentialReference: z.string().trim().max(255).optional(),
})

export const createMCPConnectionSchema = baseMCPConnectionSchema.superRefine(
  (value, ctx) => {
    if (value.transport === "stdio" && !value.stdioProfile) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        path: ["stdioProfile"],
        message: "STDIO profile is required",
      })
    }
    if (value.transport !== "stdio" && !value.endpoint) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        path: ["endpoint"],
        message: "Endpoint is required",
      })
    }
    if (value.transport === "stdio" && value.endpoint) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        path: ["endpoint"],
        message: "Endpoint is not used by STDIO",
      })
    }
    if (value.authType === "none" && value.credentialReference) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        path: ["credentialReference"],
        message: "Credential reference requires bearer or OAuth",
      })
    }
  },
)

export const mcpConnectionResponseSchema = z.object({
  id: z.string(),
  tenantId: z.string(),
  projectId: z.string(),
  name: z.string(),
  alias: z.string(),
  transport: z.enum(mcpConnectionTransports),
  endpoint: z.string().nullable(),
  stdioProfile: z.string().nullable(),
  stdioArgs: z.array(z.string()),
  authType: z.enum(mcpConnectionAuthTypes),
  credentialStatus: z.enum(mcpConnectionCredentialStatuses),
  status: z.enum(mcpConnectionStatuses),
  lastTestStatus: z.enum(mcpConnectionTestStatuses),
  lastTestErrorCode: z.string().nullable(),
  lastTestedAt: z.string().nullable(),
  createdAt: z.string(),
  updatedAt: z.string(),
})

export const mcpConnectionListResponseSchema = z.array(
  mcpConnectionResponseSchema,
)

export type CreateMCPConnectionSchemaProps = z.infer<
  typeof createMCPConnectionSchema
>
