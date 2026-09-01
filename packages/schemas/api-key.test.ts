import { describe, expect, it } from "vitest"

import {
  createStateMCPAPIKeyResponseSchema,
  createStateMCPAPIKeySchema,
  stateMCPAPIKeyListResponseSchema,
} from "./api-key"

const keyMetadata = {
  id: "key-1",
  tenantId: "tenant-1",
  name: "Local MCP",
  prefix: "osk_abcdef123456",
  projectIds: ["project-1"],
  defaultProjectId: "project-1",
  scopes: ["state:read", "state:write"],
  expiresAt: null,
  revokedAt: null,
  lastUsedAt: null,
  createdBy: "user-1",
  createdAt: "2026-08-31T00:00:00Z",
}

describe("State MCP API-key schemas", () => {
  it("accepts a valid create payload", () => {
    const result = createStateMCPAPIKeySchema.safeParse({
      name: "Local MCP",
      projectIds: ["project-1"],
      defaultProjectId: "project-1",
      scopes: ["state:read"],
    })

    expect(result.success).toBe(true)
  })

  it("rejects empty projects and scopes", () => {
    const result = createStateMCPAPIKeySchema.safeParse({
      name: "Local MCP",
      projectIds: [],
      scopes: [],
    })

    expect(result.success).toBe(false)
  })

  it("validates list and one-time create responses", () => {
    expect(stateMCPAPIKeyListResponseSchema.parse([keyMetadata])).toEqual([
      keyMetadata,
    ])
    expect(
      createStateMCPAPIKeyResponseSchema.parse({
        key: "osk_abcdef_secret",
        apiKey: keyMetadata,
      }).key,
    ).toBe("osk_abcdef_secret")
  })
})
