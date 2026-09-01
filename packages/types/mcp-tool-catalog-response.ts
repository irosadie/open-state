export type MCPDiscoveryRunStatus = "succeeded" | "failed"
export type MCPDiscoveredToolAvailability = "present" | "removed"
export type MCPDiscoveredToolDriftStatus =
  | "new"
  | "unchanged"
  | "changed"
  | "removed"

export type MCPDiscoveryRunResponse = {
  id: string
  tenantId: string
  projectId: string
  connectionId: string
  status: MCPDiscoveryRunStatus
  toolCount: number
  catalogFingerprint: string | null
  errorCode: string | null
  startedAt: string
  completedAt: string
  createdBy: string
}

export type MCPDiscoveredToolResponse = {
  id: string
  tenantId: string
  projectId: string
  connectionId: string
  name: string
  title: string | null
  description: string
  inputSchema: Record<string, unknown>
  annotations: Record<string, unknown>
  fingerprint: string
  enabled: boolean
  availability: MCPDiscoveredToolAvailability
  driftStatus: MCPDiscoveredToolDriftStatus
  firstSeenAt: string
  lastSeenAt: string
  removedAt: string | null
  discoveryRunId: string | null
  createdAt: string
  updatedAt: string
}

export type MCPToolCatalogResponse = {
  connectionId: string
  tools: MCPDiscoveredToolResponse[]
  latestRun: MCPDiscoveryRunResponse | null
  lastSuccessfulRun: MCPDiscoveryRunResponse | null
}
