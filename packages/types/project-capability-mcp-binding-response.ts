export type ProjectCapabilityMCPBindingHealth =
  | "ACTIVE"
  | "MISSING_MAPPING"
  | "CONNECTION_DISABLED"
  | "TOOL_DISABLED"
  | "TOOL_REMOVED"
  | "STALE"

export type MCPToolOptionResponse = {
  connectionId: string
  connectionName: string
  connectionAlias: string
  connectionStatus: "enabled" | "disabled"
  toolId: string
  toolName: string
  toolTitle: string | null
  toolDescription: string
  inputSchema: Record<string, unknown>
  toolFingerprint: string
}

export type MCPToolOptionListResponse = MCPToolOptionResponse[]

export type ProjectCapabilityMCPBindingResponse = {
  id?: string
  tenantId: string
  projectId: string
  capabilityId: string
  capabilityName: string
  capabilityDescription: string | null
  connectionId?: string
  connectionName?: string
  connectionAlias?: string
  connectionStatus?: "enabled" | "disabled"
  toolId?: string
  toolName?: string
  toolTitle?: string | null
  toolDescription?: string
  boundToolFingerprint?: string
  currentToolFingerprint?: string
  toolEnabled?: boolean
  toolAvailability?: "present" | "removed"
  toolDriftStatus?: "new" | "unchanged" | "changed" | "removed"
  health: ProjectCapabilityMCPBindingHealth
  healthReason: string
  createdAt?: string
  updatedAt?: string
}

export type ProjectCapabilityMCPBindingListResponse =
  ProjectCapabilityMCPBindingResponse[]
