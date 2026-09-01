import { axios } from "$/services/axios"
import type { CreateStateMCPAPIKeySchemaProps } from "@openstate/schemas"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { renderHook, waitFor } from "@testing-library/react"
import type { PropsWithChildren } from "react"
import { afterEach, describe, expect, it, vi } from "vitest"

import useCreateAPIKey from "./use-create-api-key"
import useRevokeAPIKey from "./use-revoke-api-key"

vi.mock("$/services/axios", () => ({ axios: vi.fn() }))

const key = {
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

const createResponse = { key: "osk_abcdef_secret", apiKey: key }

const wrapper = (queryClient: QueryClient) =>
  function Wrapper({ children }: PropsWithChildren) {
    return (
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    )
  }

afterEach(() => vi.clearAllMocks())

describe("State MCP API-key hooks", () => {
  it("creates a key and invalidates the metadata list", async () => {
    vi.mocked(axios).mockResolvedValue(createResponse)
    const queryClient = new QueryClient({
      defaultOptions: { mutations: { retry: false } },
    })
    const invalidateQueries = vi.spyOn(queryClient, "invalidateQueries")
    const { result } = renderHook(() => useCreateAPIKey(), {
      wrapper: wrapper(queryClient),
    })
    const payload: CreateStateMCPAPIKeySchemaProps = {
      name: "Local MCP",
      projectIds: ["project-1"],
      defaultProjectId: "project-1",
      scopes: ["state:read"],
    }

    await result.current.mutateAsync(payload)

    expect(axios).toHaveBeenCalledWith(
      expect.objectContaining({
        method: "POST",
        url: "/api-keys",
        data: payload,
      }),
    )
    await waitFor(() =>
      expect(invalidateQueries).toHaveBeenCalledWith({
        queryKey: ["adminAPIKeys"],
      }),
    )
  })

  it("revokes a key through the scoped endpoint", async () => {
    vi.mocked(axios).mockResolvedValue(key)
    const queryClient = new QueryClient({
      defaultOptions: { mutations: { retry: false } },
    })
    const { result } = renderHook(() => useRevokeAPIKey(), {
      wrapper: wrapper(queryClient),
    })

    await result.current.mutateAsync("key-1")

    expect(axios).toHaveBeenCalledWith(
      expect.objectContaining({
        method: "POST",
        url: "/api-keys/key-1/revoke",
      }),
    )
  })
})
