import {
  type JsonSchemaType,
  McpServer,
  WebStandardStreamableHTTPServerTransport,
  fromJsonSchema,
} from "@modelcontextprotocol/server"

import type { ProviderScenario, ProviderToolScenario } from "./scenario"
import { PadelScenarioStore, type ToolData, isStoreError } from "./store"

type ToolOutput = Record<string, unknown>

export type ProviderMock = {
  handleMcpRequest: (request: Request) => Promise<Response>
  providerName: string
}

function textResult(data: ToolOutput) {
  return {
    content: [{ type: "text" as const, text: JSON.stringify(data) }],
    structuredContent: data,
  }
}

function errorResult(code: string, message: string) {
  return {
    content: [
      { type: "text" as const, text: JSON.stringify({ code, message }) },
    ],
    isError: true,
  }
}

function readString(
  args: Record<string, unknown>,
  name: string,
): string | null {
  const value = args[name]
  return typeof value === "string" && value.length > 0 ? value : null
}

function readNumber(
  args: Record<string, unknown>,
  name: string,
): number | null {
  const value = args[name]
  return typeof value === "number" ? value : null
}

async function executeTool(
  tool: ProviderToolScenario,
  store: PadelScenarioStore,
  args: Record<string, unknown>,
): Promise<ReturnType<typeof errorResult> | ReturnType<typeof textResult>> {
  if (tool.delayMs > 0) {
    await Bun.sleep(tool.delayMs)
  }

  if (tool.error) {
    return errorResult(tool.error.code, tool.error.message)
  }

  if (tool.operation === "availability") {
    const venueId = readString(args, "venue_id")
    const date = readString(args, "date")
    if (!venueId || !date) {
      return errorResult("invalid_input", "venue_id and date are required")
    }
    return textResult(store.checkAvailability(venueId, date))
  }

  if (tool.operation === "padel_availability") {
    const venueId = readString(args, "venue_id")
    const date = readString(args, "date")
    if (!venueId || !date) {
      return errorResult("invalid_input", "venue_id and date are required")
    }
    return textResult(store.padelAvailability(venueId, date))
  }

  if (tool.operation === "booking" || tool.operation === "padel_booking") {
    const venueId = readString(args, "venue_id")
    const date = readString(args, "date")
    const time = readString(args, "time")
    if (!venueId || !date || !time) {
      return errorResult(
        "invalid_input",
        "venue_id, date, and time are required",
      )
    }
    const result =
      tool.operation === "padel_booking"
        ? store.createPadelCourtBooking(venueId, date, time)
        : store.createBooking(venueId, date, time)
    return isStoreError(result)
      ? errorResult(result.code, result.message)
      : textResult(result)
  }

  if (tool.operation === "padel_payment_create") {
    const bookingId = readString(args, "booking_id")
    return bookingId
      ? resultText(store.createPadelPayment(bookingId))
      : errorResult("invalid_input", "booking_id is required")
  }

  if (tool.operation === "padel_payment_verify") {
    const paymentId = readString(args, "payment_id")
    return paymentId
      ? resultText(store.verifyPadelPayment(paymentId))
      : errorResult("invalid_input", "payment_id is required")
  }

  if (tool.operation === "padel_notification") {
    const bookingId = readString(args, "booking_id")
    const message = readString(args, "message")
    return bookingId && message
      ? resultText(store.sendPadelNotification(bookingId, message))
      : errorResult("invalid_input", "booking_id and message are required")
  }

  if (tool.operation === "food_cart_add") {
    const menuId = readString(args, "menu_id")
    const quantity = readNumber(args, "quantity")
    if (!menuId || quantity === null) {
      return errorResult("invalid_input", "menu_id and quantity are required")
    }
    return resultText(
      store.addFoodCart(readString(args, "cart_id"), menuId, quantity),
    )
  }

  if (tool.operation === "food_cart_get") {
    const cartId = readString(args, "cart_id")
    return cartId
      ? resultText(store.getFoodCart(cartId))
      : errorResult("invalid_input", "cart_id is required")
  }

  if (tool.operation === "food_order_create") {
    const cartId = readString(args, "cart_id")
    const deliveryAddress = readString(args, "delivery_address")
    return cartId && deliveryAddress
      ? resultText(store.createFoodOrder(cartId, deliveryAddress))
      : errorResult(
          "invalid_input",
          "cart_id and delivery_address are required",
        )
  }

  if (tool.operation === "food_order_track") {
    const orderId = readString(args, "order_id")
    return orderId
      ? resultText(store.trackFoodOrder(orderId))
      : errorResult("invalid_input", "order_id is required")
  }

  if (tool.operation === "food_payment_create") {
    const orderId = readString(args, "order_id")
    return orderId
      ? resultText(store.createFoodPayment(orderId))
      : errorResult("invalid_input", "order_id is required")
  }

  if (tool.operation === "food_payment_verify") {
    const paymentId = readString(args, "payment_id")
    return paymentId
      ? resultText(store.verifyFoodPayment(paymentId))
      : errorResult("invalid_input", "payment_id is required")
  }

  if (tool.operation === "doctor_schedule") {
    return textResult(store.doctorSchedule())
  }

  if (tool.operation === "doctor_reserve") {
    const scheduleId = readString(args, "schedule_id")
    return scheduleId
      ? resultText(store.reserveDoctorAppointment(scheduleId))
      : errorResult("invalid_input", "schedule_id is required")
  }

  if (tool.operation === "doctor_confirm") {
    const reservationId = readString(args, "reservation_id")
    return reservationId
      ? resultText(store.confirmDoctorReservation(reservationId))
      : errorResult("invalid_input", "reservation_id is required")
  }

  if (tool.operation === "doctor_cancel") {
    const bookingId = readString(args, "booking_id")
    return bookingId
      ? resultText(store.cancelDoctorAppointment(bookingId))
      : errorResult("invalid_input", "booking_id is required")
  }

  if (tool.operation === "doctor_booking") {
    const scheduleId = readString(args, "schedule_id")
    return scheduleId
      ? resultText(store.bookDoctorAppointment(scheduleId))
      : errorResult("invalid_input", "schedule_id is required")
  }

  if (tool.operation === "doctor_payment_create") {
    const bookingId = readString(args, "booking_id")
    return bookingId
      ? resultText(store.createDoctorPayment(bookingId))
      : errorResult("invalid_input", "booking_id is required")
  }

  if (tool.operation === "doctor_payment_verify") {
    const paymentId = readString(args, "payment_id")
    return paymentId
      ? resultText(store.verifyDoctorPayment(paymentId))
      : errorResult("invalid_input", "payment_id is required")
  }

  if (tool.operation === "doctor_notification") {
    const bookingId = readString(args, "booking_id")
    const message = readString(args, "message")
    return bookingId && message
      ? resultText(store.sendDoctorNotification(bookingId, message))
      : errorResult("invalid_input", "booking_id and message are required")
  }

  return textResult(tool.result ?? {})
}

