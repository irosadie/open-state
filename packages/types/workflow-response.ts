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
  createdAt: string
  updatedAt: string
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
}
