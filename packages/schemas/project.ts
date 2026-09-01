import { z } from "zod"

export const projectStatuses = ["ACTIVE", "ARCHIVED"] as const

export const projectResponseSchema = z.object({
  id: z.string(),
  tenantId: z.string(),
  name: z.string(),
  slug: z.string(),
  status: z.enum(projectStatuses),
  createdAt: z.string(),
  updatedAt: z.string(),
})

export const projectListResponseSchema = z.array(projectResponseSchema)

export type ProjectResponseSchemaProps = z.infer<typeof projectResponseSchema>
export type ProjectListResponseSchemaProps = z.infer<
  typeof projectListResponseSchema
>
