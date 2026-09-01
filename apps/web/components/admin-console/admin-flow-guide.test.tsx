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
  it("explains the tenant to state path and makes Project navigable", () => {
    render(<AdminFlowGuide currentStep="workflow" />)

    expect(screen.getByText("Tenant")).toBeTruthy()
    expect(screen.getByText("Project")).toBeTruthy()
    expect(screen.getByText("Intent")).toBeTruthy()
    expect(screen.getByText("Workflow")).toBeTruthy()
    expect(screen.getByText("State")).toBeTruthy()
    expect(screen.getAllByText(/Default Project/).length).toBeGreaterThan(0)
    expect(
      screen.getByRole("link", { name: /Tenant/ }).getAttribute("href"),
    ).toBe("/admin/tenant")
    expect(
      screen.getByRole("link", { name: /State/ }).getAttribute("href"),
    ).toBe("/state-builder")
    expect(
      screen.getByRole("link", { name: /Project/ }).getAttribute("href"),
    ).toBe("/admin/projects")
    expect(
      screen.getByRole("link", { name: /Intent/ }).getAttribute("href"),
    ).toBe("/admin/intents")
  })

  it("carries a selected project into downstream steps", () => {
    render(
      <AdminFlowGuide
        currentStep="project"
        projectId="project-1"
        projectName="Padel"
      />,
    )

    expect(screen.getByText(/Padel/)).toBeTruthy()
    expect(
      screen.getByRole("link", { name: /Intent/ }).getAttribute("href"),
    ).toBe("/admin/intents?projectId=project-1")
    expect(
      screen.getByRole("link", { name: /Workflow/ }).getAttribute("href"),
    ).toBe("/admin/workflows?projectId=project-1")
  })
})
