export type SimulationStateResponse = {
  id: string
  name: string
  kind: "START" | "STATE" | "DECISION" | "EVENT" | "END"
}

export type SimulationCandidateResponse = {
  transitionId: string
  event: string
  priority: number
  guardPassed: boolean
  guardError?: string
}

export type SimulationCapabilityRequestResponse = {
  name: string
  mock: true
  status: "PLANNED"
}

export type SimulationStepResponse = {
  sequence: number
  outcome: "ENTERED" | "TRANSITIONED" | "REJECTED"
  eventType?: string
  eventPayload?: Record<string, unknown>
  stateBefore: SimulationStateResponse
  stateAfter?: SimulationStateResponse
  candidates: SimulationCandidateResponse[]
  selectedTransitionId?: string
  context: Record<string, unknown>
  capabilityRequests: SimulationCapabilityRequestResponse[]
  errorCode?: "EVENT_NOT_ALLOWED" | "GUARD_FAILED" | "GUARD_EVALUATION_ERROR"
  errorMessage?: string
}

export type SimulationResultResponse = {
  finalState: SimulationStateResponse
  finalContext: Record<string, unknown>
  finalStatus: "RUNNING" | "COMPLETED"
  steps: SimulationStepResponse[]
}
