import { z } from "zod"

// Provider type of a capability (PRD §59).
export const providerTypes = ["MCP", "INTERNAL", "HTTP", "FUTURE"] as const

export const providerTypeLabels = [
  { label: "MCP", value: "MCP" },
  { label: "Internal", value: "INTERNAL" },
  { label: "HTTP", value: "HTTP" },
  { label: "Future", value: "FUTURE" },
]

export const getProviderTypeLabel = (value: (typeof providerTypes)[number]) => {
  return providerTypeLabels.find((l) => l.value === value)?.label ?? value
}

// Lifecycle of a capability in the registry.
export const capabilityStatuses = ["ACTIVE", "INACTIVE", "DISABLED"] as const

export const capabilityStatusLabels = [
  { label: "Active", value: "ACTIVE" },
  { label: "Inactive", value: "INACTIVE" },
  { label: "Disabled", value: "DISABLED" },
]

export const getCapabilityStatusLabel = (
  value: (typeof capabilityStatuses)[number],
) => {
  return capabilityStatusLabels.find((l) => l.value === value)?.label ?? value
}

// Scope type of a capability binding (PRD §60).
export const bindingScopeTypes = ["TENANT", "WORKFLOW", "STATE"] as const

export const bindingScopeTypeLabels = [
  { label: "Tenant", value: "TENANT" },
  { label: "Workflow", value: "WORKFLOW" },
  { label: "State", value: "STATE" },
]

export const getBindingScopeTypeLabel = (
  value: (typeof bindingScopeTypes)[number],
) => {
  return bindingScopeTypeLabels.find((l) => l.value === value)?.label ?? value
}

// Permission of a capability binding.
export const bindingPermissions = ["ALLOW", "DENY"] as const

export const bindingPermissionLabels = [
  { label: "Allow", value: "ALLOW" },
  { label: "Deny", value: "DENY" },
]

export const getBindingPermissionLabel = (
  value: (typeof bindingPermissions)[number],
) => {
  return bindingPermissionLabels.find((l) => l.value === value)?.label ?? value
}

// Free-form JSON object (input/output schema, payload). Optional values default
// to an empty object.
const jsonSchemaField = z.record(z.string(), z.unknown()).optional()

// Register a capability (PRD §59).
export const createCapabilitySchema = z.object({
  name: z.string().min(1, "Name is required"),
  description: z.string().optional(),
  providerType: z.enum(providerTypes, {
    errorMap: () => ({ message: "Provider type is required" }),
  }),
  providerId: z.string().optional(),
  providerTool: z.string().optional(),
  inputSchema: jsonSchemaField,
  outputSchema: jsonSchemaField,
  version: z.number().int().positive().optional(),
  credentialReference: z.string().optional(),
})

export type CreateCapabilitySchemaProps = z.infer<typeof createCapabilitySchema>

// Update a capability's mutable fields (partial/PATCH).
export const updateCapabilitySchema = z.object({
  description: z.string().optional(),
  providerType: z.enum(providerTypes).optional(),
  providerId: z.string().optional(),
  providerTool: z.string().optional(),
  inputSchema: jsonSchemaField,
  outputSchema: jsonSchemaField,
  status: z.enum(capabilityStatuses).optional(),
  version: z.number().int().positive().optional(),
  credentialReference: z.string().optional(),
})

export type UpdateCapabilitySchemaProps = z.infer<typeof updateCapabilitySchema>

// Bind a capability to a tenant/workflow/state scope (PRD §60).
export const bindingSchema = z.object({
  scopeType: z.enum(bindingScopeTypes, {
    errorMap: () => ({ message: "Scope type is required" }),
  }),
  scopeId: z.string().min(1, "Scope id is required"),
  permission: z.enum(bindingPermissions).optional(),
})

export type BindingSchemaProps = z.infer<typeof bindingSchema>

// Test-invoke a capability in sandbox/mock mode (PRD §2064).
export const testInvocationSchema = z.object({
  payload: z.record(z.string(), z.unknown()).default({}),
  scopeId: z.string().optional(),
})

export type TestInvocationSchemaProps = z.infer<typeof testInvocationSchema>
