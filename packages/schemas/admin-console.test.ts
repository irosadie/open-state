import { describe, expect, it } from "vitest"
import {
  eventPageSchema,
  tenantRoles,
  updateMembershipRoleSchema,
  updateTenantSchema,
} from "./admin-console"

describe("admin console schemas", () => {
  it("accepts valid tenant and role mutations", () => {
    expect(
      updateTenantSchema.parse({
        name: "Acme",
        slug: "acme",
        description: "Tenant",
      }),
    ).toEqual({ name: "Acme", slug: "acme", description: "Tenant" })
    expect(updateMembershipRoleSchema.parse({ role: "OPERATOR" }).role).toBe(
      "OPERATOR",
    )
    expect(tenantRoles).toContain("OWNER")
  })

  it("rejects invalid role and event pagination payloads", () => {
    expect(() =>
      updateMembershipRoleSchema.parse({ role: "SUPERUSER" }),
    ).toThrow()
    expect(() => eventPageSchema.parse({ data: [], page: 0 })).toThrow()
  })
})
