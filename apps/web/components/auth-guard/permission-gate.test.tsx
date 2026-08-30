import { hasPermission } from "$/utils/rbac"
import { fireEvent, render, screen } from "@testing-library/react"
import { describe, expect, it, vi } from "vitest"
import { PermissionGate } from "./permission-gate"

const useAuthorizationMock = vi.hoisted(() => vi.fn())

vi.mock("$/providers/authorization-provider", () => ({
  useAuthorization: useAuthorizationMock,
}))

describe("PermissionGate", () => {
  it("does not render a denied action or invoke its handler", () => {
    const onClick = vi.fn()
    useAuthorizationMock.mockReturnValue({
      status: "ready",
      permissions: ["workflow:read"],
      hasPermission: (permission: string) =>
        hasPermission(permission, ["workflow:read"]),
    })

    render(
      <PermissionGate action="workflow:publish">
        <button type="button" onClick={onClick}>
          Publish
        </button>
      </PermissionGate>,
    )

    expect(screen.queryByRole("button", { name: "Publish" })).toBeNull()
    expect(onClick).not.toHaveBeenCalled()
  })

  it("renders and allows a granted action", () => {
    const onClick = vi.fn()
    useAuthorizationMock.mockReturnValue({
      status: "ready",
      permissions: ["workflow:*"] as readonly string[],
      hasPermission: (permission: string) =>
        hasPermission(permission, ["workflow:*"]),
    })

    render(
      <PermissionGate action="workflow:publish">
        <button type="button" onClick={onClick}>
          Publish
        </button>
      </PermissionGate>,
    )

    fireEvent.click(screen.getByRole("button", { name: "Publish" }))
    expect(onClick).toHaveBeenCalledOnce()
  })
})
