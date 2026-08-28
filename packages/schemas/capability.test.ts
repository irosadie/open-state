import { describe, expect, it } from "vitest"
import {
  bindingSchema,
  createCapabilitySchema,
  getBindingPermissionLabel,
  getBindingScopeTypeLabel,
  getCapabilityStatusLabel,
  getProviderTypeLabel,
  testInvocationSchema,
  updateCapabilitySchema,
} from "./capability"

describe("Create Capability Schema Validation", () => {
  it("validates a valid capability", () => {
    const result = createCapabilitySchema.safeParse({
      name: "payment.create",
      providerType: "MCP",
      version: 1,
    })

    expect(result.success).toBe(true)
  })

  it("rejects missing name", () => {
    const result = createCapabilitySchema.safeParse({ providerType: "MCP" })

    expect(result.success).toBe(false)
  })

  it("rejects invalid providerType", () => {
    const result = createCapabilitySchema.safeParse({
      name: "payment.create",
      providerType: "NOT_A_TYPE",
    })

    expect(result.success).toBe(false)
  })

  it("rejects missing providerType", () => {
    const result = createCapabilitySchema.safeParse({ name: "payment.create" })

    expect(result.success).toBe(false)
  })

  it("accepts schema objects and credential reference", () => {
    const result = createCapabilitySchema.safeParse({
      name: "payment.create",
      providerType: "HTTP",
      inputSchema: { type: "object" },
      outputSchema: { type: "object" },
      credentialReference: "cred-ref-1",
    })

    expect(result.success).toBe(true)
  })
})

describe("Update Capability Schema Validation", () => {
  it("validates a partial update", () => {
    const result = updateCapabilitySchema.safeParse({ status: "ACTIVE" })

    expect(result.success).toBe(true)
  })

  it("rejects invalid status", () => {
    const result = updateCapabilitySchema.safeParse({ status: "BAD" })

    expect(result.success).toBe(false)
  })

  it("rejects invalid providerType", () => {
    const result = updateCapabilitySchema.safeParse({ providerType: "NOPE" })

    expect(result.success).toBe(false)
  })
})

describe("Binding Schema Validation", () => {
  it("validates a valid binding", () => {
    const result = bindingSchema.safeParse({
      scopeType: "WORKFLOW",
      scopeId: "wf-1",
    })

    expect(result.success).toBe(true)
  })

  it("rejects missing scopeId", () => {
    const result = bindingSchema.safeParse({ scopeType: "STATE", scopeId: "" })

    expect(result.success).toBe(false)
  })

  it("rejects invalid scopeType", () => {
    const result = bindingSchema.safeParse({
      scopeType: "GLOBAL",
      scopeId: "x",
    })

    expect(result.success).toBe(false)
  })

  it("rejects invalid permission", () => {
    const result = bindingSchema.safeParse({
      scopeType: "TENANT",
      scopeId: "t",
      permission: "MAYBE",
    })

    expect(result.success).toBe(false)
  })
})

describe("Test Invocation Schema Validation", () => {
  it("validates a payload and optional scopeId", () => {
    const result = testInvocationSchema.safeParse({
      payload: { amount: 100 },
      scopeId: "wf-1",
    })

    expect(result.success).toBe(true)
  })

  it("defaults payload to empty object", () => {
    const result = testInvocationSchema.safeParse({})

    expect(result.success).toBe(true)

    if (result.success) {
      expect(result.data.payload).toEqual({})
    }
  })
})

describe("Enum label helpers", () => {
  it("maps provider type labels", () => {
    expect(getProviderTypeLabel("MCP")).toBe("MCP")
    expect(getProviderTypeLabel("INTERNAL")).toBe("Internal")
  })

  it("maps status labels", () => {
    expect(getCapabilityStatusLabel("ACTIVE")).toBe("Active")
  })

  it("maps scope type labels", () => {
    expect(getBindingScopeTypeLabel("WORKFLOW")).toBe("Workflow")
  })

  it("maps permission labels", () => {
    expect(getBindingPermissionLabel("ALLOW")).toBe("Allow")
  })
})
