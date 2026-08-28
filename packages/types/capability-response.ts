export type CapabilityProviderType = "MCP" | "INTERNAL" | "HTTP" | "FUTURE"

export type CapabilityStatus = "ACTIVE" | "INACTIVE" | "DISABLED"

export type BindingScopeType = "TENANT" | "WORKFLOW" | "STATE"

export type BindingPermission = "ALLOW" | "DENY"

// Capability registry entry (PRD §59). Only the credential reference is exposed;
// secret values are never returned.
export type CapabilityResponse = {
  id: string
  tenantId: string
  name: string
  description?: string | null
  providerType: CapabilityProviderType
  providerId?: string | null
  inputSchema?: Record<string, unknown> | null
  outputSchema?: Record<string, unknown> | null
  status: CapabilityStatus
  version: number
  credentialReference?: string | null
  createdAt: string
  updatedAt: string
}

// Capability binding to a tenant/workflow/state scope (PRD §60).
export type CapabilityBindingResponse = {
  id: string
  tenantId: string
  capabilityId: string
  scopeType: BindingScopeType
  scopeId: string
  permission: BindingPermission
  createdAt: string
  updatedAt: string
}

// Normalized result of a sandbox/mock capability invocation (PRD §2064).
export type CapabilityInvocationResultResponse = {
  data?: Record<string, unknown> | null
  fromMock: boolean
  durationMs?: number
  event?: string | null
}

// Classified capability failure (PRD §87). Never a raw provider error.
export type CapabilityErrorResponse = {
  kind?:
    | "TIMEOUT"
    | "UNAUTHORIZED"
    | "VALIDATION"
    | "UNAVAILABLE"
    | "BUSINESS"
    | "EXTERNAL"
    | "INTERNAL"
  code: string
  message: string
}
