import { describe, expect, it } from "vitest"
import {
  canAccessAction,
  canAccessRoute,
  getRoutePolicy,
  resolveAuthorizedPath,
  sanitizeCallbackPath,
} from "./route-policy"

describe("route policy", () => {
  it("requires the matching read permission for protected routes", () => {
    expect(canAccessRoute("/admin/audit", ["audit:read"])).toBe(true)
    expect(canAccessRoute("/admin/audit", ["capability:read"])).toBe(false)
    expect(canAccessRoute("/state-builder", ["workflow:read"])).toBe(true)
    expect(canAccessRoute("/admin/instances", ["instance:read"])).toBe(true)
    expect(canAccessRoute("/unknown", ["workflow:*"])).toBe(false)
  })

  it("matches dynamic detail routes and public auth routes", () => {
    expect(getRoutePolicy("/admin/capabilities/cap-1")?.id).toBe("capabilities")
    expect(canAccessRoute("/login", [])).toBe(true)
    expect(canAccessRoute("/register", [])).toBe(true)
  })

  it("uses registered action policies and denies unknown actions", () => {
    expect(canAccessAction("workflow:publish", ["workflow:*"])).toBe(true)
    expect(canAccessAction("workflow:publish", ["workflow:read"])).toBe(false)
    expect(canAccessAction("unknown:action", ["unknown:action"])).toBe(false)
  })

  it("keeps callback paths same-origin and non-public", () => {
    expect(sanitizeCallbackPath("/admin/audit?actor=user-1")).toBe(
      "/admin/audit?actor=user-1",
    )
    expect(sanitizeCallbackPath("//evil.example")).toBe(null)
    expect(sanitizeCallbackPath("https://evil.example")).toBe(null)
    expect(sanitizeCallbackPath("/login")).toBe(null)
  })

  it("retains an authorized callback or selects the first authorized landing", () => {
    expect(resolveAuthorizedPath("/state-builder", ["workflow:read"])).toBe(
      "/state-builder",
    )
    expect(
      resolveAuthorizedPath("/admin/capabilities", ["workflow:read"]),
    ).toBe("/state-builder")
    expect(resolveAuthorizedPath(null, ["audit:read"])).toBe("/admin/audit")
    expect(resolveAuthorizedPath(null, ["capability:*"])).toBe(
      "/admin/capabilities",
    )
    expect(resolveAuthorizedPath(null, ["instance:read"])).toBe(
      "/admin/runtime-instances",
    )
    expect(resolveAuthorizedPath(null, [])).toBe(null)
  })
})
