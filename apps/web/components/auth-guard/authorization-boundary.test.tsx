import { render, screen, waitFor } from "@testing-library/react"
import { describe, expect, it, vi } from "vitest"
import { AuthorizationBoundary } from "./authorization-boundary"

const usePathnameMock = vi.hoisted(() => vi.fn())
const replaceMock = vi.hoisted(() => vi.fn())
const useAuthorizationMock = vi.hoisted(() => vi.fn())

vi.mock("next/navigation", () => ({
  usePathname: usePathnameMock,
  useRouter: () => ({ replace: replaceMock }),
}))

vi.mock("$/providers/authorization-provider", () => ({
  useAuthorization: useAuthorizationMock,
}))

describe("AuthorizationBoundary", () => {
  it("redirects unauthenticated protected navigation to login", async () => {
    usePathnameMock.mockReturnValue("/admin/audit")
    useAuthorizationMock.mockReturnValue({
      status: "unauthenticated",
      permissions: [],
      refresh: vi.fn(),
    })

    render(
      <AuthorizationBoundary>
        <p>Protected content</p>
      </AuthorizationBoundary>,
    )

    expect(screen.queryByText("Protected content")).toBeNull()
    await waitFor(() =>
      expect(replaceMock).toHaveBeenCalledWith(
        "/login?callbackUrl=%2Fadmin%2Faudit",
      ),
    )
  })

  it("keeps protected content unmounted while authorization is loading", () => {
    usePathnameMock.mockReturnValue("/admin/audit")
    useAuthorizationMock.mockReturnValue({
      status: "loading",
      permissions: [],
      refresh: vi.fn(),
    })

    render(
      <AuthorizationBoundary>
        <p>Protected content</p>
      </AuthorizationBoundary>,
    )

    expect(screen.getByText("Checking access…")).toBeTruthy()
    expect(screen.queryByText("Protected content")).toBeNull()
  })

  it("shows a stable denied state without mounting protected content", () => {
    usePathnameMock.mockReturnValue("/admin/audit")
    useAuthorizationMock.mockReturnValue({
      status: "ready",
      permissions: [],
      refresh: vi.fn(),
    })

    render(
      <AuthorizationBoundary>
        <p>Protected content</p>
      </AuthorizationBoundary>,
    )

    expect(screen.getByRole("heading", { name: "Access denied" })).toBeTruthy()
    expect(screen.queryByText("Protected content")).toBeNull()
  })
})
