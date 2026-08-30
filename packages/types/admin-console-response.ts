export type TenantRole = "OWNER" | "ADMIN" | "EDITOR" | "OPERATOR" | "VIEWER"

export type TenantResponse = {
  id: string
  name: string
  slug: string
  description: string
  createdAt: string
  updatedAt: string
}

export type TenantMembershipResponse = {
  roleAssignmentId: string
  userId: string
  tenantId: string
  role: TenantRole
  email: string
  name: string
  status: string
  photo?: string | null
  createdAt: string
  updatedAt: string
}

export type TenantMembershipPageResponse = {
  data: TenantMembershipResponse[]
  page: number
  pageSize: number
  total: number
  hasNext: boolean
}

export type InstanceResponse = {
  id: string
  tenantId: string
  workflowId: string
  workflowVersionId: string
  correlationKey?: string | null
  status: string
  version: number
  currentStateInstanceId?: string | null
  startedAt?: string | null
  completedAt?: string | null
  expiresAt?: string | null
  createdAt: string
  updatedAt: string
}

export type EventResponse = {
  id: string
  tenantId: string
  eventId: string
  type: string
  source: string
  aggregateId?: string | null
  workflowInstanceId?: string | null
  sequence: number
  timestamp: string
  payload: Record<string, unknown>
  correlationId?: string | null
  causationId?: string | null
  createdAt: string
}

export type EventPageResponse = {
  data: EventResponse[]
  page: number
  pageSize: number
  total: number
  hasNext: boolean
}
