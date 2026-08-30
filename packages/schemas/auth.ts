import { z } from "zod"

export const userRoles = [
  "OWNER",
  "ADMIN",
  "EDITOR",
  "OPERATOR",
  "VIEWER",
] as const

export const userRoleSchema = z.enum(userRoles)
export type UserRole = (typeof userRoles)[number]

export const permissionSchema = z.string().min(1)

export const authorizationSnapshotSchema = z.object({
  role: userRoleSchema.nullable().optional().catch(null),
  permissions: z.array(permissionSchema).optional().default([]).catch([]),
})

export type AuthorizationSnapshot = z.infer<typeof authorizationSnapshotSchema>
