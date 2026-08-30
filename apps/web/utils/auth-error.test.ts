import { describe, expect, it } from "vitest"
import { getApiErrorStatus, getAuthRecoveryAction } from "./auth-error"

describe("authentication error recovery", () => {
  it("classifies status codes without treating forbidden as expired", () => {
    expect(getApiErrorStatus({ status: 401 })).toBe(401)
    expect(getAuthRecoveryAction(401)).toBe("login")
    expect(getAuthRecoveryAction(403)).toBe("forbidden")
    expect(getAuthRecoveryAction(500)).toBe("none")
  })

  it("ignores malformed error values", () => {
    expect(getApiErrorStatus(null)).toBeUndefined()
    expect(getApiErrorStatus({ status: "403" })).toBeUndefined()
  })
})
