import { z } from "zod"

// Lifecycle of a workflow definition (PRD §9).
export const workflowStatuses = [
  "DRAFT",
  "VALIDATING",
  "VALID",
  "PUBLISHED",
  "ARCHIVED",
] as const

export const workflowStatusLabels = [
  { label: "Draft", value: "DRAFT" },
  { label: "Validating", value: "VALIDATING" },
  { label: "Valid", value: "VALID" },
  { label: "Published", value: "PUBLISHED" },
  { label: "Archived", value: "ARCHIVED" },
]

export const getWorkflowStatusLabel = (
  value: (typeof workflowStatuses)[number],
) => {
  return workflowStatusLabels.find((l) => l.value === value)?.label ?? value
}

// Lifecycle of a workflow version snapshot (PRD §9).
export const versionStatuses = [
  "DRAFT",
  "VALIDATING",
  "VALID",
  "PUBLISHED",
  "ARCHIVED",
] as const

export const versionStatusLabels = [
  { label: "Draft", value: "DRAFT" },
  { label: "Validating", value: "VALIDATING" },
  { label: "Valid", value: "VALID" },
  { label: "Published", value: "PUBLISHED" },
  { label: "Archived", value: "ARCHIVED" },
]

export const getVersionStatusLabel = (
  value: (typeof versionStatuses)[number],
) => {
  return versionStatusLabels.find((l) => l.value === value)?.label ?? value
}

// Create a workflow definition draft (PRD 146).
export const createWorkflowSchema = z.object({
  projectId: z.string().optional(),
  slug: z.string().min(1, "Slug is required"),
  name: z.string().min(1, "Name is required"),
  description: z.string().optional(),
  definition: z.record(z.string(), z.unknown()),
})

export type CreateWorkflowSchemaProps = z.infer<typeof createWorkflowSchema>

// Update a workflow draft's mutable fields using optimistic concurrency (PRD §31).
export const updateWorkflowSchema = z.object({
  name: z.string().optional(),
  description: z.string().optional(),
  version: z.number().int().min(0),
  definition: z.record(z.string(), z.unknown()),
})

export type UpdateWorkflowSchemaProps = z.infer<typeof updateWorkflowSchema>

// Publish a workflow definition to an immutable, current version (PRD §3.3, §9, §55).
export const publishWorkflowSchema = z.object({
  version: z.number().int().min(0),
})

export type PublishWorkflowSchemaProps = z.infer<typeof publishWorkflowSchema>
