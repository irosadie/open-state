export type MCPConnectionTransport = "streamable_http" | "sse" | "stdio"
export type MCPConnectionAuthType = "none" | "bearer" | "oauth"
export type MCPConnectionStatus = "enabled" | "disabled"
export type MCPConnectionCredentialStatus =
  | "configured"
  | "missing"
  | "action_required"
export type MCPConnectionTestStatus = "never" | "ready" | "failed" | "disabled"

export type MCPConnectionResponse = {
  id: string
  tenantId: string
  projectId: string
  name: string
  alias: string
  transport: MCPConnectionTransport
  endpoint: string | null
  stdioProfile: string | null
  stdioArgs: string[]
  authType: MCPConnectionAuthType
  credentialStatus: MCPConnectionCredentialStatus
  status: MCPConnectionStatus
  lastTestStatus: MCPConnectionTestStatus
  lastTestErrorCode: string | null
  lastTestedAt: string | null
  createdAt: string
  updatedAt: string
}

export type MCPConnectionListResponse = MCPConnectionResponse[]
