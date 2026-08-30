import { describe, expect, it } from "vitest"
import {
  hasAllPermissions,
  hasAnyPermission,
  hasPermission,
} from "./permissions"

describe("permission matching", () => {
  it("matches exact permissions", () => {
    expect(hasPermission("audit:read", ["audit:read"])).toBe(true)
    expect(hasPermission("audit:write", ["audit:read"])).toBe(false)
  })

  it("matches resource wildcards using server semantics", () => {
    expect(hasPermission("workflow:publish", ["workflow:*"])).toBe(true)
    expect(hasPermission("workflow:read", ["workflow:*extra"])).toBe(false)
    expect(hasPermission("capability:read", ["workflow:*"])).toBe(false)
  })

  it("supports any and all permission checks", () => {
    const granted = ["workflow:read", "instance:read"]

    expect(hasAnyPermission(["audit:read", "instance:read"], granted)).toBe(
      true,
    )
    expect(hasAllPermissions(["workflow:read", "instance:read"], granted)).toBe(
      true,
    )
    expect(hasAllPermissions(["workflow:read", "audit:read"], granted)).toBe(
      false,
    )
  })

  it("defaults unknown or empty requirements to denied", () => {
    expect(hasPermission("", ["*"])).toBe(false)
    expect(hasPermission("unknown:read", [])).toBe(false)
  })
})
