import { axios } from "$/services/axios"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { renderHook, waitFor } from "@testing-library/react"
import type { PropsWithChildren } from "react"
import { afterEach, describe, expect, it, vi } from "vitest"
import useGetRuntimeDebugTrace from "./use-get-runtime-debug-trace"
import useListRuntimeInstances from "./use-list-runtime-instances"

vi.mock("$/services/axios", () => ({ axios: vi.fn() }))

const createWrapper = (client: QueryClient) =>
  function Wrapper({ children }: PropsWithChildren) {
    return <QueryClientProvider client={client}>{children}</QueryClientProvider>
  }

afterEach(() => {
  vi.clearAllMocks()
})

describe("runtime inspector hooks", () => {
  it("sends tenant-scoped list filters and parses the response", async () => {
    vi.mocked(axios).mockResolvedValue({
      data: [],
      page: 2,
      pageSize: 10,
      total: 0,
      hasNext: false,
    })
    const client = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    })
    const { result } = renderHook(
      () =>
        useListRuntimeInstances({
          status: "RUNNING",
          workflowId: "wf-1",
          correlationKey: "conversation-1",
          page: 2,
          pageSize: 10,
        }),
      { wrapper: createWrapper(client) },
    )

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(axios).toHaveBeenCalledWith(
      expect.objectContaining({
        url: "/runtime/instances",
        headers: { "X-Tenant-ID": "00000000-0000-0000-0000-000000000001" },
        params: {
          status: "RUNNING",
          workflowId: "wf-1",
          correlationKey: "conversation-1",
          page: 2,
          pageSize: 10,
        },
      }),
    )
  })

  it("marks a forbidden debug query without presenting it as unavailable", async () => {
    vi.mocked(axios).mockRejectedValue({ status: 403 })
    const client = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    })
    const { result } = renderHook(
      () => useGetRuntimeDebugTrace({ id: "instance-1" }),
      { wrapper: createWrapper(client) },
    )

    await waitFor(() => expect(result.current.isError).toBe(true))
    expect(result.current.isForbidden).toBe(true)
    expect(result.current.data).toBeUndefined()
  })
})
