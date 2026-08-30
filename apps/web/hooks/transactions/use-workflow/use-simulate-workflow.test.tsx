import { axios } from "$/services/axios"
import type { SimulationResultResponse } from "@openstate/types"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { renderHook } from "@testing-library/react"
import type { PropsWithChildren } from "react"
import { afterEach, describe, expect, it, vi } from "vitest"
import useSimulateWorkflow from "./use-simulate-workflow"

vi.mock("$/services/axios", () => ({ axios: vi.fn() }))

const resultFixture: SimulationResultResponse = {
  finalState: { id: "start", name: "START", kind: "START" },
  finalContext: {},
  finalStatus: "RUNNING",
  steps: [],
}

function createWrapper(client: QueryClient) {
  return function Wrapper({ children }: PropsWithChildren) {
    return <QueryClientProvider client={client}>{children}</QueryClientProvider>
  }
}

afterEach(() => {
  vi.clearAllMocks()
})

describe("useSimulateWorkflow", () => {
  it("posts the current snapshot with the tenant header", async () => {
    vi.mocked(axios).mockResolvedValue(resultFixture)
    const client = new QueryClient({
      defaultOptions: { mutations: { retry: false } },
    })
    const { result } = renderHook(() => useSimulateWorkflow(), {
      wrapper: createWrapper(client),
    })

    await result.current.mutateAsync({
      definition: { slug: "unsaved", nodes: [], transitions: [] },
      initialContext: { actor: "operator" },
      events: [{ type: "workflow.started", payload: {} }],
    })

    expect(axios).toHaveBeenCalledWith(
      expect.objectContaining({
        method: "POST",
        url: "/workflows/simulate",
        headers: { "X-Tenant-ID": "00000000-0000-0000-0000-000000000001" },
      }),
    )
  })

  it("rejects malformed event scripts before issuing a request", async () => {
    const client = new QueryClient({
      defaultOptions: { mutations: { retry: false } },
    })
    const { result } = renderHook(() => useSimulateWorkflow(), {
      wrapper: createWrapper(client),
    })

    await expect(
      result.current.mutateAsync({
        definition: {},
        events: [{ type: " " }],
      }),
    ).rejects.toThrow()
    expect(axios).not.toHaveBeenCalled()
  })
})
