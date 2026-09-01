import { describe, expect, it } from "vitest"
import { intentListResponseSchema, intentResponseSchema } from "./intent"

const bookingIntent = {
  id: "intent-1",
  tenantId: "tenant-1",
  projectId: "project-1",
  workflowId: "workflow-1",
  key: "BOOKING_PADEL",
  name: "Booking Padel",
  description: "Book a padel court",
  examples: ["saya mau order lapangan"],
  workflowSlug: "booking-padel",
}

describe("intent response schemas", () => {
  it("accepts a routable intent list", () => {
    expect(intentListResponseSchema.parse([bookingIntent])).toEqual([
      bookingIntent,
    ])
  })

  it("rejects an incomplete intent item", () => {
    expect(() => intentResponseSchema.parse({ key: "BOOKING_PADEL" })).toThrow()
  })
})