function resultText(result: ToolData | { code: string; message: string }) {
  return isStoreError(result)
    ? errorResult(result.code, result.message)
    : textResult(result)
}

function inputSchemaFor(tool: ProviderToolScenario): Record<string, unknown> {
  // Static fixture tools intentionally return a canned catalog response. Keep
  // them permissive so callers can pass the search context they collected
  // without the mock rejecting it before the scenario handler runs.
  const properties = tool.inputSchema.properties
  const required = tool.inputSchema.required
  const hasNoProperties =
    properties === undefined ||
    (typeof properties === "object" &&
      properties !== null &&
      !Array.isArray(properties) &&
      Object.keys(properties).length === 0)
  const hasNoRequiredFields =
    required === undefined || (Array.isArray(required) && required.length === 0)

  if (
    tool.operation === "static" &&
    tool.inputSchema.type === "object" &&
    tool.inputSchema.additionalProperties === false &&
    hasNoProperties &&
    hasNoRequiredFields
  ) {
    return { ...tool.inputSchema, additionalProperties: true }
  }

  return tool.inputSchema
}

export async function createProviderMock(
  scenario: ProviderScenario,
): Promise<ProviderMock> {
  const store = new PadelScenarioStore(scenario)
  const server = new McpServer({
    name: scenario.provider.name,
    version: scenario.provider.version,
  })

  for (const tool of scenario.tools) {
    server.registerTool(
      tool.name,
      {
        description: tool.description,
        inputSchema: fromJsonSchema<Record<string, unknown>>(
          inputSchemaFor(tool) as JsonSchemaType,
        ),
      },
      async (args) => executeTool(tool, store, args),
    )
  }

  const transport = new WebStandardStreamableHTTPServerTransport({
    enableJsonResponse: true,
    sessionIdGenerator: undefined,
  })
  await server.connect(transport)

  return {
    handleMcpRequest: (request) => transport.handleRequest(request),
    providerName: scenario.provider.name,
  }
}
