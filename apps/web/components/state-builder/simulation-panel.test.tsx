import type { SimulationResultResponse } from "@openstate/types"
import { fireEvent, render, screen } from "@testing-library/react"
import type { ReactNode } from "react"
import { describe, expect, it, vi } from "vitest"
import { SimulationPanel } from "./simulation-panel"

vi.mock("$/providers/authorization-provider", () => ({
  useAuthorization: () => ({
    status: "ready",
    permissions: ["workflow:simulate"],
    hasPermission: () => true,
    refresh: async () => undefined,
  }),
}))

vi.mock("$/components/auth-guard/permission-gate", () => ({
  PermissionGate: ({ children }: { children: ReactNode }) => children,
}))

const baseProps = (): Parameters<typeof SimulationPanel>[0] => ({
  initialContextText: "{}",
  events: [{ id: "event-1", type: "payment.success", payloadText: "{}" }],
  result: null as SimulationResultResponse | null,
  error: null,
  isRunning: false,
  stale: false,
  selectedSequence: null,
  onInitialContextChange: vi.fn(),
  onAddEvent: vi.fn(),
  onUpdateEvent: vi.fn(),
  onRemoveEvent: vi.fn(),
  onRun: vi.fn(),
  onReset: vi.fn(),
  onSelectStep: vi.fn(),
  onClose: vi.fn(),
})

describe("SimulationPanel", () => {
  it("blocks malformed context before sending a simulation", () => {
    const props = { ...baseProps(), initialContextText: "{broken" }
    render(<SimulationPanel {...props} />)

    const runButton = screen.getByRole("button", { name: /run simulation/i })
    expect((runButton as HTMLButtonElement).disabled).toBe(true)
    expect(screen.getByText(/bukan JSON yang valid/i)).toBeTruthy()
    fireEvent.click(runButton)
    expect(props.onRun).not.toHaveBeenCalled()
  })

  it("submits an unsaved event script and highlights mock trace data", () => {
    const props = baseProps()
    props.result = {
      finalState: { id: "end", name: "Done", kind: "END" },
      finalContext: { paid: true },
      finalStatus: "COMPLETED",
      steps: [
        {
          sequence: 0,
          outcome: "ENTERED",
          stateBefore: { id: "start", name: "Start", kind: "START" },
          stateAfter: { id: "start", name: "Start", kind: "START" },
          candidates: [],
          context: {},
          capabilityRequests: [],
        },
        {
          sequence: 1,
          outcome: "TRANSITIONED",
          eventType: "payment.success",
          eventPayload: {},
          stateBefore: { id: "start", name: "Start", kind: "START" },
          stateAfter: { id: "end", name: "Done", kind: "END" },
          candidates: [
            {
              transitionId: "transition-1",
              event: "payment.success",
              priority: 1,
              guardPassed: true,
            },
          ],
          selectedTransitionId: "transition-1",
          context: { paid: true },
          capabilityRequests: [
            { name: "payments.capture", mock: true, status: "PLANNED" },
          ],
        },
      ],
    }
    render(<SimulationPanel {...props} />)

    fireEvent.click(screen.getByRole("button", { name: /run simulation/i }))
    expect(props.onRun).toHaveBeenCalledWith({
      initialContext: {},
      events: [{ type: "payment.success", payload: {} }],
    })
    expect(screen.getByText(/mock · payments\.capture/i)).toBeTruthy()
    expect(screen.getAllByText(/transition-1/i).length).toBeGreaterThan(0)

    fireEvent.click(
      screen.getByText("Event 1").closest("button") as HTMLElement,
    )
    expect(props.onSelectStep).toHaveBeenCalledWith(1)
  })

  it("shows a rejected step without inventing a destination state", () => {
    const props = baseProps()
    props.result = {
      finalState: { id: "start", name: "Start", kind: "START" },
      finalContext: {},
      finalStatus: "RUNNING",
      steps: [
        {
          sequence: 0,
          outcome: "ENTERED",
          stateBefore: { id: "start", name: "Start", kind: "START" },
          candidates: [],
          context: {},
          capabilityRequests: [],
        },
        {
          sequence: 1,
          outcome: "REJECTED",
          eventType: "unknown.event",
          stateBefore: { id: "start", name: "Start", kind: "START" },
          candidates: [],
          context: {},
          capabilityRequests: [],
          errorCode: "EVENT_NOT_ALLOWED",
          errorMessage: "event is not allowed from the current state",
        },
      ],
    }
    render(<SimulationPanel {...props} />)

    expect(screen.getByText("Start → ditolak")).toBeTruthy()
    expect(screen.getByText(/event is not allowed/i)).toBeTruthy()
  })
})
