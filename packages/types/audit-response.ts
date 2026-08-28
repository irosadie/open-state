// Audit trail entry (PRD 50). Safe external representation; never exposes
// secrets.
export type AuditEntryResponse = {
  id: string
  tenantId: string
  actor: string
  action: string
  resourceType: string
  resourceId: string
  before?: Record<string, unknown> | null
  after?: Record<string, unknown> | null
  correlationId?: string | null
  occurredAt: string
}

// Paginated envelope for the audit trail (PRD 50).
export type AuditPageResponse = {
  data: AuditEntryResponse[]
  page: number
  pageSize: number
  total: number
  hasNext: boolean
}
