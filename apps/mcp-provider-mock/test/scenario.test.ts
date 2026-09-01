import { describe, expect, test } from "bun:test"

import { ScenarioValidationError, parseScenario } from "../src/scenario"

describe("scenario contract", () => {
  test("rejects duplicate tool names", () => {
    expect(() =>
      parseScenario({
        provider: { name: "provider", version: "1.0.0" },
        tools: [
          {
            description: "first",
            inputSchema: { type: "object" },
            name: "padel.cek_available",
            operation: "static",
          },
          {
            description: "second",
            inputSchema: { type: "object" },
            name: "padel.cek_available",
            operation: "static",
          },
        ],
      }),
    ).toThrow(ScenarioValidationError)
  })

  test("rejects tools without an MCP input schema", () => {
    expect(() =>
      parseScenario({
        provider: { name: "provider", version: "1.0.0" },
        tools: [
          {
            description: "invalid",
            inputSchema: "not-a-schema",
            name: "padel.cek_available",
            operation: "static",
          },
        ],
      }),
    ).toThrow(ScenarioValidationError)
  })
})
