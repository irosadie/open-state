import { fireEvent, render, screen } from "@testing-library/react"
import { beforeEach, describe, expect, it, vi } from "vitest"

import {
  useCreateMCPConnection,
  useDeleteMCPConnection,
  useDiagnoseMCPConnection,
  useDisableMCPConnection,
  useDisconnectMCPOAuth,
  useEnableMCPConnection,
  useListMCPConnections,
  useListMCPTools,
  useRefreshMCPTools,
  useResetMCPConnectionHealth,
  useSetMCPToolEnabled,
  useStartMCPOAuth,
  useTestMCPConnection,
  useUpdateMCPConnection,
} from "$/hooks/transactions/use-mcp-connection"
import { useProjectsList } from "$/hooks/transactions/use-project"
import { useAuthorization } from "$/providers/authorization-provider"
import MCPConnectionsPageContent from "./mcp-connections-page-content"

const push = vi.fn()

vi.mock("next/navigation", () => ({
  useRouter: () => ({ push }),
  useSearchParams: () => new URLSearchParams("projectId=project-1"),
}))
vi.mock("$/providers/authorization-provider", () => ({
  useAuthorization: vi.fn(),
}))
vi.mock("$/hooks/transactions/use-project", () => ({
  useProjectsList: vi.fn(),
}))
vi.mock("$/hooks/transactions/use-mcp-connection", () => ({
  useCreateMCPConnection: vi.fn(),
  useDeleteMCPConnection: vi.fn(),
  useDiagnoseMCPConnection: vi.fn(),
  useDisconnectMCPOAuth: vi.fn(),
  useDisableMCPConnection: vi.fn(),
  useEnableMCPConnection: vi.fn(),
  useListMCPConnections: vi.fn(),
  useListMCPTools: vi.fn(),
  useRefreshMCPTools: vi.fn(),
  useResetMCPConnectionHealth: vi.fn(),
  useSetMCPToolEnabled: vi.fn(),
  useStartMCPOAuth: vi.fn(),
  useTestMCPConnection: vi.fn(),
  useUpdateMCPConnection: vi.fn(),
}))

const authorization = {
  status: "ready" as const,
  permissions: ["mcp_connection:read", "mcp_connection:create"],
  hasPermission: (permission: string) =>
    permission.startsWith("mcp_connection:"),
}

beforeEach(() => {
  vi.clearAllMocks()
  vi.mocked(useAuthorization).mockReturnValue(authorization as never)
  vi.mocked(useProjectsList).mockReturnValue({
    data: [{ id: "project-1", name: "Padel", slug: "padel" }],
  } as never)
  vi.mocked(useListMCPConnections).mockReturnValue({
    data: [],
    isLoading: false,
    isError: false,
  } as never)
  vi.mocked(useListMCPTools).mockReturnValue({
    data: {
      connectionId: "",
      tools: [],
      latestRun: null,
      lastSuccessfulRun: null,
    },
    isLoading: false,
    isError: false,
  } as never)
  const mutation = { mutateAsync: vi.fn(), isPending: false }
  vi.mocked(useCreateMCPConnection).mockReturnValue(mutation as never)
  vi.mocked(useUpdateMCPConnection).mockReturnValue(mutation as never)
  vi.mocked(useDeleteMCPConnection).mockReturnValue(mutation as never)
  vi.mocked(useDisableMCPConnection).mockReturnValue(mutation as never)
  vi.mocked(useEnableMCPConnection).mockReturnValue(mutation as never)
  vi.mocked(useTestMCPConnection).mockReturnValue(mutation as never)
  vi.mocked(useDiagnoseMCPConnection).mockReturnValue(mutation as never)
  vi.mocked(useResetMCPConnectionHealth).mockReturnValue(mutation as never)
  vi.mocked(useStartMCPOAuth).mockReturnValue(mutation as never)
  vi.mocked(useDisconnectMCPOAuth).mockReturnValue(mutation as never)
  vi.mocked(useRefreshMCPTools).mockReturnValue(mutation as never)
  vi.mocked(useSetMCPToolEnabled).mockReturnValue(mutation as never)
})

describe("MCP connections page", () => {
  it("hides the registry from users without read permission", () => {
    vi.mocked(useAuthorization).mockReturnValue({
      ...authorization,
      hasPermission: () => false,
    } as never)
    render(<MCPConnectionsPageContent />)
    expect(
      screen.getByText("You are not authorized to view MCP connections."),
    ).toBeTruthy()
  })

  it("validates the form before submitting", () => {
    render(<MCPConnectionsPageContent />)
    fireEvent.click(screen.getByRole("button", { name: "Create connection" }))
    expect(screen.getByText("Name is required")).toBeTruthy()
    expect(screen.getByText("Endpoint is required")).toBeTruthy()
  })

  it("renders the stored tool catalog without invoking a provider", () => {
    vi.mocked(useListMCPConnections).mockReturnValue({
      data: [
        {
          id: "connection-1",
          tenantId: "tenant-1",
          projectId: "project-1",
          name: "Padel provider",
          alias: "padel-provider",
          transport: "streamable_http",
          endpoint: "http://localhost:8031/mcp",
          stdioProfile: null,
          stdioArgs: [],
          authType: "none",
          credentialStatus: "configured",
          status: "enabled",
          lastTestStatus: "ready",
          lastTestErrorCode: null,
          lastTestedAt: null,
          createdAt: "2026-01-01T00:00:00Z",
          updatedAt: "2026-01-01T00:00:00Z",
        },
      ],
      isLoading: false,
      isError: false,
    } as never)
    vi.mocked(useListMCPTools).mockReturnValue({
      data: {
        connectionId: "connection-1",
        tools: [
          {
            id: "tool-1",
            tenantId: "tenant-1",
            projectId: "project-1",
            connectionId: "connection-1",
            name: "padel.court.availability",
            title: null,
            description: "Check available courts",
            inputSchema: {
              type: "object",
              properties: { venue_id: { type: "string" } },
            },
            annotations: {},
            fingerprint:
              "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
            enabled: true,
            availability: "present",
            driftStatus: "unchanged",
            firstSeenAt: "2026-01-01T00:00:00Z",
            lastSeenAt: "2026-01-01T00:00:00Z",
            removedAt: null,
            discoveryRunId: "run-1",
            createdAt: "2026-01-01T00:00:00Z",
            updatedAt: "2026-01-01T00:00:00Z",
          },
        ],
        latestRun: null,
        lastSuccessfulRun: null,
      },
      isLoading: false,
      isError: false,
    } as never)

    render(<MCPConnectionsPageContent />)
    fireEvent.click(screen.getByRole("button", { name: "View tools" }))

    expect(screen.getByText("padel.court.availability")).toBeTruthy()
    expect(screen.getByText("venue_id")).toBeTruthy()
    expect(screen.getByText("Check available courts")).toBeTruthy()
  })
})
