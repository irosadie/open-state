// Lifecycle of a workflow definition (PRD §9).
export type WorkflowStatus =
  | "DRAFT"
  | "VALIDATING"
  | "VALID"
  | "PUBLISHED"
  | "ARCHIVED"

// Lifecycle of a workflow version snapshot (PRD §9).
export type WorkflowVersionStatus =
  | "DRAFT"
  | "VALIDATING"
  | "VALID"
  | "PUBLISHED"
  | "ARCHIVED"

// Workflow definition root returned by the Builder API (PRD 146).
export type WorkflowResponse = {
  id: string
  tenantId: string
  projectId: string
  slug: string
  name: string
  description?: string | null
  status: WorkflowStatus
  currentVersion: number
  version: number
  definition: Record<string, unknown>
  createdAt: string
  updatedAt: string
}

export type WorkflowValidationIssueResponse = {
  code: string
  message: string
  nodeId?: string
  transitionId?: string
}

export type WorkflowDiffItemResponse = {
  id: string
  definition?: Record<string, unknown>
  changedFields?: string[]
}

export type WorkflowDiffGroupResponse = {
  added: WorkflowDiffItemResponse[]
  removed: WorkflowDiffItemResponse[]
  changed: WorkflowDiffItemResponse[]
}

export type WorkflowDiffResponse = {
  workflowId: string
  baseVersion: number
  targetVersion: number
  nodes: WorkflowDiffGroupResponse
  transitions: WorkflowDiffGroupResponse
}

// Immutable workflow version snapshot (PRD §3.3, §9).
export type WorkflowVersionResponse = {
  id: string
  workflowId: string
  versionNo: number
  status: WorkflowVersionStatus
  isCurrent: boolean
  createdAt: string
  updatedAt: string
  definition: Record<string, unknown>
}
