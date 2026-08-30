import { describe, expect, it } from "vitest"
import { authorizationSnapshotSchema, userRoleSchema } from "./auth"

describe("authorization schemas", () => {
  it("accepts the tenant role and permission snapshot", () => {
    expect(
      authorizationSnapshotSchema.parse({
        role: "EDITOR",
        permissions: ["workflow:read"],
      }),
    ).toEqual({ role: "EDITOR", permissions: ["workflow:read"] })
  })

  it("defaults malformed authorization fields to least privilege", () => {
    expect(
      authorizationSnapshotSchema.parse({
        role: "UNKNOWN",
        permissions: [""],
      }),
    ).toEqual({ role: null, permissions: [] })
  })

  it("keeps the role set tenant-scoped and explicit", () => {
    expect(userRoleSchema.safeParse("OWNER").success).toBe(true)
    expect(userRoleSchema.safeParse("USER").success).toBe(false)
  })
})
