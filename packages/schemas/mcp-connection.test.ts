import { describe, expect, it } from "vitest"
import { createMCPConnectionSchema } from "./mcp-connection"

describe("MCP connection schema", () => {
  it("requires transport-specific configuration", () => {
    expect(
      createMCPConnectionSchema.safeParse({
        name: "Padel",
        alias: "padel",
        transport: "streamable_http",
        authType: "none",
        stdioArgs: [],
      }).success,
    ).toBe(false)
    expect(
      createMCPConnectionSchema.safeParse({
        name: "Padel",
        alias: "padel",
        transport: "streamable_http",
        endpoint: "http://localhost:8031/mcp",
        authType: "none",
        stdioArgs: [],
      }).success,
    ).toBe(true)
  })
})
