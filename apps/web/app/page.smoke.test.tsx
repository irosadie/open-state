import { render, screen } from "@testing-library/react"
import { describe, expect, it } from "vitest"
import HomePage from "./page"

describe("HomePage smoke", () => {
  it("renders the OpenState landing page", () => {
    render(<HomePage />)

    expect(
      screen.getByText("Enterprise conversation state orchestration"),
    ).toBeTruthy()
    expect(screen.getByText("OpenState web console siap dipakai")).toBeTruthy()
    expect(
      screen.getByText("Semua route utama OpenState siap digunakan."),
    ).toBeTruthy()
  })
})
