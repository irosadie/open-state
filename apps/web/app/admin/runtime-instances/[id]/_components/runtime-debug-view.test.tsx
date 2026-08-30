import type { RuntimeTraceResponse } from "@openstate/types"
import { render, screen } from "@testing-library/react"
import { describe, expect, it, vi } from "vitest"
import { RuntimeDebugView } from "./runtime-debug-view"

const trace: RuntimeTraceResponse = {
  available: true,
  data: [
    {
      id: "trace-1",
      turnId: "turn-1",
      sequence: 1,
      stage: "STATE_LOOKUP",
      source: "OPENSTATE",
      status: "STARTED",
      occurredAt: "2026-08-29T00:00:00Z",
      correlationId: "corr-1",
      durationMs: null,
      reasonCode: null,
      errorCode: null,
      providerAlias: null,
      providerReference: null,
      summary: null,
      attributes: {},
    },
    {
      id: "trace-2",
      turnId: "turn-1",
      sequence: 2,
      stage: "RAG_INTEGRATION",
      source: "EXTERNAL_PROVIDER",
      status: "SUCCEEDED",
      occurredAt: "2026-08-29T00:00:00Z",
      correlationId: "corr-1",
      durationMs: 120,
      reasonCode: null,
      errorCode: null,
      providerAlias: "retrieval-service",
      providerReference: "op-123",
      summary: "Retrieved sanitized metadata",
      attributes: { document: "[REDACTED]", resultCount: 2 },
    },
  ],
}

const baseQuery = {
  data: trace,
  error: null,
  isError: false,
  isForbidden: false,
  isLoading: false,
  refetch: vi.fn(),
}

describe("RuntimeDebugView", () => {
  it("renders ordered sanitized provider metadata", () => {
    render(<RuntimeDebugView query={baseQuery} />)

    expect(screen.getByText("STATE_LOOKUP")).toBeTruthy()
    expect(screen.getByText("RAG_INTEGRATION")).toBeTruthy()
    const entries = screen.getAllByTestId(/runtime-debug-entry-/)
    expect(entries[0]?.getAttribute("data-testid")).toBe(
      "runtime-debug-entry-trace-1",
    )
    expect(entries[1]?.getAttribute("data-testid")).toBe(
      "runtime-debug-entry-trace-2",
    )
    expect(screen.getByText("External provider metadata")).toBeTruthy()
    expect(screen.getByText(/REDACTED/)).toBeTruthy()
    expect(screen.getByText("Provider alias: retrieval-service")).toBeTruthy()
  })

  it("distinguishes forbidden and unavailable trace states", () => {
    const { rerender } = render(
      <RuntimeDebugView
        query={{ ...baseQuery, data: undefined, isForbidden: true }}
      />,
    )
    expect(screen.getByText(/not authorized for your role/i)).toBeTruthy()

    rerender(
      <RuntimeDebugView
        query={{ ...baseQuery, data: { available: false, data: [] } }}
      />,
    )
    expect(screen.getByText(/No trace has been recorded/i)).toBeTruthy()
  })
})
