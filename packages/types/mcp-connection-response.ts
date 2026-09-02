export type MCPConnectionTransport = "streamable_http" | "sse" | "stdio"
export type MCPConnectionAuthType = "none" | "bearer" | "oauth"
export type MCPConnectionStatus = "enabled" | "disabled"
export type MCPConnectionCredentialStatus =
  | "configured"
  | "missing"
  | "action_required"
export type MCPConnectionTestStatus = "never" | "ready" | "failed" | "disabled"
export type MCPConnectionHealthStatus =
  | "unknown"
  | "healthy"
  | "degraded"
  | "unavailable"
  | "action_required"
  | "circuit_open"
export type MCPOAuthStatus =
  | "disconnected"
  | "connected"
  | "expired"
  | "action_required"

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
  oauthAuthorizationEndpoint: string | null
  oauthTokenEndpoint: string | null
  oauthClientId: string | null
  oauthScopes: string[]
  oauthRedirectUri: string | null
  oauthStatus: MCPOAuthStatus
  status: MCPConnectionStatus
  lastTestStatus: MCPConnectionTestStatus
  lastTestErrorCode: string | null
  lastTestedAt: string | null
  healthStatus: MCPConnectionHealthStatus
  healthReason: string | null
  lastSuccessAt: string | null
  consecutiveFailures: number
  circuitOpenedAt: string | null
  timeoutMs: number
  maxConcurrency: number
  rateLimitPerSecond: number
  rateLimitBurst: number
  retryMax: number
  circuitFailureThreshold: number
  circuitRecoverySeconds: number
  createdAt: string
  updatedAt: string
}

export type MCPConnectionListResponse = MCPConnectionResponse[]

export type MCPOAuthStartResponse = {
  authorizationUrl: string
  status: MCPOAuthStatus
  expiresAt: string
}

export type MCPOAuthStatusResponse = {
  status: MCPOAuthStatus
  expiresAt: string | null
  credentialStatus: MCPConnectionCredentialStatus
}
