import { describe, expect, it } from "vitest"
import { useStateBuilderStore } from "./state-builder.store"

const simulationResult = {
  finalState: { id: "done", name: "Done", kind: "END" as const },
  finalContext: {},
  finalStatus: "COMPLETED" as const,
  steps: [
    {
      sequence: 0,
      outcome: "ENTERED" as const,
      stateBefore: { id: "start", name: "Start", kind: "START" as const },
      candidates: [],
      context: {},
      capabilityRequests: [],
    },
    {
      sequence: 1,
      outcome: "TRANSITIONED" as const,
      eventType: "finish",
      stateBefore: { id: "start", name: "Start", kind: "START" as const },
      stateAfter: { id: "done", name: "Done", kind: "END" as const },
      candidates: [],
      selectedTransitionId: "transition-1",
      context: {},
      capabilityRequests: [],
    },
  ],
}

describe("State Builder simulation state", () => {
  it("keeps trace focus transient and marks a result stale", () => {
    const store = useStateBuilderStore
    store.getState().resetSimulation()
    store.getState().setSimulationResult(simulationResult, "snapshot-a")

    expect(store.getState().simulationFocusTarget).toEqual({
      nodeIds: ["start"],
      transitionId: null,
    })
    store.getState().selectSimulationStep(1)
    expect(store.getState().simulationFocusTarget).toEqual({
      nodeIds: ["start", "done"],
      transitionId: "transition-1",
    })

    store.getState().markSimulationStale()
    expect(store.getState().simulationStale).toBe(true)
    expect(store.getState().simulationFocusTarget).toBeNull()
    expect(store.getState().getSimulationSnapshot()).not.toHaveProperty(
      "simulationResult",
    )
    store.getState().resetSimulation()
  })
})
