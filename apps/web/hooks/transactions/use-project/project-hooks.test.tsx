import { axios } from "$/services/axios"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { renderHook, waitFor } from "@testing-library/react"
import type { PropsWithChildren } from "react"
import { afterEach, describe, expect, it, vi } from "vitest"
import useProjectsList from "./use-list-projects"

vi.mock("$/services/axios", () => ({ axios: vi.fn() }))

const wrapper = (queryClient: QueryClient) =>
  function Wrapper({ children }: PropsWithChildren) {
    return (
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    )
  }

afterEach(() => vi.clearAllMocks())

describe("project discovery hooks", () => {
  it("loads and validates tenant-scoped projects", async () => {
    vi.mocked(axios).mockResolvedValue([
      {
        id: "project-1",
        tenantId: "tenant-1",
        name: "Padel",
        slug: "padel",
        status: "ACTIVE",
        createdAt: "2026-08-31T00:00:00Z",
        updatedAt: "2026-08-31T00:00:00Z",
      },
    ])
    const queryClient = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    })
    const { result } = renderHook(() => useProjectsList(), {
      wrapper: wrapper(queryClient),
    })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data?.[0]?.slug).toBe("padel")
    expect(axios).toHaveBeenCalledWith(
      expect.objectContaining({
        method: "GET",
        url: "/projects",
        headers: { "X-Tenant-ID": "00000000-0000-0000-0000-000000000001" },
      }),
    )
  })
})
