import { describe, expect, it } from "vitest"
import {
  getSimulationFinalStatusLabel,
  getSimulationOutcomeLabel,
  simulationWorkflowSchema,
} from "./simulation"

describe("Simulation Workflow Schema", () => {
  it("accepts a workflow snapshot with context and ordered events", () => {
    const result = simulationWorkflowSchema.safeParse({
      definition: { nodes: [], transitions: [] },
      initialContext: { customer: { id: "c-1" } },
      events: [{ type: "workflow.started", payload: { source: "builder" } }],
    })

    expect(result.success).toBe(true)
  })

  it("defaults optional context, payload, and events", () => {
    const result = simulationWorkflowSchema.parse({ definition: { nodes: [] } })

    expect(result.initialContext).toEqual({})
    expect(result.events).toEqual([])
  })

  it("rejects an empty event type and more than 100 events", () => {
    expect(
      simulationWorkflowSchema.safeParse({
        definition: {},
        events: [{ type: " " }],
      }).success,
    ).toBe(false)

    expect(
      simulationWorkflowSchema.safeParse({
        definition: {},
        events: Array.from({ length: 101 }, () => ({ type: "tick" })),
      }).success,
    ).toBe(false)
  })
})

describe("Simulation labels", () => {
  it("maps known and unknown labels", () => {
    expect(getSimulationOutcomeLabel("TRANSITIONED")).toBe("Transitioned")
    expect(getSimulationFinalStatusLabel("COMPLETED")).toBe("Completed")
    expect(getSimulationOutcomeLabel("UNKNOWN" as never)).toBe("UNKNOWN")
  })
})
