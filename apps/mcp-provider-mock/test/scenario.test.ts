import { describe, expect, test } from "bun:test"

import {
  ScenarioValidationError,
  loadScenario,
  parseScenario,
} from "../src/scenario"

const fixturePath = (name: string) =>
  new URL(`../fixtures/${name}`, import.meta.url).pathname

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

  test("loads every doctor scenario with the compatibility tool catalog", async () => {
    const fixtureNames = [
      "doctor-happy.json",
      "doctor-no-results.json",
      "doctor-unavailable.json",
      "doctor-queue-full.json",
      "doctor-conflict.json",
      "doctor-payment-failed.json",
      "doctor-notification-failed.json",
      "doctor-provider-error.json",
      "doctor-timeout.json",
      "doctor-invalid-output.json",
    ]
    const base = await loadScenario(fixturePath("doctor.json"))
    const expectedTools = base.tools.map((tool) => tool.name)

    for (const fixtureName of fixtureNames) {
      const scenario = await loadScenario(fixturePath(fixtureName))
      expect(scenario.tools.map((tool) => tool.name)).toEqual(expectedTools)
      expect(scenario.provider.name).toContain("doctor-provider-mock")
    }
  })
})
