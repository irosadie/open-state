import { fireEvent, render, screen, waitFor } from "@testing-library/react"
import { beforeEach, describe, expect, it, vi } from "vitest"

import {
  useAdminAPIKeyCreate,
  useAdminAPIKeyRevoke,
  useAdminAPIKeys,
} from "$/hooks/transactions/use-admin"
import { useProjectsList } from "$/hooks/transactions/use-project"
import { useAuthorization } from "$/providers/authorization-provider"
import type {
  CreateStateMCPAPIKeyResponse,
  StateMCPAPIKeyResponse,
} from "@openstate/types"
import APIKeysPageContent from "./api-keys-page-content"

vi.mock("$/providers/authorization-provider", () => ({
  useAuthorization: vi.fn(),
}))

vi.mock("$/hooks/transactions/use-admin", () => ({
  useAdminAPIKeyCreate: vi.fn(),
  useAdminAPIKeys: vi.fn(),
  useAdminAPIKeyRevoke: vi.fn(),
}))

vi.mock("$/hooks/transactions/use-project", () => ({
  useProjectsList: vi.fn(),
}))

const key: StateMCPAPIKeyResponse = {
  id: "key-1",
  tenantId: "tenant-1",
  name: "Local MCP",
  prefix: "osk_abcdef123456",
  projectIds: ["project-1"],
  defaultProjectId: "project-1",
  scopes: ["state:read"],
  expiresAt: null,
  revokedAt: null,
  lastUsedAt: null,
  createdBy: "user-1",
  createdAt: "2026-08-31T00:00:00Z",
}

const projects = [
  {
    id: "project-1",
    tenantId: "tenant-1",
    name: "Padel",
    slug: "padel",
    status: "ACTIVE" as const,
    createdAt: "2026-08-31T00:00:00Z",
    updatedAt: "2026-08-31T00:00:00Z",
  },
  {
    id: "project-2",
    tenantId: "tenant-1",
    name: "Doctor",
    slug: "doctor",
    status: "ACTIVE" as const,
    createdAt: "2026-08-31T00:00:00Z",
    updatedAt: "2026-08-31T00:00:00Z",
  },
]

const authorization = {
  status: "ready" as const,
  role: "OWNER",
  permissions: ["api_key:read", "api_key:create", "api_key:revoke"],
  hasPermission: (permission: string) =>
    ["api_key:read", "api_key:create", "api_key:revoke"].includes(permission),
  refresh: async () => undefined,
}

beforeEach(() => {
  vi.clearAllMocks()
  vi.mocked(useAuthorization).mockReturnValue(authorization)
  vi.mocked(useAdminAPIKeys).mockReturnValue({
    data: [key],
    isLoading: false,
    isError: false,
    error: null,
    refetch: vi.fn(),
  } as unknown as ReturnType<typeof useAdminAPIKeys>)
  vi.mocked(useAdminAPIKeyCreate).mockReturnValue({
    mutateAsync: vi.fn(),
    isPending: false,
  } as unknown as ReturnType<typeof useAdminAPIKeyCreate>)
  vi.mocked(useAdminAPIKeyRevoke).mockReturnValue({
    mutate: vi.fn(),
    isPending: false,
  } as unknown as ReturnType<typeof useAdminAPIKeyRevoke>)
  vi.mocked(useProjectsList).mockReturnValue({
    data: projects,
    isLoading: false,
    isError: false,
    error: null,
    refetch: vi.fn(),
  } as unknown as ReturnType<typeof useProjectsList>)
})

describe("State MCP API keys page", () => {
  it("renders metadata without exposing the raw secret", () => {
    render(<APIKeysPageContent />)

    expect(screen.getByText("Local MCP")).toBeTruthy()
    expect(screen.getByText("osk_abcdef123456_••••••")).toBeTruthy()
    expect(screen.queryByText(/osk_abcdef123456_secret/)).toBeNull()
  })

  it("creates a key and displays its one-time secret", async () => {
    const result: CreateStateMCPAPIKeyResponse = {
      key: "osk_newprefix_secret",
      apiKey: key,
    }
    const mutateAsync = vi.fn().mockResolvedValue(result)
    vi.mocked(useAdminAPIKeyCreate).mockReturnValue({
      mutateAsync,
      isPending: false,
    } as unknown as ReturnType<typeof useAdminAPIKeyCreate>)

    render(<APIKeysPageContent />)
    fireEvent.change(screen.getByRole("textbox", { name: /^Name/ }), {
      target: { value: "Support bot" },
    })
    const projectInput = screen.getByRole("combobox", {
      name: /^Project IDs/,
    })
    fireEvent.change(projectInput, { target: { value: "project-1" } })
    fireEvent.keyDown(projectInput, { key: "Enter", code: "Enter" })
    fireEvent.click(screen.getByRole("button", { name: "Create API key" }))

    await waitFor(() => expect(mutateAsync).toHaveBeenCalled())
    expect(mutateAsync).toHaveBeenCalledWith({
      name: "Support bot",
      projectIds: ["project-1"],
      defaultProjectId: "project-1",
      scopes: ["state:read"],
      expiresAt: undefined,
    })
    expect(await screen.findByText("osk_newprefix_secret")).toBeTruthy()
  })

  it("supports multiple projects and constrains the default selection", () => {
    render(<APIKeysPageContent />)
    const projectInput = screen.getByRole("combobox", {
      name: /^Project IDs/,
    })
    fireEvent.change(projectInput, { target: { value: "project-1" } })
    fireEvent.keyDown(projectInput, { key: "Enter", code: "Enter" })
    fireEvent.change(projectInput, { target: { value: "project-2" } })
    fireEvent.keyDown(projectInput, { key: "Enter", code: "Enter" })

    expect(
      screen.getByRole("button", {
        name: "Remove Padel (padel) — project-1",
      }),
    ).toBeTruthy()
    expect(
      screen.getByRole("button", {
        name: "Remove Doctor (doctor) — project-2",
      }),
    ).toBeTruthy()
    expect(
      screen
        .getByRole("combobox", { name: /^Default project ID/ })
        .getAttribute("aria-disabled"),
    ).not.toBe("true")
  })

  it("does not query key metadata without read permission", () => {
    vi.mocked(useAuthorization).mockReturnValue({
      ...authorization,
      permissions: [],
      hasPermission: () => false,
    })

    render(<APIKeysPageContent />)

    expect(useAdminAPIKeys).toHaveBeenCalledWith(false)
    expect(
      screen.getByText("You are not authorized to manage State MCP API keys."),
    ).toBeTruthy()
  })
})
