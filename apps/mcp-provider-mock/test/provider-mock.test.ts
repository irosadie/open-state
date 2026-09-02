import { describe, expect, test } from "bun:test"
import {
  Client,
  StreamableHTTPClientTransport,
} from "@modelcontextprotocol/client"

import { type ProviderMockApp, startProviderMock } from "../src/app"

const fixturePath = (name: string) =>
  new URL(`../fixtures/${name}`, import.meta.url).pathname

async function withApp<T>(
  fixture: string,
  run: (app: ProviderMockApp) => Promise<T>,
): Promise<T> {
  const app = await startProviderMock({
    port: 0,
    scenarioPath: fixturePath(fixture),
  })

  try {
    return await run(app)
  } finally {
    app.stop()
  }
}

async function withClient<T>(
  app: ProviderMockApp,
  run: (client: Client) => Promise<T>,
): Promise<T> {
  const client = new Client({ name: "provider-mock-test", version: "1.0.0" })
  const transport = new StreamableHTTPClientTransport(new URL(`${app.url}/mcp`))
  await client.connect(transport)

  try {
    return await run(client)
  } finally {
    await client.close()
  }
}

function structuredData(value: unknown): Record<string, unknown> {
  if (typeof value !== "object" || value === null || Array.isArray(value)) {
    throw new Error("tool result did not include structured object data")
  }

  return value as Record<string, unknown>
}

function arrayData(value: unknown, field: string): unknown[] {
  const item = structuredData(value)[field]

  if (!Array.isArray(item)) {
    throw new Error(`tool result field ${field} was not an array`)
  }

  return item
}

function stringData(value: unknown, field: string): string {
  const item = structuredData(value)[field]

  if (typeof item !== "string") {
    throw new Error(`tool result field ${field} was not a string`)
  }

  return item
}

