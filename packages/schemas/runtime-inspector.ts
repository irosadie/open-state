import { z } from "zod"

export const runtimeInstanceStatuses = [
  "CREATED",
  "RUNNING",
  "WAITING",
  "COMPLETED",
  "CANCELLED",
  "FAILED",
  "EXPIRED",
  "ABORTED",
  "SUSPENDED",
] as const

export const runtimeTraceStages = [
  "INTENT_RESOLUTION",
  "WORKFLOW_LOOKUP",
  "STATE_LOOKUP",
  "CONTEXT_RESOLUTION",
  "RAG_INTEGRATION",
  "MCP_ACTIVITY",
  "LLM_INTEGRATION",
  "EVENT_HANDLING",
  "GUARD_EVALUATION",
  "TRANSITION_SELECTION",
] as const

export const runtimeTraceSources = ["OPENSTATE", "EXTERNAL_PROVIDER"] as const

export const runtimeTraceStatuses = [
  "STARTED",
  "SUCCEEDED",
  "FAILED",
  "NOT_RECORDED",
] as const

export const runtimeInstanceQuerySchema = z.object({
  status: z.enum(runtimeInstanceStatuses).optional(),
  workflowId: z.string().min(1).optional(),
  correlationKey: z.string().min(1).optional(),
  page: z.coerce.number().int().positive().optional(),
  pageSize: z.coerce.number().int().positive().max(100).optional(),
})

export const runtimeDebugTraceQuerySchema = z.object({
  turnId: z.string().min(1).optional(),
})

const nullableString = z.string().nullable().optional()
const runtimeStateResponseSchema = z.object({
  id: z.string(),
  key: z.string(),
  name: z.string(),
  status: z.string(),
  enteredAt: z.string(),
  exitedAt: nullableString,
})

export const runtimeWorkflowResponseSchema = z.object({
  id: z.string(),
  name: z.string(),
  slug: z.string(),
  versionId: z.string(),
  version: z.number().int(),
})

export const runtimeInstanceSummaryResponseSchema = z.object({
  id: z.string(),
  workflow: runtimeWorkflowResponseSchema,
  status: z.string(),
  currentState: runtimeStateResponseSchema.nullable().optional(),
  correlationId: nullableString,
  lastActivityAt: z.string(),
})

export const runtimeInstanceListResponseSchema = z.object({
  data: z.array(runtimeInstanceSummaryResponseSchema),
  page: z.number().int(),
  pageSize: z.number().int(),
  total: z.number().int(),
  hasNext: z.boolean(),
})

export const runtimeContextResponseSchema = z.object({
  available: z.record(z.string(), z.unknown()),
  missing: z.array(z.string()),
  redacted: z.boolean(),
})

export const runtimeTimelineEntryResponseSchema = z.object({
  id: z.string(),
  kind: z.string(),
  type: z.string(),
  label: z.string(),
  status: z.string(),
  sequence: z.number().int(),
  occurredAt: z.string(),
  correlationId: nullableString,
  reasonCode: nullableString,
})

export const runtimeInstanceDetailResponseSchema = z.object({
  summary: runtimeInstanceSummaryResponseSchema,
  currentState: runtimeStateResponseSchema.nullable().optional(),
  context: runtimeContextResponseSchema,
  timeline: z.array(runtimeTimelineEntryResponseSchema),
  auditCorrelationIds: z.array(z.string()),
})

export const runtimeTraceEntryResponseSchema = z.object({
  id: z.string(),
  turnId: nullableString,
  sequence: z.number().int(),
  stage: z.enum(runtimeTraceStages),
  source: z.enum(runtimeTraceSources),
  status: z.enum(runtimeTraceStatuses),
  occurredAt: z.string(),
  correlationId: nullableString,
  durationMs: z.number().int().nullable().optional(),
  reasonCode: nullableString,
  errorCode: nullableString,
  providerAlias: nullableString,
  providerReference: nullableString,
  summary: nullableString,
  attributes: z.record(z.string(), z.unknown()),
})

export const runtimeTraceResponseSchema = z.object({
  available: z.boolean(),
  data: z.array(runtimeTraceEntryResponseSchema),
})

export type RuntimeInstanceQuerySchemaProps = z.infer<
  typeof runtimeInstanceQuerySchema
>
export type RuntimeDebugTraceQuerySchemaProps = z.infer<
  typeof runtimeDebugTraceQuerySchema
>
