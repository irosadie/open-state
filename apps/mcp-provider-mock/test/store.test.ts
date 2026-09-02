import { describe, expect, test } from "bun:test"

import { loadScenario } from "../src/scenario"
import { ProviderScenarioStore, isStoreError } from "../src/store"

const fixturePath = (name: string) =>
  new URL(`../fixtures/${name}`, import.meta.url).pathname

function stringField(record: Record<string, unknown>, field: string): string {
  const value = record[field]
  if (typeof value !== "string") {
    throw new Error(`${field} was not a string`)
  }
  return value
}

describe("doctor scenario store", () => {
  test("preserves lifecycle state across invalid and repeated operations", async () => {
    const scenario = await loadScenario(fixturePath("doctor-happy.json"))
    const store = new ProviderScenarioStore(scenario)

    const reservation = store.reserveDoctorAppointment("sch-happy-001")
    if (isStoreError(reservation)) {
      throw new Error(reservation.message)
    }
    expect(reservation).toMatchObject({
      reservation_id: "RES-0001",
      status: "RESERVED",
    })

    const repeatedReservation = store.reserveDoctorAppointment("sch-happy-001")
    expect(isStoreError(repeatedReservation)).toBe(true)
    if (!isStoreError(repeatedReservation)) {
      throw new Error("repeated reservation unexpectedly succeeded")
    }
    expect(repeatedReservation.code).toBe("schedule_conflict")

    const confirmed = store.confirmDoctorReservation(
      stringField(reservation, "reservation_id"),
    )
    if (isStoreError(confirmed)) {
      throw new Error(confirmed.message)
    }
    expect(confirmed).toMatchObject({
      booking_id: "BKGD-0001",
      status: "CONFIRMED",
    })

    const repeatedConfirmation = store.confirmDoctorReservation(
      stringField(reservation, "reservation_id"),
    )
    expect(isStoreError(repeatedConfirmation)).toBe(true)

    const bookingId = stringField(confirmed, "booking_id")
    const cancelled = store.cancelDoctorAppointment(bookingId)
    if (isStoreError(cancelled)) {
      throw new Error(cancelled.message)
    }
    expect(cancelled).toMatchObject({ status: "CANCELLED" })

    const repeatedCancellation = store.cancelDoctorAppointment(bookingId)
    expect(isStoreError(repeatedCancellation)).toBe(true)

    const rebooked = store.reserveDoctorAppointment("sch-happy-001")
    if (isStoreError(rebooked)) {
      throw new Error(rebooked.message)
    }
    expect(rebooked).toMatchObject({
      reservation_id: "RES-0002",
      status: "RESERVED",
    })

    const unknownBooking = store.cancelDoctorAppointment("BKGD-UNKNOWN")
    expect(isStoreError(unknownBooking)).toBe(true)
  })

  test("returns a conflict for a schedule with a pre-existing appointment", async () => {
    const scenario = await loadScenario(fixturePath("doctor-conflict.json"))
    const store = new ProviderScenarioStore(scenario)

    const reservation = store.reserveDoctorAppointment("sch-conflict-001")
    expect(isStoreError(reservation)).toBe(true)
    if (!isStoreError(reservation)) {
      throw new Error("conflicted schedule unexpectedly succeeded")
    }
    expect(reservation).toMatchObject({ code: "schedule_conflict" })

    const directBooking = store.bookDoctorAppointment("sch-conflict-001")
    expect(isStoreError(directBooking)).toBe(true)
    if (!isStoreError(directBooking)) {
      throw new Error("conflicted direct booking unexpectedly succeeded")
    }
    expect(directBooking).toMatchObject({ code: "schedule_conflict" })
  })

  test("rejects payment and notification for cancelled appointments", async () => {
    const scenario = await loadScenario(fixturePath("doctor-happy.json"))
    const store = new ProviderScenarioStore(scenario)

    const booking = store.bookDoctorAppointment("sch-happy-002")
    if (isStoreError(booking)) {
      throw new Error(booking.message)
    }
    const bookingId = stringField(booking, "booking_id")
    const cancelled = store.cancelDoctorAppointment(bookingId)
    if (isStoreError(cancelled)) {
      throw new Error(cancelled.message)
    }

    const payment = store.createDoctorPayment(bookingId)
    expect(isStoreError(payment)).toBe(true)
    const notification = store.sendDoctorNotification(
      bookingId,
      "This should not be delivered.",
    )
    expect(isStoreError(notification)).toBe(true)
  })
})
