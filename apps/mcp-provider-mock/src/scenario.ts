import { dirname, resolve } from "node:path"

import { z } from "zod"

const jsonObjectSchema = z.record(z.string(), z.unknown())

const toolErrorSchema = z.object({
  code: z.string().min(1),
  message: z.string().min(1),
})

const operationSchema = z.enum([
  "availability",
  "padel_availability",
  "booking",
  "padel_booking",
  "padel_payment_create",
  "padel_payment_verify",
  "padel_notification",
  "food_cart_add",
  "food_cart_get",
  "food_order_create",
  "food_order_track",
  "food_payment_create",
  "food_payment_verify",
  "doctor_schedule",
  "doctor_reserve",
  "doctor_confirm",
  "doctor_cancel",
  "doctor_booking",
  "doctor_payment_create",
  "doctor_payment_verify",
  "doctor_notification",
  "static",
])

const toolSchema = z.object({
  name: z.string().min(1),
  description: z.string().min(1),
  inputSchema: jsonObjectSchema,
  operation: operationSchema,
  delayMs: z.number().int().nonnegative().default(0),
  result: jsonObjectSchema.optional(),
  error: toolErrorSchema.optional(),
})

const toolOverrideSchema = z.object({
  name: z.string().min(1),
  description: z.string().min(1).optional(),
  inputSchema: jsonObjectSchema.optional(),
  operation: operationSchema.optional(),
  delayMs: z.number().int().nonnegative().optional(),
  result: jsonObjectSchema.optional(),
  error: toolErrorSchema.optional(),
})

const providerSchema = z.object({
  name: z.string().min(1),
  version: z.string().min(1),
})

const slotSchema = z.object({
  venueId: z.string().min(1),
  date: z.string().min(1),
  time: z.string().min(1),
})

const bookingSchema = slotSchema.extend({
  bookingReference: z.string().min(1),
})

const scenarioDataSchema = z
  .object({
    slots: z.array(slotSchema).default([]),
    bookings: z.array(bookingSchema).default([]),
  })
  .catchall(z.unknown())
  .default({ slots: [], bookings: [] })

const scenarioSchema = z
  .object({
    provider: providerSchema,
    tools: z.array(toolSchema).min(1),
    data: scenarioDataSchema,
  })
  .superRefine((scenario, context) => {
    const names = new Set<string>()

    for (const tool of scenario.tools) {
      if (names.has(tool.name)) {
        context.addIssue({
          code: "custom",
          message: `duplicate tool name: ${tool.name}`,
          path: ["tools"],
        })
      }

      names.add(tool.name)
    }
  })

const scenarioDocumentSchema = z.object({
  extends: z.string().min(1).optional(),
  provider: providerSchema.partial().optional(),
  tools: z.array(toolOverrideSchema).optional(),
  data: z.record(z.string(), z.unknown()).optional(),
})

export type ProviderScenario = z.infer<typeof scenarioSchema>
export type ProviderToolScenario = z.infer<typeof toolSchema>
export type PadelSlot = z.infer<typeof slotSchema>
export type PadelBooking = z.infer<typeof bookingSchema>

export class ScenarioValidationError extends Error {
  constructor(message: string) {
    super(message)
    this.name = "ScenarioValidationError"
  }
}

export function parseScenario(value: unknown): ProviderScenario {
  const result = scenarioSchema.safeParse(value)

  if (result.success) {
    return result.data
  }

  throw new ScenarioValidationError(
    result.error.issues.map((issue) => issue.message).join("; "),
  )
}

function parseScenarioDocument(value: unknown, path: string) {
  const result = scenarioDocumentSchema.safeParse(value)

  if (result.success) {
    return result.data
  }

  throw new ScenarioValidationError(
    `invalid scenario document ${path}: ${result.error.issues
      .map((issue) => issue.message)
      .join("; ")}`,
  )
}

function mergeTools(
  baseTools: ProviderToolScenario[],
  overrides: Array<z.infer<typeof toolOverrideSchema>>,
) {
  const tools = new Map<string, unknown>(
    baseTools.map((tool) => [tool.name, tool]),
  )

  for (const override of overrides) {
    const baseTool = tools.get(override.name)
    tools.set(override.name, baseTool ? { ...baseTool, ...override } : override)
  }

  return [...tools.values()]
}

function mergeScenario(
  base: ProviderScenario,
  patch: z.infer<typeof scenarioDocumentSchema>,
): unknown {
  return {
    provider: { ...base.provider, ...patch.provider },
    tools: mergeTools(base.tools, patch.tools ?? []),
    data: { ...base.data, ...patch.data },
  }
}

async function loadScenarioFile(
  path: string,
  parents: string[],
): Promise<ProviderScenario> {
  const resolvedPath = resolve(path)

  if (parents.includes(resolvedPath)) {
    throw new ScenarioValidationError(
      `scenario inheritance cycle detected at: ${resolvedPath}`,
    )
  }

  let raw: string

  try {
    raw = await Bun.file(resolvedPath).text()
  } catch {
    throw new ScenarioValidationError(`unable to read scenario: ${path}`)
  }

  let value: unknown

  try {
    value = JSON.parse(raw)
  } catch {
    throw new ScenarioValidationError(`invalid JSON scenario: ${path}`)
  }

  const document = parseScenarioDocument(value, resolvedPath)
  if (!document.extends) {
    return parseScenario(value)
  }

  const parent = await loadScenarioFile(
    resolve(dirname(resolvedPath), document.extends),
    [...parents, resolvedPath],
  )

  return parseScenario(mergeScenario(parent, document))
}

export async function loadScenario(path: string): Promise<ProviderScenario> {
  return loadScenarioFile(path, [])
}
