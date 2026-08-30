import { describe, expect, it } from "vitest"
import { type MenuProps, rbacFilterMenu } from "./rbac-filter-menu"

const menu: MenuProps[] = [
  { key: "audit", label: "Audit", href: "/admin/audit" },
  { key: "builder", label: "Builder", href: "/state-builder" },
  {
    key: "capabilities",
    label: "Capabilities",
    href: "/admin/capabilities",
    hasPermissions: ["capability:read"],
  },
]

describe("rbacFilterMenu", () => {
  it("uses registered route policies and shared wildcard semantics", () => {
    const filtered = rbacFilterMenu(menu, ["workflow:*"])

    expect(filtered.map((item) => item.key)).toEqual(["builder"])
  })

  it("keeps route and explicit action requirements aligned", () => {
    const filtered = rbacFilterMenu(menu, ["audit:read", "capability:read"])

    expect(filtered.map((item) => item.key)).toEqual(["audit", "capabilities"])
  })
})
