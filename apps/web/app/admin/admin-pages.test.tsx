import { fireEvent, render, screen } from "@testing-library/react"
import type { ReactNode } from "react"
import { beforeEach, describe, expect, it, vi } from "vitest"

import {
  useAdminEvent,
  useAdminEvents,
  useAdminInstanceCommand,
  useAdminInstances,
  useAdminMemberRemove,
  useAdminMemberRoleUpdate,
  useAdminMembers,
  useAdminTenant,
  useAdminTenantUpdate,
} from "$/hooks/transactions/use-admin"
import { useAuthorization } from "$/providers/authorization-provider"
import EventsPageContent from "./events/events-page-content"
import InstancesPageContent from "./instances/instances-page-content"
import TenantPageContent from "./tenant/tenant-page-content"

vi.mock("$/providers/authorization-provider", () => ({
  useAuthorization: vi.fn(),
}))
vi.mock("$/hooks/transactions/use-admin", () => ({
  useAdminEvent: vi.fn(),
  useAdminEvents: vi.fn(),
  useAdminInstanceCommand: vi.fn(),
  useAdminInstances: vi.fn(),
  useAdminMemberRemove: vi.fn(),
  useAdminMemberRoleUpdate: vi.fn(),
  useAdminMembers: vi.fn(),
  useAdminTenant: vi.fn(),
  useAdminTenantUpdate: vi.fn(),
}))
vi.mock("next/link", () => ({
  default: ({ children, href }: { children: ReactNode; href: string }) => (
    <a href={href}>{children}</a>
  ),
}))

const authorization = {
  status: "ready" as const,
  role: "OWNER",
  permissions: ["tenant:*", "user:*", "instance:*"],
  hasPermission: (permission: string) =>
    permission === "tenant:read" ||
    permission === "tenant:update" ||
    permission === "user:read" ||
    permission === "user:update" ||
    permission.startsWith("instance:"),
  refresh: async () => undefined,
}

const tenant = {
  id: "tenant-1",
  name: "Tenant",
  slug: "tenant",
  description: "Description",
  createdAt: "2026-01-01T00:00:00Z",
  updatedAt: "2026-01-01T00:00:00Z",
}

const instance = {
  id: "instance-1",
  tenantId: "tenant-1",
  workflowId: "workflow-1",
  workflowVersionId: "version-1",
  status: "FAILED",
  version: 2,
  createdAt: "2026-01-01T00:00:00Z",
  updatedAt: "2026-01-01T00:00:00Z",
}

beforeEach(() => {
  vi.restoreAllMocks()
  vi.clearAllMocks()
  vi.mocked(useAuthorization).mockReturnValue(authorization)
  vi.mocked(useAdminTenant).mockReturnValue({
    data: tenant,
    isLoading: false,
    isError: false,
    error: null,
  } as unknown as ReturnType<typeof useAdminTenant>)
  vi.mocked(useAdminTenantUpdate).mockReturnValue({
    mutate: vi.fn(),
    isPending: false,
  } as unknown as ReturnType<typeof useAdminTenantUpdate>)
  vi.mocked(useAdminMembers).mockReturnValue({
    data: { data: [], page: 1, pageSize: 20, total: 0, hasNext: false },
    isLoading: false,
    isError: false,
    error: null,
  } as unknown as ReturnType<typeof useAdminMembers>)
  vi.mocked(useAdminMemberRoleUpdate).mockReturnValue({
    mutate: vi.fn(),
    isPending: false,
  } as unknown as ReturnType<typeof useAdminMemberRoleUpdate>)
  vi.mocked(useAdminMemberRemove).mockReturnValue({
    mutate: vi.fn(),
    isPending: false,
  } as unknown as ReturnType<typeof useAdminMemberRemove>)
  vi.mocked(useAdminInstances).mockReturnValue({
    data: [instance],
    isLoading: false,
    isError: false,
    error: null,
  } as unknown as ReturnType<typeof useAdminInstances>)
  vi.mocked(useAdminEvents).mockReturnValue({
    data: { data: [], page: 1, pageSize: 20, total: 0, hasNext: false },
    isLoading: false,
    isError: false,
    error: null,
  } as unknown as ReturnType<typeof useAdminEvents>)
  vi.mocked(useAdminEvent).mockReturnValue({
    data: undefined,
    isLoading: false,
    isError: false,
    error: null,
  } as unknown as ReturnType<typeof useAdminEvent>)
  vi.mocked(useAdminInstanceCommand).mockReturnValue({
    mutate: vi.fn(),
    isPending: false,
  } as unknown as ReturnType<typeof useAdminInstanceCommand>)
})

describe("Admin Console pages", () => {
  it("requires confirmation before tenant mutations", async () => {
    const mutate = vi.fn()
    vi.mocked(useAdminTenantUpdate).mockReturnValue({
      mutate,
      isPending: false,
    } as unknown as ReturnType<typeof useAdminTenantUpdate>)
    vi.spyOn(window, "confirm").mockReturnValue(false)

    render(<TenantPageContent />)
    fireEvent.change(screen.getByLabelText("Name"), {
      target: { value: "Tenant" },
    })
    fireEvent.change(screen.getByLabelText("Slug"), {
      target: { value: "tenant" },
    })
    fireEvent.change(screen.getByLabelText("Description"), {
      target: { value: "Description" },
    })
    fireEvent.submit(
      screen
        .getByRole("button", { name: "Save changes" })
        .closest("form") as HTMLFormElement,
    )

    expect(mutate).not.toHaveBeenCalled()
  })

  it("keeps the event browser read-only", () => {
    render(<EventsPageContent />)

    expect(
      screen.queryByRole("button", { name: /delete|replay|inject|edit/i }),
    ).toBeNull()
    expect(screen.getByText(/read-only/i)).toBeTruthy()
  })

  it("renders lifecycle conflict feedback after a confirmed command", () => {
    const mutation = {
      isPending: false,
      mutate: vi.fn(
        (
          _variables: unknown,
          options?: { onError?: (error: unknown) => void },
        ) => options?.onError?.({ message: "optimistic lock conflict" }),
      ),
    }
    vi.mocked(useAdminInstanceCommand).mockReturnValue(
      mutation as unknown as ReturnType<typeof useAdminInstanceCommand>,
    )
    vi.spyOn(window, "confirm").mockReturnValue(true)

    render(<InstancesPageContent />)
    fireEvent.click(screen.getByRole("button", { name: "Retry" }))

    expect(screen.getByText("optimistic lock conflict")).toBeTruthy()
  })
})
