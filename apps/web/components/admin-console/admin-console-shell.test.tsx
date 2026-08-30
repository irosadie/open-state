import { render, screen } from "@testing-library/react"
import type { ReactNode } from "react"
import { describe, expect, it, vi } from "vitest"

import { useAuthorization } from "$/providers/authorization-provider"
import { AdminConsoleShell } from "./admin-console-shell"

vi.mock("$/providers/authorization-provider", () => ({
  useAuthorization: vi.fn(),
}))

vi.mock("next/navigation", () => ({
  usePathname: vi.fn(() => "/admin"),
  useRouter: vi.fn(() => ({ push: vi.fn() })),
}))

vi.mock("next/link", () => ({
  default: ({
    children,
    href,
    ...props
  }: { children: ReactNode; href: string }) => (
    <a href={href} {...props}>
      {children}
    </a>
  ),
}))

describe("AdminConsoleShell", () => {
  it("hides tenant and member navigation for a read-only user", () => {
    vi.mocked(useAuthorization).mockReturnValue({
      status: "ready",
      role: "VIEWER",
      permissions: ["workflow:read", "instance:read"],
      hasPermission: (permission) =>
        ["workflow:read", "instance:read"].includes(permission),
      refresh: async () => undefined,
    })

    render(
      <AdminConsoleShell>
        <div>content</div>
      </AdminConsoleShell>,
    )

    expect(screen.getByRole("link", { name: "Workflows" })).toBeTruthy()
    expect(screen.getByRole("link", { name: "Instances" })).toBeTruthy()
    expect(screen.queryByRole("link", { name: "Tenant settings" })).toBeNull()
    expect(screen.queryByRole("link", { name: "Members & roles" })).toBeNull()
  })

  it("shows tenant management navigation for an Owner", () => {
    vi.mocked(useAuthorization).mockReturnValue({
      status: "ready",
      role: "OWNER",
      permissions: ["tenant:*", "user:*", "workflow:read"],
      hasPermission: (permission) =>
        permission === "tenant:read" || permission === "user:read",
      refresh: async () => undefined,
    })

    render(
      <AdminConsoleShell>
        <div>content</div>
      </AdminConsoleShell>,
    )

    expect(screen.getByRole("link", { name: "Tenant settings" })).toBeTruthy()
    expect(screen.getByRole("link", { name: "Members & roles" })).toBeTruthy()
  })
})
