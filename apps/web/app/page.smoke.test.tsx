import { render, screen } from "@testing-library/react"
import { describe, expect, it } from "vitest"
import HomePage from "./page"

describe("HomePage smoke", () => {
  it("renders the starter landing page", () => {
    render(<HomePage />)

    expect(
      screen.getByText("Starter kosong untuk mulai vibe coding dari nol"),
    ).toBeTruthy()
    expect(screen.getByText("Homepage starter sudah siap dipakai")).toBeTruthy()
    expect(
      screen.getByText("Tidak ada route demo yang dikunci di starter ini."),
    ).toBeTruthy()
  })
})
