import { axios } from "$/services/axios"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { renderHook, waitFor } from "@testing-library/react"
import type { PropsWithChildren } from "react"
import { afterEach, describe, expect, it, vi } from "vitest"

import useAdminInstanceCommand from "./use-instance-command"
import useUpdateTenant from "./use-update-tenant"

vi.mock("$/services/axios", () => ({ axios: vi.fn() }))

const tenant = {
  id: "tenant-1",
  name: "Tenant",
  slug: "tenant",
  description: "",
  createdAt: "2026-01-01T00:00:00Z",
  updatedAt: "2026-01-01T00:00:00Z",
}

const instance = {
  id: "instance-1",
  tenantId: "tenant-1",
  workflowId: "workflow-1",
  workflowVersionId: "version-1",
  status: "SUSPENDED",
  version: 2,
  createdAt: "2026-01-01T00:00:00Z",
  updatedAt: "2026-01-01T00:00:00Z",
}

function wrapper(queryClient: QueryClient) {
  return function Wrapper({ children }: PropsWithChildren) {
    return (
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    )
  }
}

afterEach(() => vi.clearAllMocks())

describe("Admin Console hooks", () => {
  it("invalidates tenant data only after a successful update", async () => {
    vi.mocked(axios).mockResolvedValue(tenant)
    const queryClient = new QueryClient({
      defaultOptions: { mutations: { retry: false } },
    })
    const invalidate = vi.spyOn(queryClient, "invalidateQueries")
    const { result } = renderHook(() => useUpdateTenant(), {
      wrapper: wrapper(queryClient),
    })

    await result.current.mutateAsync({
      name: tenant.name,
      slug: tenant.slug,
      description: tenant.description,
    })

    await waitFor(() =>
      expect(invalidate).toHaveBeenCalledWith({ queryKey: ["adminTenant"] }),
    )
  })

  it("refreshes runtime and event queries after a successful command", async () => {
    vi.mocked(axios).mockResolvedValue(instance)
    const queryClient = new QueryClient({
      defaultOptions: { mutations: { retry: false } },
    })
    const invalidate = vi.spyOn(queryClient, "invalidateQueries")
    const { result } = renderHook(() => useAdminInstanceCommand("resume"), {
      wrapper: wrapper(queryClient),
    })

    await result.current.mutateAsync({ id: instance.id })

    await waitFor(() => {
      expect(invalidate).toHaveBeenCalledWith({ queryKey: ["adminInstances"] })
      expect(invalidate).toHaveBeenCalledWith({
        queryKey: ["runtimeInstancesList"],
      })
      expect(invalidate).toHaveBeenCalledWith({ queryKey: ["adminEvents"] })
    })
  })
})
