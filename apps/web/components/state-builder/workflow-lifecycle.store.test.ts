import { beforeEach, describe, expect, it, vi } from "vitest"
import { useStateBuilderStore } from "./state-builder.store"
import {
  createWorkflowApi,
  publishWorkflowApi,
  updateWorkflowApi,
} from "./utils/workflow-api"

vi.mock("./utils/workflow-api", () => ({
  createWorkflowApi: vi.fn(),
  getWorkflowApi: vi.fn(),
  publishWorkflowApi: vi.fn(),
  updateWorkflowApi: vi.fn(),
}))

const createdWorkflow = {
  id: "wf-1",
  tenantId: "tenant-1",
  projectId: "project-1",
  slug: "padel-booking",
  name: "Padel Booking",
  description: "",
  status: "DRAFT" as const,
  currentVersion: 0,
  version: 0,
  definition: {},
  createdAt: "2026-01-01T00:00:00Z",
  updatedAt: "2026-01-01T00:00:00Z",
}

describe("State Builder workflow persistence", () => {
  beforeEach(() => {
    vi.mocked(createWorkflowApi).mockReset()
    vi.mocked(updateWorkflowApi).mockReset()
    vi.mocked(publishWorkflowApi).mockReset()
    useStateBuilderStore.setState({
      apiWorkflowId: null,
      apiVersion: 0,
      activeProjectId: undefined,
      saveError: null,
      saveConflict: false,
      isSaving: false,
    })
  })

  it("creates a server draft with the complete graph snapshot", async () => {
    vi.mocked(createWorkflowApi).mockResolvedValue(createdWorkflow)

    await useStateBuilderStore.getState().persist()

    expect(createWorkflowApi).toHaveBeenCalledWith(
      expect.objectContaining({
        definition: expect.objectContaining({
          nodes: expect.any(Array),
          transitions: expect.any(Array),
        }),
      }),
    )
    expect(useStateBuilderStore.getState().apiWorkflowId).toBe("wf-1")
  })

  it("updates the server draft with the optimistic version", async () => {
    useStateBuilderStore.setState({ apiWorkflowId: "wf-1", apiVersion: 4 })
    vi.mocked(updateWorkflowApi).mockResolvedValue({
      ...createdWorkflow,
      id: "wf-1",
      version: 5,
    })

    await useStateBuilderStore.getState().persist()

    expect(updateWorkflowApi).toHaveBeenCalledWith(
      expect.objectContaining({ id: "wf-1", version: 4 }),
    )
    expect(useStateBuilderStore.getState().apiVersion).toBe(5)
  })

  it("keeps the selected project on a new server draft", async () => {
    useStateBuilderStore.setState({ activeProjectId: "project-1" })
    vi.mocked(createWorkflowApi).mockResolvedValue(createdWorkflow)

    await useStateBuilderStore.getState().persist()

    expect(createWorkflowApi).toHaveBeenCalledWith(
      expect.objectContaining({ projectId: "project-1" }),
    )
  })

  it("publishes only the persisted draft version and returns its snapshot", async () => {
    vi.mocked(createWorkflowApi).mockResolvedValue(createdWorkflow)
    vi.mocked(publishWorkflowApi).mockResolvedValue({
      ...createdWorkflow,
      id: "version-1",
      workflowId: "wf-1",
      versionNo: 1,
      status: "PUBLISHED",
      isCurrent: true,
      definition: {},
    })

    const published = await useStateBuilderStore.getState().publish()

    expect(publishWorkflowApi).toHaveBeenCalledWith({
      id: "wf-1",
      version: 0,
    })
    expect(vi.mocked(publishWorkflowApi).mock.calls[0]?.[0]).not.toHaveProperty(
      "definition",
    )
    expect(published.versionNo).toBe(1)
  })
})
