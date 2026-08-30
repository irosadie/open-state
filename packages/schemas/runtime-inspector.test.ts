import { describe, expect, it } from "vitest"
import {
  runtimeDebugTraceQuerySchema,
  runtimeInstanceQuerySchema,
} from "./runtime-inspector"

describe("runtime inspector schemas", () => {
  it("coerces valid pagination query values", () => {
    expect(
      runtimeInstanceQuerySchema.parse({
        status: "RUNNING",
        page: "2",
        pageSize: "50",
      }),
    ).toMatchObject({ status: "RUNNING", page: 2, pageSize: 50 })
  })

  it("rejects invalid filters and oversized pages", () => {
    expect(
      runtimeInstanceQuerySchema.safeParse({ status: "UNKNOWN" }).success,
    ).toBe(false)
    expect(
      runtimeInstanceQuerySchema.safeParse({ pageSize: 101 }).success,
    ).toBe(false)
    expect(runtimeDebugTraceQuerySchema.safeParse({ turnId: "" }).success).toBe(
      false,
    )
  })
})
