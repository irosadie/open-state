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
export const mcpConnectionHealthStatuses = [
  "unknown",
  "healthy",
  "degraded",
  "unavailable",
  "action_required",
  "circuit_open",
] as const
export const mcpOAuthStatuses = [
  "disconnected",
  "connected",
  "expired",
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
  credentialValue: z.string().optional(),
  oauthAuthorizationEndpoint: z.string().trim().optional(),
  oauthTokenEndpoint: z.string().trim().optional(),
  oauthClientId: z.string().trim().optional(),
  oauthClientSecretValue: z.string().optional(),
  oauthScopes: z.array(z.string().trim()).max(32).optional(),
  oauthRedirectUri: z.string().trim().optional(),
  timeoutMs: z.number().int().min(100).max(300000).optional(),
  maxConcurrency: z.number().int().min(1).max(256).optional(),
  rateLimitPerSecond: z.number().positive().max(10000).optional(),
  rateLimitBurst: z.number().int().min(1).max(10000).optional(),
  retryMax: z.number().int().min(0).max(5).optional(),
  circuitFailureThreshold: z.number().int().min(1).max(100).optional(),
  circuitRecoverySeconds: z.number().int().min(1).max(86400).optional(),
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
    if (value.authType === "oauth") {
      for (const field of [
        "oauthAuthorizationEndpoint",
        "oauthTokenEndpoint",
        "oauthClientId",
        "oauthRedirectUri",
      ] as const) {
        if (!value[field]) {
          ctx.addIssue({
            code: z.ZodIssueCode.custom,
            path: [field],
            message: "This field is required for OAuth",
          })
        }
      }
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
  oauthAuthorizationEndpoint: z.string().nullable(),
  oauthTokenEndpoint: z.string().nullable(),
  oauthClientId: z.string().nullable(),
  oauthScopes: z.array(z.string()),
  oauthRedirectUri: z.string().nullable(),
  oauthStatus: z.enum(mcpOAuthStatuses),
  status: z.enum(mcpConnectionStatuses),
  lastTestStatus: z.enum(mcpConnectionTestStatuses),
  lastTestErrorCode: z.string().nullable(),
  lastTestedAt: z.string().nullable(),
  healthStatus: z.enum(mcpConnectionHealthStatuses),
  healthReason: z.string().nullable(),
  lastSuccessAt: z.string().nullable(),
  consecutiveFailures: z.number().int(),
  circuitOpenedAt: z.string().nullable(),
  timeoutMs: z.number().int(),
  maxConcurrency: z.number().int(),
  rateLimitPerSecond: z.number(),
  rateLimitBurst: z.number().int(),
  retryMax: z.number().int(),
  circuitFailureThreshold: z.number().int(),
  circuitRecoverySeconds: z.number().int(),
  createdAt: z.string(),
  updatedAt: z.string(),
})

export const mcpConnectionListResponseSchema = z.array(
  mcpConnectionResponseSchema,
)
export const mcpOAuthStartResponseSchema = z.object({
  authorizationUrl: z.string().url(),
  status: z.enum(mcpOAuthStatuses),
  expiresAt: z.string(),
})
export const mcpOAuthStatusResponseSchema = z.object({
  status: z.enum(mcpOAuthStatuses),
  expiresAt: z.string().nullable(),
  credentialStatus: z.enum(mcpConnectionCredentialStatuses),
})

export type CreateMCPConnectionSchemaProps = z.infer<
  typeof createMCPConnectionSchema
>
