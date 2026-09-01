import type {
  CapabilityResponse,
  MCPToolOptionResponse,
} from "@openstate/types"
import { fireEvent, render, screen } from "@testing-library/react"
import { beforeEach, describe, expect, it, vi } from "vitest"
import { PropertiesPanel } from "./properties-panel"
import { useStateBuilderStore } from "./state-builder.store"
import type { WorkflowDefinition } from "./types/workflow"

const mocks = vi.hoisted(() => ({
  useCapabilitiesList: vi.fn(),
  useListProjectMCPToolOptions: vi.fn(),
  useListProjectMCPBindings: vi.fn(),
  useUpsertProjectMCPBinding: vi.fn(),
  useDeleteProjectMCPBinding: vi.fn(),
  upsertBinding: vi.fn(),
  deleteBinding: vi.fn(),
}))

vi.mock("$/hooks/transactions/use-capability", () => ({
  useCapabilitiesList: mocks.useCapabilitiesList,
}))

vi.mock("$/hooks/transactions/use-project-mcp-binding", () => ({
  useListProjectMCPToolOptions: mocks.useListProjectMCPToolOptions,
  useListProjectMCPBindings: mocks.useListProjectMCPBindings,
  useUpsertProjectMCPBinding: mocks.useUpsertProjectMCPBinding,
  useDeleteProjectMCPBinding: mocks.useDeleteProjectMCPBinding,
}))

const projectID = "00000000-0000-0000-0000-000000000001"
const nodeID = "state-check-availability"

const capability = {
  id: "00000000-0000-0000-0000-000000000002",
  tenantId: projectID,
  name: "padel.availability.read",
  providerType: "MCP",
  status: "ACTIVE",
  version: 1,
  createdAt: "2026-01-01T00:00:00Z",
  updatedAt: "2026-01-01T00:00:00Z",
} as CapabilityResponse

const internalCapability = {
  ...capability,
  id: "00000000-0000-0000-0000-000000000003",
  name: "state.context.read",
  providerType: "INTERNAL",
} as CapabilityResponse

const toolOptions: MCPToolOptionResponse[] = [
  {
    connectionId: "00000000-0000-0000-0000-000000000010",
    connectionName: "Padel Provider",
    connectionAlias: "padel-provider",
    connectionStatus: "enabled",
    toolId: "00000000-0000-0000-0000-000000000011",
    toolName: "padel.check_available",
    toolTitle: null,
    toolDescription: "Check court availability",
    inputSchema: {},
    toolFingerprint: "fingerprint-a",
  },
  {
    connectionId: "00000000-0000-0000-0000-000000000020",
    connectionName: "Other Provider",
    connectionAlias: "other-provider",
    connectionStatus: "enabled",
    toolId: "00000000-0000-0000-0000-000000000021",
    toolName: "other.lookup",
    toolTitle: null,
    toolDescription: "Other provider tool",
    inputSchema: {},
    toolFingerprint: "fingerprint-b",
  },
]

const baseWorkflow: WorkflowDefinition = {
  slug: "padel-booking",
  name: "Padel Booking",
  schemaVersion: 1,
  status: "DRAFT",
  entryNodeId: nodeID,
  nodes: [
    {
      id: nodeID,
      kind: "STATE",
      name: "Check availability",
      requiredContext: [],
      capabilities: [],
      policy: {},
      position: { x: 0, y: 0 },
    },
  ],
  transitions: [],
  policy: { interruptible: "USER_REQUESTED", priority: 10 },
  triggers: [],
}

function preparePanel({
  options = toolOptions,
  bindings = [],
  nodeCapabilities = [],
}: {
  options?: MCPToolOptionResponse[]
  bindings?: unknown[]
  nodeCapabilities?: string[]
} = {}) {
  const workflow = structuredClone(baseWorkflow)
  const node = workflow.nodes[0]
  if (node) node.capabilities = nodeCapabilities
  useStateBuilderStore.setState({
    workflow,
    nodes: [],
    selectedNodeId: nodeID,
    selectedEdgeId: null,
    activeProjectId: projectID,
  })
  mocks.useCapabilitiesList.mockReturnValue({
    data: [capability, internalCapability],
    isLoading: false,
  })
  mocks.useListProjectMCPToolOptions.mockReturnValue({
    data: options,
    isLoading: false,
    isError: false,
  })
  mocks.useListProjectMCPBindings.mockReturnValue({
    data: bindings,
    isLoading: false,
  })
  mocks.useUpsertProjectMCPBinding.mockReturnValue({
    mutateAsync: mocks.upsertBinding,
    isPending: false,
  })
  mocks.useDeleteProjectMCPBinding.mockReturnValue({
    mutateAsync: mocks.deleteBinding,
    isPending: false,
  })
}

describe("PropertiesPanel MCP bindings", () => {
  beforeEach(() => {
    vi.clearAllMocks()
    preparePanel()
  })

  it("filters discovered tools by the selected project connection", () => {
    const primaryTool = toolOptions[0]
    if (!primaryTool) throw new Error("test fixture is missing a tool")

    render(
      <PropertiesPanel
        selectedNodeId={nodeID}
        selectedEdgeId={null}
        projectId={projectID}
      />,
    )

    fireEvent.change(screen.getByRole("combobox", { name: "MCP capability" }), {
      target: { value: capability.name },
    })
    fireEvent.click(screen.getByRole("button", { name: "Add" }))

    const connection = screen.getByRole("combobox", {
      name: `${capability.name} MCP connection`,
    })
    fireEvent.change(connection, {
      target: { value: primaryTool.connectionId },
    })

    const tool = screen.getByRole("combobox", {
      name: `${capability.name} MCP tool`,
    })
    expect(
      screen.getByRole("option", { name: "padel.check_available" }),
    ).toBeTruthy()
    expect(screen.queryByRole("option", { name: "other.lookup" })).toBeNull()

    fireEvent.change(tool, { target: { value: primaryTool.toolId } })
    expect(mocks.upsertBinding).toHaveBeenCalledWith({
      projectId: projectID,
      capabilityId: capability.id,
      connectionId: primaryTool.connectionId,
      toolId: primaryTool.toolId,
    })
  })

  it("guides operators to MCP Connections when the project catalog is empty", () => {
    preparePanel({ options: [] })
    render(
      <PropertiesPanel
        selectedNodeId={nodeID}
        selectedEdgeId={null}
        projectId={projectID}
      />,
    )

    expect(
      screen
        .getByRole("link", { name: "MCP Connections" })
        .getAttribute("href"),
    ).toBe("/admin/mcp")
    expect(screen.queryByText(/http:\/\//i)).toBeNull()
  })

  it("shows the provider alias and reason when a binding becomes unavailable", () => {
    preparePanel({
      nodeCapabilities: [capability.name],
      bindings: [
        {
          tenantId: projectID,
          projectId: projectID,
          capabilityId: capability.id,
          capabilityName: capability.name,
          connectionId: toolOptions[0]?.connectionId,
          connectionAlias: toolOptions[0]?.connectionAlias,
          toolId: toolOptions[0]?.toolId,
          toolName: toolOptions[0]?.toolName,
          health: "TOOL_REMOVED",
          healthReason: "Tool is no longer present in the latest catalog",
        },
      ],
    })
    render(
      <PropertiesPanel
        selectedNodeId={nodeID}
        selectedEdgeId={null}
        projectId={projectID}
      />,
    )

    expect(screen.getByText(/TOOL REMOVED/i)).toBeTruthy()
    expect(
      screen.getByText(/Tool is no longer present in the latest catalog/i),
    ).toBeTruthy()
  })
})
