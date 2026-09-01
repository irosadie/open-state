import { render, screen } from "@testing-library/react"
import type { ReactNode } from "react"
import { beforeEach, describe, expect, it, vi } from "vitest"

import { useProjectsList } from "$/hooks/transactions/use-project"
import { useAuthorization } from "$/providers/authorization-provider"
import ProjectsPageContent from "./projects-page-content"

vi.mock("$/providers/authorization-provider", () => ({
  useAuthorization: vi.fn(),
}))
vi.mock("$/hooks/transactions/use-project", () => ({
  useProjectsList: vi.fn(),
}))
vi.mock("next/navigation", () => ({
  useSearchParams: () => new URLSearchParams(),
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

const authorization = {
  status: "ready" as const,
  role: "OWNER",
  permissions: ["workflow:read"],
  hasPermission: (permission: string) => permission === "workflow:read",
  refresh: async () => undefined,
}

const projects = [
  {
    id: "project-1",
    tenantId: "tenant-1",
    name: "Padel",
    slug: "padel",
    status: "ACTIVE" as const,
    createdAt: "2026-01-01T00:00:00Z",
    updatedAt: "2026-01-01T00:00:00Z",
  },
]

beforeEach(() => {
  vi.clearAllMocks()
  vi.mocked(useAuthorization).mockReturnValue(authorization)
  vi.mocked(useProjectsList).mockReturnValue({
    data: projects,
    isLoading: false,
    isError: false,
    error: null,
    refetch: vi.fn(),
  } as unknown as ReturnType<typeof useProjectsList>)
})

describe("ProjectsPageContent", () => {
  it("offers scoped Intent and Workflow destinations", () => {
    render(<ProjectsPageContent />)

    expect(screen.getByText("Padel")).toBeTruthy()
    expect(
      screen.getByRole("link", { name: "Use project" }).getAttribute("href"),
    ).toBe("/admin/intents?projectId=project-1")
    expect(
      screen.getByRole("link", { name: "Open Workflows" }).getAttribute("href"),
    ).toBe("/admin/workflows?projectId=project-1")
    expect(
      screen.getByRole("link", { name: /Project/ }).getAttribute("href"),
    ).toBe("/admin/projects")
  })

  it("does not request projects for an unauthorized user", () => {
    vi.mocked(useAuthorization).mockReturnValue({
      ...authorization,
      permissions: [],
      hasPermission: () => false,
    })

    render(<ProjectsPageContent />)

    expect(
      screen.getByText("You are not authorized to view projects."),
    ).toBeTruthy()
    expect(vi.mocked(useProjectsList)).toHaveBeenCalledWith({ enabled: false })
  })
})
