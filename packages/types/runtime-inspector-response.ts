export type RuntimeWorkflowResponse = {
  id: string
  name: string
  slug: string
  versionId: string
  version: number
}

export type RuntimeStateResponse = {
  id: string
  key: string
  name: string
  status: string
  enteredAt: string
  exitedAt?: string | null
}

export type RuntimeInstanceSummaryResponse = {
  id: string
  workflow: RuntimeWorkflowResponse
  status: string
  currentState?: RuntimeStateResponse | null
  correlationId?: string | null
  lastActivityAt: string
}

export type RuntimeInstanceListResponse = {
  data: RuntimeInstanceSummaryResponse[]
  page: number
  pageSize: number
  total: number
  hasNext: boolean
}

export type RuntimeContextResponse = {
  available: Record<string, unknown>
  missing: string[]
  redacted: boolean
}

export type RuntimeTimelineEntryResponse = {
  id: string
  kind: string
  type: string
  label: string
  status: string
  sequence: number
  occurredAt: string
  correlationId?: string | null
  reasonCode?: string | null
}

export type RuntimeInstanceDetailResponse = {
  summary: RuntimeInstanceSummaryResponse
  currentState?: RuntimeStateResponse | null
  context: RuntimeContextResponse
  timeline: RuntimeTimelineEntryResponse[]
  auditCorrelationIds: string[]
}

export type RuntimeTraceEntryResponse = {
  id: string
  turnId?: string | null
  sequence: number
  stage: string
  source: string
  status: string
  occurredAt: string
  correlationId?: string | null
  durationMs?: number | null
  reasonCode?: string | null
  errorCode?: string | null
  providerAlias?: string | null
  providerReference?: string | null
  summary?: string | null
  attributes: Record<string, unknown>
}

export type RuntimeTraceResponse = {
  available: boolean
  data: RuntimeTraceEntryResponse[]
}
