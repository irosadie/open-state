import { render, screen } from "@testing-library/react"
import type { ReactNode } from "react"
import { describe, expect, it, vi } from "vitest"

import { AdminFlowGuide } from "./admin-flow-guide"

vi.mock("next/link", () => ({
  default: ({ children, href }: { children: ReactNode; href: string }) => (
    <a href={href}>{children}</a>
  ),
}))

describe("AdminFlowGuide", () => {
  it("explains the tenant to builder path and default project scope", () => {
    render(<AdminFlowGuide currentStep="workflow" />)

    expect(screen.getByText("Tenant")).toBeTruthy()
    expect(screen.getByText("Project")).toBeTruthy()
    expect(screen.getByText("Workflow")).toBeTruthy()
    expect(screen.getByText("Builder")).toBeTruthy()
    expect(screen.getAllByText(/Default Project/).length).toBeGreaterThan(0)
    expect(
      screen.getByText(/Project settings and switching are not available yet/),
    ).toBeTruthy()
    expect(
      screen.getByRole("link", { name: /Tenant/ }).getAttribute("href"),
    ).toBe("/admin/tenant")
    expect(
      screen.getByRole("link", { name: /Builder/ }).getAttribute("href"),
    ).toBe("/state-builder")
  })
})
