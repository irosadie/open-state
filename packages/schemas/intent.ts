import { z } from "zod"

export const intentResponseSchema = z.object({
  id: z.string(),
  tenantId: z.string(),
  projectId: z.string(),
  workflowId: z.string(),
  key: z.string(),
  name: z.string(),
  description: z.string(),
  examples: z.array(z.string()),
  workflowSlug: z.string(),
})

// The shared axios client unwraps the API's { data: [...] } envelope for list
// endpoints, so this schema validates the list returned to frontend hooks.
export const intentListResponseSchema = z.array(intentResponseSchema)

export type IntentResponseSchemaProps = z.infer<typeof intentResponseSchema>
export type IntentListResponseSchemaProps = z.infer<
  typeof intentListResponseSchema
>
