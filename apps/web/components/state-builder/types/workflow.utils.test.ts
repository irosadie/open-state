import { describe, expect, it } from "vitest"
import type {
  TransitionDefinition,
  WorkflowDefinition,
  WorkflowNode,
} from "./workflow"
import { validateWorkflow } from "./workflow.utils"

function node(
  id: string,
  kind: WorkflowNode["kind"],
  isTerminal = false,
): WorkflowNode {
  return {
    id,
    kind,
    name: id,
    requiredContext: [],
    capabilities: [],
    policy: {},
    isTerminal,
    position: { x: 0, y: 0 },
  }
}

function transition(
  id: string,
  sourceStateId: string,
  targetStateId: string,
): TransitionDefinition {
  return {
    id,
    sourceStateId,
    targetStateId,
    event: id,
    guards: [],
    priority: 1,
  }
}

function workflow(transitions: TransitionDefinition[]): WorkflowDefinition {
  return {
    slug: "cycle-test",
    name: "Cycle test",
    schemaVersion: 1,
    status: "DRAFT",
    entryNodeId: "start",
    nodes: [
      node("start", "START"),
      node("loop-a", "STATE"),
      node("loop-b", "STATE"),
      node("done", "END", true),
    ],
    transitions,
    policy: { interruptible: "USER_REQUESTED", priority: 1 },
    triggers: [],
  }
}

describe("validateWorkflow cycle detection", () => {
  it("does not warn when a cycle has a path to an END node", () => {
    const result = validateWorkflow(
      workflow([
        transition("start-loop", "start", "loop-a"),
        transition("loop-forward", "loop-a", "loop-b"),
        transition("loop-back", "loop-b", "loop-a"),
        transition("loop-exit", "loop-b", "done"),
      ]),
    )

    expect(result.issues.some((issue) => issue.code === "CYCLE_NO_EXIT")).toBe(
      false,
    )
  })

  it("warns when a cycle cannot reach an END node", () => {
    const result = validateWorkflow(
      workflow([
        transition("start-loop", "start", "loop-a"),
        transition("start-done", "start", "done"),
        transition("loop-forward", "loop-a", "loop-b"),
        transition("loop-back", "loop-b", "loop-a"),
      ]),
    )

    expect(result.issues.some((issue) => issue.code === "CYCLE_NO_EXIT")).toBe(
      true,
    )
  })
})
