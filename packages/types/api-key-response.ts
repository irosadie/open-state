export type MCPAPIKeyScope = "state:read" | "state:write" | "capability:invoke"

export type StateMCPAPIKeyResponse = {
  id: string
  tenantId: string
  name: string
  prefix: string
  projectIds: string[]
  defaultProjectId: string | null
  scopes: MCPAPIKeyScope[]
  expiresAt: string | null
  revokedAt: string | null
  lastUsedAt: string | null
  createdBy: string
  createdAt: string
}

export type StateMCPAPIKeyListResponse = StateMCPAPIKeyResponse[]

export type CreateStateMCPAPIKeyResponse = {
  key: string
  apiKey: StateMCPAPIKeyResponse
}
