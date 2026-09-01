import { fireEvent, render, screen } from "@testing-library/react"
import type { ReactNode } from "react"
import { beforeEach, describe, expect, it, vi } from "vitest"

import { useIntentsList } from "$/hooks/transactions/use-intent"
import { useAuthorization } from "$/providers/authorization-provider"
import IntentsPageContent from "./intents-page-content"

vi.mock("$/providers/authorization-provider", () => ({
  useAuthorization: vi.fn(),
}))
vi.mock("$/hooks/transactions/use-intent", () => ({
  useIntentsList: vi.fn(),
}))
vi.mock("next/link", () => ({
  default: ({
    children,
    href,
    ...props
  }: { children: ReactNode; href: string }) => (
    <a href={href} {...props}>
      {children}
    </a>
  ),
}))
vi.mock("next/navigation", () => ({
  useSearchParams: () => new URLSearchParams(),
}))

const authorization = {
  status: "ready" as const,
  role: "OWNER",
  permissions: ["workflow:read"],
  hasPermission: (permission: string) => permission === "workflow:read",
  refresh: async () => undefined,
}

const bookingIntent = {
  id: "intent-1",
  tenantId: "tenant-1",
  projectId: "project-1",
  workflowId: "workflow-1",
  key: "BOOKING_PADEL",
  name: "Booking Padel",
  description: "Book a padel court",
  examples: ["saya mau order lapangan"],
  workflowSlug: "booking-padel",
}

const queryResult = (overrides: Record<string, unknown> = {}) =>
  ({
    data: [bookingIntent],
    isLoading: false,
    isError: false,
    error: null,
    refetch: vi.fn(),
    ...overrides,
  }) as unknown as ReturnType<typeof useIntentsList>

beforeEach(() => {
  vi.clearAllMocks()
  vi.mocked(useAuthorization).mockReturnValue(authorization)
  vi.mocked(useIntentsList).mockReturnValue(queryResult())
})

describe("IntentsPageContent", () => {
  it("renders BOOKING_PADEL examples and its Builder destination", () => {
    render(<IntentsPageContent />)

    expect(screen.getByText("BOOKING_PADEL")).toBeTruthy()
    expect(screen.getByText(/saya mau order lapangan/)).toBeTruthy()
    expect(screen.getByText("booking-padel")).toBeTruthy()
    expect(
      screen.getByRole("link", { name: /Open Builder/ }).getAttribute("href"),
    ).toBe("/state-builder/workflow-1")
    expect(screen.getAllByText(/Default Project/).length).toBeGreaterThan(0)
  })

  it("renders an empty state without a misleading workflow list", () => {
    vi.mocked(useIntentsList).mockReturnValue(queryResult({ data: [] }))

    render(<IntentsPageContent />)

    expect(screen.getByText("No published intents yet")).toBeTruthy()
    expect(screen.queryByText("booking-padel")).toBeNull()
  })

  it("renders a recoverable API error", () => {
    const refetch = vi.fn()
    vi.mocked(useIntentsList).mockReturnValue(
      queryResult({
        data: undefined,
        isError: true,
        error: { error: "intent service unavailable" },
        refetch,
      }),
    )

    render(<IntentsPageContent />)

    expect(screen.getByText("intent service unavailable")).toBeTruthy()
    fireEvent.click(screen.getByRole("button", { name: "Retry" }))
    expect(refetch).toHaveBeenCalledOnce()
  })

  it("keeps direct access unauthorized", () => {
    vi.mocked(useAuthorization).mockReturnValue({
      ...authorization,
      permissions: [],
      hasPermission: () => false,
    })

    render(<IntentsPageContent />)

    expect(
      screen.getByText("You are not authorized to view intents."),
    ).toBeTruthy()
    expect(screen.queryByText("BOOKING_PADEL")).toBeNull()
  })
})
