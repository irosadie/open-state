import { axios } from "$/services/axios"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { renderHook, waitFor } from "@testing-library/react"
import type { PropsWithChildren } from "react"
import { afterEach, describe, expect, it, vi } from "vitest"
import useCreateWorkflow from "./use-create-workflow"
import usePublishWorkflow from "./use-publish-workflow"

vi.mock("$/services/axios", () => ({ axios: vi.fn() }))

const workflow = {
  id: "wf-1",
  tenantId: "tenant-1",
  projectId: "project-1",
  slug: "padel-booking",
  name: "Padel Booking",
  status: "DRAFT" as const,
  currentVersion: 0,
  version: 0,
  definition: { nodes: [], transitions: [] },
  createdAt: "2026-01-01T00:00:00Z",
  updatedAt: "2026-01-01T00:00:00Z",
}

function createWrapper(queryClient: QueryClient) {
  return function Wrapper({ children }: PropsWithChildren) {
    return (
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    )
  }
}

afterEach(() => {
  vi.clearAllMocks()
})

describe("workflow API hooks", () => {
  it("shapes create requests and invalidates list and workflow queries", async () => {
    vi.mocked(axios).mockResolvedValue(workflow)
    const queryClient = new QueryClient({
      defaultOptions: { mutations: { retry: false } },
    })
    const invalidateQueries = vi.spyOn(queryClient, "invalidateQueries")
    const { result } = renderHook(() => useCreateWorkflow(), {
      wrapper: createWrapper(queryClient),
    })

    await result.current.mutateAsync({
      slug: workflow.slug,
      name: workflow.name,
      definition: workflow.definition,
    })

    expect(axios).toHaveBeenCalledWith(
      expect.objectContaining({
        method: "POST",
        data: expect.objectContaining({ definition: workflow.definition }),
      }),
    )
    await waitFor(() => {
      expect(invalidateQueries).toHaveBeenCalledWith({
        queryKey: ["workflowsList"],
      })
      expect(invalidateQueries).toHaveBeenCalledWith({
        queryKey: ["workflowsGet", workflow.id],
      })
    })
  })

  it("publishes with only the expected version and invalidates history and diff", async () => {
    vi.mocked(axios).mockResolvedValue({
      ...workflow,
      id: "version-1",
      workflowId: workflow.id,
      versionNo: 1,
      status: "PUBLISHED",
      isCurrent: true,
    })
    const queryClient = new QueryClient({
      defaultOptions: { mutations: { retry: false } },
    })
    const invalidateQueries = vi.spyOn(queryClient, "invalidateQueries")
    const { result } = renderHook(() => usePublishWorkflow(), {
      wrapper: createWrapper(queryClient),
    })

    await result.current.mutateAsync({
      id: workflow.id,
      payload: { version: workflow.version },
    })

    expect(axios).toHaveBeenCalledWith(
      expect.objectContaining({
        method: "POST",
        data: { version: 0 },
      }),
    )
    expect(vi.mocked(axios).mock.calls[0]?.[0]).not.toHaveProperty(
      "data.definition",
    )
    await waitFor(() => {
      expect(invalidateQueries).toHaveBeenCalledWith({
        queryKey: ["workflowsVersions", workflow.id],
      })
      expect(invalidateQueries).toHaveBeenCalledWith({
        queryKey: ["workflowsCompare", workflow.id],
      })
    })
  })
})
