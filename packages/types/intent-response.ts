// Read-only canonical intent mapping returned by the Admin Console catalog.
export type IntentResponse = {
  id: string
  tenantId: string
  projectId: string
  workflowId: string
  key: string
  name: string
  description: string
  examples: string[]
  workflowSlug: string
}

export type IntentListResponse = IntentResponse[]
