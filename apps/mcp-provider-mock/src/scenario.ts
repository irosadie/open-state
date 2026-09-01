import { z } from "zod"

const jsonObjectSchema = z.record(z.string(), z.unknown())

const toolErrorSchema = z.object({
  code: z.string().min(1),
  message: z.string().min(1),
})

const toolSchema = z.object({
  name: z.string().min(1),
  description: z.string().min(1),
  inputSchema: jsonObjectSchema,
  operation: z.enum([
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
  ]),
  delayMs: z.number().int().nonnegative().default(0),
  result: jsonObjectSchema.optional(),
  error: toolErrorSchema.optional(),
})

const slotSchema = z.object({
  venueId: z.string().min(1),
  date: z.string().min(1),
  time: z.string().min(1),
})

const bookingSchema = slotSchema.extend({
  bookingReference: z.string().min(1),
})

const scenarioSchema = z
  .object({
    provider: z.object({
      name: z.string().min(1),
      version: z.string().min(1),
    }),
    tools: z.array(toolSchema).min(1),
    data: z
      .object({
        slots: z.array(slotSchema).default([]),
        bookings: z.array(bookingSchema).default([]),
      })
      .catchall(z.unknown())
      .default({ slots: [], bookings: [] }),
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

export async function loadScenario(path: string): Promise<ProviderScenario> {
  let raw: string

  try {
    raw = await Bun.file(path).text()
  } catch {
    throw new ScenarioValidationError(`unable to read scenario: ${path}`)
  }

  try {
    return parseScenario(JSON.parse(raw))
  } catch (error) {
    if (error instanceof ScenarioValidationError) {
      throw error
    }

    throw new ScenarioValidationError(`invalid JSON scenario: ${path}`)
  }
}
