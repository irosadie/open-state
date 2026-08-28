import { describe, expect, it } from "vitest"
import {
  createWorkflowSchema,
  getVersionStatusLabel,
  getWorkflowStatusLabel,
  publishWorkflowSchema,
  updateWorkflowSchema,
} from "./workflow"

describe("Create Workflow Schema Validation", () => {
  it("validates a valid workflow draft", () => {
    const result = createWorkflowSchema.safeParse({
      slug: "padel-booking",
      name: "Padel Booking",
      description: "booking flow",
    })

    expect(result.success).toBe(true)
  })

  it("rejects missing slug", () => {
    const result = createWorkflowSchema.safeParse({ name: "Padel Booking" })

    expect(result.success).toBe(false)
  })

  it("rejects missing name", () => {
    const result = createWorkflowSchema.safeParse({ slug: "padel-booking" })

    expect(result.success).toBe(false)
  })
})

describe("Update Workflow Schema Validation", () => {
  it("validates a valid update", () => {
    const result = updateWorkflowSchema.safeParse({
      name: "Padel Booking v2",
      version: 3,
    })

    expect(result.success).toBe(true)
  })

  it("rejects missing version", () => {
    const result = updateWorkflowSchema.safeParse({ name: "Padel Booking" })

    expect(result.success).toBe(false)
  })

  it("rejects negative version", () => {
    const result = updateWorkflowSchema.safeParse({ version: -1 })

    expect(result.success).toBe(false)
  })
})

describe("Publish Workflow Schema Validation", () => {
  it("validates a valid publish", () => {
    const result = publishWorkflowSchema.safeParse({
      version: 1,
      definition: { nodes: [] },
    })

    expect(result.success).toBe(true)
  })

  it("rejects missing definition", () => {
    const result = publishWorkflowSchema.safeParse({ version: 1 })

    expect(result.success).toBe(false)
  })
})

describe("Status label helpers", () => {
  it("maps workflow status labels", () => {
    expect(getWorkflowStatusLabel("DRAFT")).toBe("Draft")
    expect(getWorkflowStatusLabel("PUBLISHED")).toBe("Published")
    expect(getWorkflowStatusLabel("UNKNOWN" as never)).toBe("UNKNOWN")
  })

  it("maps version status labels", () => {
    expect(getVersionStatusLabel("PUBLISHED")).toBe("Published")
    expect(getVersionStatusLabel("UNKNOWN" as never)).toBe("UNKNOWN")
  })
})