describe("MCP provider mock", () => {
  test("does not become ready when its scenario cannot be loaded", async () => {
    await expect(
      startProviderMock({
        port: 0,
        scenarioPath: fixturePath("missing.json"),
      }),
    ).rejects.toThrow("unable to read scenario")
  })

  test("reports readiness and discovers only configured tools", async () => {
    await withApp("padel.json", async (app) => {
      const ready = await fetch(`${app.url}/health/ready`)
      expect(ready.ok).toBe(true)
      expect(await ready.json()).toEqual({
        provider: "padel-provider-mock",
        status: "ready",
      })

      await withClient(app, async (client) => {
        const { tools } = await client.listTools()
        expect(tools.map((tool) => tool.name)).toEqual([
          "padel.cek_available",
          "padel.create_booking",
          "location.list",
          "padel.court.search",
          "padel.court.availability",
          "padel.court.book",
          "padel.payment.create",
          "padel.payment.verify",
          "padel.notification.send",
        ])

        const search = await client.callTool({
          arguments: {},
          name: "padel.court.search",
        })
        const courts = arrayData(search.structuredContent, "courts")
        expect(courts[0]).toMatchObject({ name: "GOR Senayan Court A" })
      })
    })
  })

  test("isolates food-order tools and returns English fixture data", async () => {
    await withApp("food-order.json", async (app) => {
      await withClient(app, async (client) => {
        const { tools } = await client.listTools()
        expect(tools.map((tool) => tool.name)).toEqual([
          "food.menu.list",
          "food.cart.add",
          "food.cart.get",
          "food.order.create",
          "food.payment.create",
          "food.payment.verify",
          "food.order.track",
        ])

        const menu = await client.callTool({
          arguments: {},
          name: "food.menu.list",
        })
        const items = arrayData(menu.structuredContent, "menu")
        expect(items[0]).toMatchObject({
          description: "Fried rice with egg, chicken, and crackers",
        })
      })
    })
  })

  test("isolates doctor tools and returns English fixture data", async () => {
    await withApp("doctor.json", async (app) => {
      await withClient(app, async (client) => {
        const { tools } = await client.listTools()
        const names = tools.map((tool) => tool.name)
        expect(names).toContain("doctor.search")
        expect(names).toContain("doctor.notification.send")
        expect(names).not.toContain("food.menu.list")
        expect(names).not.toContain("padel.court.search")

        const doctor = await client.callTool({
          arguments: {},
          name: "doctor.search",
        })
        const doctors = arrayData(doctor.structuredContent, "doctors")
        expect(doctors[0]).toMatchObject({
          specialization: "Internal Medicine",
        })

        const contextualLookup = await client.callTool({
          arguments: { specialty: "Cardiology", location: "Jakarta" },
          name: "doctor.lookup",
        })
        expect(contextualLookup.isError).not.toBe(true)
        expect(
          structuredData(contextualLookup.structuredContent),
        ).toMatchObject({
          doctor: { id: "doc-001" },
        })
      })
    })
  })

  test("reads availability, books a slot, and rejects a duplicate booking", async () => {
    await withApp("padel.json", async (app) => {
      await withClient(app, async (client) => {
        const availability = await client.callTool({
          arguments: { date: "2026-09-01", venue_id: "padel-senayan" },
          name: "padel.cek_available",
        })
        expect(structuredData(availability.structuredContent)).toMatchObject({
          availableSlots: ["18:00", "19:00"],
        })

        const booking = await client.callTool({
          arguments: {
            date: "2026-09-01",
            time: "18:00",
            venue_id: "padel-senayan",
          },
          name: "padel.create_booking",
        })
        expect(structuredData(booking.structuredContent)).toMatchObject({
          bookingReference: "PAD-0001",
        })

        const updatedAvailability = await client.callTool({
          arguments: { date: "2026-09-01", venue_id: "padel-senayan" },
          name: "padel.cek_available",
        })
        expect(
          structuredData(updatedAvailability.structuredContent),
        ).toMatchObject({ availableSlots: ["19:00"] })

        const duplicate = await client.callTool({
          arguments: {
            date: "2026-09-01",
            time: "18:00",
            venue_id: "padel-senayan",
          },
          name: "padel.create_booking",
        })
        expect(duplicate.isError).toBe(true)

        const payment = await client.callTool({
          arguments: { booking_id: "PAD-0001" },
          name: "padel.payment.create",
        })
        const paymentId = stringData(payment.structuredContent, "payment_id")
        expect(paymentId).toBe("PAY-PADEL-0001")

        const verified = await client.callTool({
          arguments: { payment_id: paymentId },
          name: "padel.payment.verify",
        })
        expect(structuredData(verified.structuredContent)).toMatchObject({
          status: "PAID",
        })

        const notification = await client.callTool({
          arguments: {
            booking_id: "PAD-0001",
            message: "Your padel booking is confirmed.",
          },
          name: "padel.notification.send",
        })
        expect(structuredData(notification.structuredContent)).toMatchObject({
          status: "DELIVERED",
        })
      })
    })
  })

  test("writes and reads a food order lifecycle", async () => {
    await withApp("food-order.json", async (app) => {
      await withClient(app, async (client) => {
        const added = await client.callTool({
          arguments: { menu_id: "menu-001", quantity: 2 },
          name: "food.cart.add",
        })
        const cartId = stringData(added.structuredContent, "cart_id")
        expect(cartId).toBe("CART-0001")

        await client.callTool({
          arguments: { cart_id: cartId, menu_id: "menu-004", quantity: 1 },
          name: "food.cart.add",
        })
        const cart = await client.callTool({
          arguments: { cart_id: cartId },
          name: "food.cart.get",
        })
        expect(structuredData(cart.structuredContent)).toMatchObject({
          total: 78000,
          item_count: 3,
        })

        const order = await client.callTool({
          arguments: {
            cart_id: cartId,
            delivery_address: "Sudirman Street No. 1, Jakarta",
          },
          name: "food.order.create",
        })
        const orderId = stringData(order.structuredContent, "order_id")
        expect(orderId).toBe("ORD-0001")

        const payment = await client.callTool({
          arguments: { order_id: orderId },
          name: "food.payment.create",
        })
        const paymentId = stringData(payment.structuredContent, "payment_id")
        await client.callTool({
          arguments: { payment_id: paymentId },
          name: "food.payment.verify",
        })

        const tracked = await client.callTool({
          arguments: { order_id: orderId },
          name: "food.order.track",
        })
        expect(structuredData(tracked.structuredContent)).toMatchObject({
          status: "ON_DELIVERY",
          payment_status: "PAID",
        })

        const emptyOrder = await client.callTool({
          arguments: {
            cart_id: "CART-EMPTY",
            delivery_address: "Sudirman Street No. 1, Jakarta",
          },
          name: "food.order.create",
        })
        expect(emptyOrder.isError).toBe(true)
      })
    })
  })

  test("writes and reads a doctor appointment lifecycle", async () => {
    await withApp("doctor.json", async (app) => {
      await withClient(app, async (client) => {
        const reservation = await client.callTool({
          arguments: { schedule_id: "sch-001" },
          name: "booking.reserve",
        })
        const reservationId = stringData(
          reservation.structuredContent,
          "reservation_id",
        )
        expect(reservationId).toBe("RES-0001")

        const confirmed = await client.callTool({
          arguments: { reservation_id: reservationId },
          name: "booking.confirm",
        })
        const bookingId = stringData(confirmed.structuredContent, "booking_id")
        expect(bookingId).toBe("BKGD-0001")

        const payment = await client.callTool({
          arguments: { booking_id: bookingId },
          name: "doctor.payment.create",
        })
        const paymentId = stringData(payment.structuredContent, "payment_id")
        await client.callTool({
          arguments: { payment_id: paymentId },
          name: "doctor.payment.verify",
        })

        const notification = await client.callTool({
          arguments: {
            booking_id: bookingId,
            message: "Your doctor appointment is confirmed.",
          },
          name: "doctor.notification.send",
        })
        expect(structuredData(notification.structuredContent)).toMatchObject({
          status: "DELIVERED",
        })

        const scheduleAfterBooking = await client.callTool({
          arguments: {},
          name: "doctor.schedule",
        })
        const schedulesAfterBooking = arrayData(
          scheduleAfterBooking.structuredContent,
          "schedules",
        )
        expect(schedulesAfterBooking[0]).toMatchObject({ available: false })

        await client.callTool({
          arguments: { booking_id: bookingId },
          name: "booking.cancel",
        })
        const scheduleAfterCancel = await client.callTool({
          arguments: {},
          name: "doctor.schedule",
        })
        const schedulesAfterCancel = arrayData(
          scheduleAfterCancel.structuredContent,
          "schedules",
        )
        expect(schedulesAfterCancel[0]).toMatchObject({ available: true })

        const directBooking = await client.callTool({
          arguments: { schedule_id: "sch-002" },
          name: "doctor.book",
        })
        expect(structuredData(directBooking.structuredContent)).toMatchObject({
          status: "CONFIRMED",
        })
        const duplicate = await client.callTool({
          arguments: { schedule_id: "sch-002" },
          name: "doctor.book",
        })
        expect(duplicate.isError).toBe(true)
      })
    })
  })

  test("runs the complete doctor happy path with dynamic identifiers", async () => {
    await withApp("doctor-happy.json", async (app) => {
      await withClient(app, async (client) => {
        const { tools } = await client.listTools()
        expect(tools.map((tool) => tool.name)).toContain("schedule.check")
        expect(tools.map((tool) => tool.name)).toContain("booking.reserve")

        const lookup = await client.callTool({
          arguments: {},
          name: "doctor.lookup",
        })
        expect(structuredData(lookup.structuredContent)).toMatchObject({
          doctor: { id: "doc-001", specialization: "Internal Medicine" },
        })

        const catalog = await client.callTool({
          arguments: {},
          name: "catalog.doctor_list",
        })
        expect(arrayData(catalog.structuredContent, "doctors")).toHaveLength(5)

        const availability = await client.callTool({
          arguments: {},
          name: "schedule.check",
        })
        const availabilityData = structuredData(availability.structuredContent)
        expect(availabilityData).toMatchObject({
          schedule_available: true,
          schedule_id: "sch-happy-001",
        })
        const scheduleId = stringData(
          availability.structuredContent,
          "schedule_id",
        )

        const queue = await client.callTool({
          arguments: {},
          name: "queue.check",
        })
        expect(structuredData(queue.structuredContent)).toMatchObject({
          queue_available: true,
          schedule_id: scheduleId,
        })

        const reservation = await client.callTool({
          arguments: { schedule_id: scheduleId },
          name: "booking.reserve",
        })
        const reservationId = stringData(
          reservation.structuredContent,
          "reservation_id",
        )

        const conflict = await client.callTool({
          arguments: { schedule_id: scheduleId },
          name: "booking.reserve",
        })
        expect(conflict.isError).toBe(true)
        expect(conflict.content[0]).toMatchObject({
          text: expect.stringContaining("schedule_conflict"),
        })

        const confirmed = await client.callTool({
          arguments: { reservation_id: reservationId },
          name: "booking.confirm",
        })
        const bookingId = stringData(confirmed.structuredContent, "booking_id")
        expect(structuredData(confirmed.structuredContent)).toMatchObject({
          status: "CONFIRMED",
          payment_status: "PENDING",
        })

        const payment = await client.callTool({
          arguments: { booking_id: bookingId },
          name: "doctor.payment.create",
        })
        const paymentId = stringData(payment.structuredContent, "payment_id")
        expect(paymentId).toBe("PAY-DOC-0001")

        const duplicatePayment = await client.callTool({
          arguments: { booking_id: bookingId },
          name: "doctor.payment.create",
        })
        expect(duplicatePayment.isError).toBe(true)

        const verified = await client.callTool({
          arguments: { payment_id: paymentId },
          name: "doctor.payment.verify",
        })
        expect(structuredData(verified.structuredContent)).toMatchObject({
          status: "PAID",
          booking_id: bookingId,
        })

        const notification = await client.callTool({
          arguments: {
            booking_id: bookingId,
            message: "Your doctor appointment is confirmed.",
          },
          name: "doctor.notification.send",
        })
        expect(structuredData(notification.structuredContent)).toMatchObject({
          status: "DELIVERED",
          booking_id: bookingId,
        })

        const cancelled = await client.callTool({
          arguments: { booking_id: bookingId },
          name: "booking.cancel",
        })
        expect(structuredData(cancelled.structuredContent)).toMatchObject({
          status: "CANCELLED",
        })

        const scheduleAfterCancel = await client.callTool({
          arguments: {},
          name: "doctor.schedule",
        })
        expect(
          arrayData(scheduleAfterCancel.structuredContent, "schedules")[0],
        ).toMatchObject({
          schedule_id: scheduleId,
          available: true,
        })

        const repeatedCancel = await client.callTool({
          arguments: { booking_id: bookingId },
          name: "booking.cancel",
        })
        expect(repeatedCancel.isError).toBe(true)

        const directBooking = await client.callTool({
          arguments: { schedule_id: "sch-happy-002" },
          name: "doctor.book",
        })
        expect(structuredData(directBooking.structuredContent)).toMatchObject({
          status: "CONFIRMED",
        })
        const duplicateDirectBooking = await client.callTool({
          arguments: { schedule_id: "sch-happy-002" },
          name: "doctor.book",
        })
        expect(duplicateDirectBooking.isError).toBe(true)
      })
    })
  })

  test("covers doctor catalog and availability outcomes", async () => {
    await withApp("doctor-no-results.json", async (app) => {
      await withClient(app, async (client) => {
        const lookup = await client.callTool({
          arguments: {},
          name: "doctor.lookup",
        })
        expect(structuredData(lookup.structuredContent)).toMatchObject({
          doctor: null,
          found: false,
        })

        const search = await client.callTool({
          arguments: {},
          name: "doctor.search",
        })
        expect(arrayData(search.structuredContent, "doctors")).toHaveLength(0)

        const recommendations = await client.callTool({
          arguments: {},
          name: "doctor.recommend",
        })
        expect(
          arrayData(recommendations.structuredContent, "recommendations"),
        ).toHaveLength(0)

        const schedule = await client.callTool({
          arguments: {},
          name: "doctor.schedule",
        })
        expect(arrayData(schedule.structuredContent, "schedules")).toHaveLength(
          0,
        )
      })
    })

    await withApp("doctor-unavailable.json", async (app) => {
      await withClient(app, async (client) => {
        const schedule = await client.callTool({
          arguments: {},
          name: "schedule.check",
        })
        expect(structuredData(schedule.structuredContent)).toMatchObject({
          schedule_available: false,
        })
        expect(
          arrayData(schedule.structuredContent, "alternatives"),
        ).toHaveLength(2)

        const unavailableReservation = await client.callTool({
          arguments: { schedule_id: "sch-unavailable-001" },
          name: "booking.reserve",
        })
        expect(unavailableReservation.isError).toBe(true)
        expect(unavailableReservation.content[0]).toMatchObject({
          text: expect.stringContaining("slot_unavailable"),
        })
      })
    })

    await withApp("doctor-queue-full.json", async (app) => {
      await withClient(app, async (client) => {
        const queue = await client.callTool({
          arguments: {},
          name: "queue.check",
        })
        expect(structuredData(queue.structuredContent)).toMatchObject({
          queue_available: false,
          next_available: "sch-queue-002",
        })
      })
    })

    await withApp("doctor-happy.json", async (app) => {
      await withClient(app, async (client) => {
        const missingSchedule = await client.callTool({
          arguments: { schedule_id: "sch-missing" },
          name: "booking.reserve",
        })
        expect(missingSchedule.isError).toBe(true)
        expect(missingSchedule.content[0]).toMatchObject({
          text: expect.stringContaining("schedule_not_found"),
        })

        const schedule = await client.callTool({
          arguments: {},
          name: "doctor.schedule",
        })
        expect(
          arrayData(schedule.structuredContent, "schedules")[0],
        ).toMatchObject({
          schedule_id: "sch-happy-001",
          available: true,
        })
      })
    })
  })

  test("preserves a pre-existing doctor appointment on conflict", async () => {
    await withApp("doctor-conflict.json", async (app) => {
      await withClient(app, async (client) => {
        const reservation = await client.callTool({
          arguments: { schedule_id: "sch-conflict-001" },
          name: "booking.reserve",
        })
        expect(reservation.isError).toBe(true)
        expect(reservation.content[0]).toMatchObject({
          text: expect.stringContaining("schedule_conflict"),
        })

        const directBooking = await client.callTool({
          arguments: { schedule_id: "sch-conflict-001" },
          name: "doctor.book",
        })
        expect(directBooking.isError).toBe(true)
        expect(directBooking.content[0]).toMatchObject({
          text: expect.stringContaining("schedule_conflict"),
        })

        const existingPayment = await client.callTool({
          arguments: { booking_id: "BKGD-0401" },
          name: "doctor.payment.create",
        })
        expect(structuredData(existingPayment.structuredContent)).toMatchObject(
          {
            booking_id: "BKGD-0401",
            status: "PENDING",
          },
        )
      })
    })
  })

  test("exposes independent doctor payment and notification failures", async () => {
    await withApp("doctor-payment-failed.json", async (app) => {
      await withClient(app, async (client) => {
        const reservation = await client.callTool({
          arguments: { schedule_id: "sch-payment-001" },
          name: "booking.reserve",
        })
        const reservationId = stringData(
          reservation.structuredContent,
          "reservation_id",
        )
        const appointment = await client.callTool({
          arguments: { reservation_id: reservationId },
          name: "booking.confirm",
        })
        const bookingId = stringData(
          appointment.structuredContent,
          "booking_id",
        )

        const payment = await client.callTool({
          arguments: { booking_id: bookingId },
          name: "doctor.payment.create",
        })
        expect(payment.isError).toBe(true)
        expect(payment.content[0]).toMatchObject({
          text: expect.stringContaining("payment_declined"),
        })

        const retry = await client.callTool({
          arguments: { booking_id: bookingId },
          name: "doctor.payment.create",
        })
        expect(retry.content[0]).toMatchObject({
          text: expect.stringContaining("payment_declined"),
        })
      })
    })

    await withApp("doctor-notification-failed.json", async (app) => {
      await withClient(app, async (client) => {
        const booking = await client.callTool({
          arguments: { schedule_id: "sch-notification-001" },
          name: "doctor.book",
        })
        const bookingId = stringData(booking.structuredContent, "booking_id")
        const notification = await client.callTool({
          arguments: {
            booking_id: bookingId,
            message: "Your doctor appointment is confirmed.",
          },
          name: "doctor.notification.send",
        })
        expect(notification.isError).toBe(true)
        expect(notification.content[0]).toMatchObject({
          text: expect.stringContaining("notification_delivery_failed"),
        })

        const aliasNotification = await client.callTool({
          arguments: {
            booking_id: bookingId,
            message: "Your doctor appointment is confirmed.",
          },
          name: "notification.send_confirmation",
        })
        expect(aliasNotification.isError).toBe(true)
      })
    })
  })

  test("keeps provider error, timeout, and malformed output observable", async () => {
    await withApp("doctor-provider-error.json", async (app) => {
      await withClient(app, async (client) => {
        const result = await client.callTool({
          arguments: {},
          name: "doctor.lookup",
        })
        expect(result.isError).toBe(true)
        expect(result.content[0]).toMatchObject({
          text: expect.stringContaining("doctor_catalog_unavailable"),
        })
      })
    })

    await withApp("doctor-invalid-output.json", async (app) => {
      await withClient(app, async (client) => {
        const { tools } = await client.listTools()
        expect(tools.map((tool) => tool.name)).toContain("doctor.lookup")
        const result = await client.callTool({
          arguments: {},
          name: "doctor.lookup",
        })
        expect(structuredData(result.structuredContent)).toMatchObject({
          doctor: "this should be an object",
          status: "intentionally-invalid",
        })
      })
    })

    await withApp("doctor-timeout.json", async (app) => {
      await withClient(app, async (client) => {
        const call = client.callTool({
          arguments: {},
          name: "schedule.check",
        })
        const timedOut = await Promise.race([
          call.then(() => false),
          Bun.sleep(25).then(() => true),
        ])
        expect(timedOut).toBe(true)
        expect(structuredData((await call).structuredContent)).toMatchObject({
          schedule_id: "sch-timeout-001",
        })
      })
    })
  })

  test("returns MCP tool errors without applying a booking", async () => {
    await withApp("padel-error.json", async (app) => {
      await withClient(app, async (client) => {
        const result = await client.callTool({
          arguments: { date: "2026-09-01", venue_id: "padel-senayan" },
          name: "padel.cek_available",
        })

        expect(result.isError).toBe(true)
      })
    })
  })

  test("rejects invalid input before mutating scenario state", async () => {
    await withApp("padel.json", async (app) => {
      await withClient(app, async (client) => {
        await client.listTools()

        const invalidBooking = await client.callTool({
          arguments: { date: "2026-09-01", venue_id: "padel-senayan" },
          name: "padel.create_booking",
        })
        expect(invalidBooking.isError).toBe(true)

        const availability = await client.callTool({
          arguments: { date: "2026-09-01", venue_id: "padel-senayan" },
          name: "padel.cek_available",
        })
        expect(structuredData(availability.structuredContent)).toMatchObject({
          availableSlots: ["18:00", "19:00"],
        })
      })
    })
  })

  test("does not complete a delayed scenario before its configured delay", async () => {
    await withApp("padel-delay.json", async (app) => {
      await withClient(app, async (client) => {
        const call = client.callTool({
          arguments: { date: "2026-09-01", venue_id: "padel-senayan" },
          name: "padel.cek_available",
        })
        const pending = await Promise.race([
          call.then(() => false),
          Bun.sleep(20).then(() => true),
        ])

        expect(pending).toBe(true)
        expect((await call).isError).not.toBe(true)
      })
    })
  })

  test("starts each process with a clean scenario state", async () => {
    await withApp("padel.json", async (app) => {
      await withClient(app, async (client) => {
        await client.callTool({
          arguments: {
            date: "2026-09-01",
            time: "18:00",
            venue_id: "padel-senayan",
          },
          name: "padel.create_booking",
        })
      })
    })

    await withApp("padel.json", async (app) => {
      await withClient(app, async (client) => {
        const availability = await client.callTool({
          arguments: { date: "2026-09-01", venue_id: "padel-senayan" },
          name: "padel.cek_available",
        })

        expect(structuredData(availability.structuredContent)).toMatchObject({
          availableSlots: ["18:00", "19:00"],
        })
      })
    })
  })
})
