import type { PadelBooking, PadelSlot, ProviderScenario } from "./scenario"

export type ToolData = Record<string, unknown>

export type ToolError = {
  code: string
  message: string
}

export type StoreResult = ToolData | ToolError

type RecordValue = Record<string, unknown>

function isRecord(value: unknown): value is RecordValue {
  return typeof value === "object" && value !== null && !Array.isArray(value)
}

function copyRecord(value: RecordValue): RecordValue {
  return structuredClone(value)
}

function copyRecords(value: unknown): RecordValue[] {
  return Array.isArray(value) && value.every(isRecord)
    ? value.map(copyRecord)
    : []
}

function fieldString(record: RecordValue, name: string): string | null {
  const value = record[name]
  return typeof value === "string" && value.length > 0 ? value : null
}

function fieldNumber(record: RecordValue, name: string): number | null {
  const value = record[name]
  return typeof value === "number" ? value : null
}

function error(code: string, message: string): ToolError {
  return { code, message }
}

function isToolError(value: StoreResult): value is ToolError {
  return "code" in value && "message" in value
}

export function isStoreError(value: StoreResult): value is ToolError {
  return isToolError(value)
}

export class ProviderScenarioStore {
  private readonly padelBookings: PadelBooking[]
  private readonly padelSlots: PadelSlot[]
  private readonly records: Map<string, RecordValue[]>
  private readonly counters: Map<string, number>

  constructor(scenario: ProviderScenario) {
    this.padelSlots = scenario.data.slots.map((slot) => ({ ...slot }))
    this.padelBookings = scenario.data.bookings.map((booking) => ({
      ...booking,
    }))
    this.records = new Map()
    this.counters = new Map()

    for (const [key, value] of Object.entries(scenario.data)) {
      if (key !== "slots" && key !== "bookings") {
        this.records.set(key, copyRecords(value))
      }
    }
  }

  private recordsFor(key: string): RecordValue[] {
    const records = this.records.get(key)
    if (records) {
      return records
    }

    const empty: RecordValue[] = []
    this.records.set(key, empty)
    return empty
  }

  private nextId(prefix: string): string {
    const next = (this.counters.get(prefix) ?? 0) + 1
    this.counters.set(prefix, next)
    return `${prefix}-${String(next).padStart(4, "0")}`
  }

  checkAvailability(
    venueId: string,
    date: string,
  ): { availableSlots: string[]; date: string; venueId: string } {
    const bookedTimes = new Set(
      this.padelBookings
        .filter(
          (booking) => booking.venueId === venueId && booking.date === date,
        )
        .map((booking) => booking.time),
    )
    const availableSlots = this.padelSlots
      .filter(
        (slot) =>
          slot.venueId === venueId &&
          slot.date === date &&
          !bookedTimes.has(slot.time),
      )
      .map((slot) => slot.time)

    return { availableSlots, date, venueId }
  }

  createBooking(
    venueId: string,
    date: string,
    time: string,
  ):
    | { bookingReference: string; date: string; time: string; venueId: string }
    | ToolError {
    const hasSlot = this.padelSlots.some(
      (slot) =>
        slot.venueId === venueId && slot.date === date && slot.time === time,
    )

    if (!hasSlot) {
      return error("slot_not_found", "requested padel slot is not configured")
    }

    const isBooked = this.padelBookings.some(
      (booking) =>
        booking.venueId === venueId &&
        booking.date === date &&
        booking.time === time,
    )

    if (isBooked) {
      return error("slot_unavailable", "requested padel slot is already booked")
    }

    const bookingReference = `PAD-${String(this.padelBookings.length + 1).padStart(4, "0")}`
    this.padelBookings.push({ bookingReference, date, time, venueId })

    return { bookingReference, date, time, venueId }
  }

  createPadelCourtBooking(
    venueId: string,
    date: string,
    time: string,
  ): StoreResult {
    const booking = this.createBooking(venueId, date, time)
    if (isToolError(booking)) {
      return booking
    }

    return {
      booking_id: booking.bookingReference,
      status: "CONFIRMED",
      venue_id: booking.venueId,
      date: booking.date,
      time: booking.time,
      payment_status: "PENDING",
    }
  }

  padelAvailability(venueId: string, date: string): ToolData {
    const availability = this.checkAvailability(venueId, date)
    const bookedSlots = this.padelSlots
      .filter(
        (slot) =>
          slot.venueId === venueId &&
          slot.date === date &&
          !availability.availableSlots.includes(slot.time),
      )
      .map((slot) => ({ date: slot.date, time: slot.time, available: false }))
    const availableSlots = availability.availableSlots.map((time) => ({
      date,
      time,
      available: true,
    }))

    return {
      court_id: venueId,
      location: venueId,
      booked_slots: bookedSlots,
      available_slots: availableSlots,
    }
  }

  createPadelPayment(bookingId: string): StoreResult {
    const booking = this.padelBookings.find(
      (item) => item.bookingReference === bookingId,
    )
    if (!booking) {
      return error("booking_not_found", "padel booking was not found")
    }

    const payments = this.recordsFor("payments")
    const existing = payments.find(
      (payment) => payment.booking_id === bookingId,
    )
    if (existing) {
      return error(
        "payment_exists",
        "a payment already exists for this booking",
      )
    }

    const payment = {
      payment_id: this.nextId("PAY-PADEL"),
      booking_id: bookingId,
      amount: 150000,
      currency: "IDR",
      status: "PENDING",
      payment_url: `https://pay.example.com/${bookingId}`,
    }
    payments.push(payment)
    return copyRecord(payment)
  }

  verifyPadelPayment(paymentId: string): StoreResult {
    const payments = this.recordsFor("payments")
    const payment = payments.find((item) => item.payment_id === paymentId)
    if (!payment) {
      return error("payment_not_found", "padel payment was not found")
    }
    if (payment.status !== "PENDING") {
      return error("invalid_transition", "padel payment is not pending")
    }

    payment.status = "PAID"
    payment.paid_at = "2026-08-30T06:50:00Z"
    return copyRecord(payment)
  }

  sendPadelNotification(bookingId: string, message: string): StoreResult {
    const booking = this.padelBookings.find(
      (item) => item.bookingReference === bookingId,
    )
    if (!booking) {
      return error("booking_not_found", "padel booking was not found")
    }

    const notifications = this.recordsFor("notifications")
    const notification = {
      notification_id: this.nextId("NOTIF-PADEL"),
      booking_id: bookingId,
      channel: "whatsapp",
      status: "DELIVERED",
      message,
    }
    notifications.push(notification)
    return copyRecord(notification)
  }

  addFoodCart(
    cartId: string | null,
    menuId: string,
    quantity: number,
  ): StoreResult {
    if (!Number.isInteger(quantity) || quantity < 1) {
      return error("invalid_input", "quantity must be a positive integer")
    }

    const menu = this.recordsFor("menuItems").find((item) => item.id === menuId)
    if (!menu) {
      return error("menu_item_not_found", "menu item was not found")
    }

    const carts = this.recordsFor("carts")
    const resolvedCartId = cartId ?? this.nextId("CART")
    let cart = carts.find((item) => item.cart_id === resolvedCartId)
    if (!cart) {
      cart = { cart_id: resolvedCartId, status: "OPEN", items: [] }
      carts.push(cart)
    }
    if (cart.status !== "OPEN") {
      return error("invalid_transition", "food cart is no longer open")
    }

    const items = Array.isArray(cart.items) ? cart.items : []
    const existing = items.find(
      (item): item is RecordValue => isRecord(item) && item.menu_id === menuId,
    )
    const price = fieldNumber(menu, "price") ?? 0
    if (existing) {
      const existingQuantity = fieldNumber(existing, "quantity") ?? 0
      existing.quantity = existingQuantity + quantity
      existing.subtotal = (existingQuantity + quantity) * price
    } else {
      items.push({
        menu_id: menuId,
        name: fieldString(menu, "name") ?? menuId,
        quantity,
        subtotal: quantity * price,
      })
    }
    cart.items = items
    return this.foodCartResult(cart)
  }

  private foodCartResult(cart: RecordValue): ToolData {
    const items = Array.isArray(cart.items)
      ? cart.items.filter(isRecord).map(copyRecord)
      : []
    const total = items.reduce(
      (sum, item) => sum + (fieldNumber(item, "subtotal") ?? 0),
      0,
    )
    return {
      cart_id: cart.cart_id,
      status: cart.status,
      items,
      total,
      item_count: items.reduce(
        (sum, item) => sum + (fieldNumber(item, "quantity") ?? 0),
        0,
      ),
    }
  }

  getFoodCart(cartId: string): StoreResult {
    const cart = this.recordsFor("carts").find(
      (item) => item.cart_id === cartId,
    )
    return cart
      ? this.foodCartResult(cart)
      : error("cart_not_found", "food cart was not found")
  }

  createFoodOrder(cartId: string, deliveryAddress: string): StoreResult {
    const cart = this.recordsFor("carts").find(
      (item) => item.cart_id === cartId,
    )
    if (!cart) {
      return error("cart_not_found", "food cart was not found")
    }
    if (cart.status !== "OPEN") {
      return error("invalid_transition", "food cart is no longer open")
    }

    const cartResult = this.foodCartResult(cart)
    const items = Array.isArray(cartResult.items) ? cartResult.items : []
    if (items.length === 0) {
      return error("cart_empty", "cannot create an order from an empty cart")
    }

    const orders = this.recordsFor("orders")
    const order = {
      order_id: this.nextId("ORD"),
      status: "PENDING_PAYMENT",
      items,
      cart_id: cartId,
      delivery_address: deliveryAddress,
      total: cartResult.total,
      delivery_fee: 5000,
      grand_total: (cartResult.total as number) + 5000,
      estimated_delivery: "30-45 minutes",
      payment_status: "PENDING",
    }
    orders.push(order)
    cart.status = "CHECKED_OUT"
    return copyRecord(order)
  }

  trackFoodOrder(orderId: string): StoreResult {
    const order = this.recordsFor("orders").find(
      (item) => item.order_id === orderId,
    )
    return order
      ? copyRecord(order)
      : error("order_not_found", "food order was not found")
  }

  createFoodPayment(orderId: string): StoreResult {
    const order = this.recordsFor("orders").find(
      (item) => item.order_id === orderId,
    )
    if (!order) {
      return error("order_not_found", "food order was not found")
    }
    const payments = this.recordsFor("payments")
    if (payments.some((item) => item.order_id === orderId)) {
      return error("payment_exists", "a payment already exists for this order")
    }

    const payment = {
      payment_id: this.nextId("PAY-FOOD"),
      order_id: orderId,
      amount: order.grand_total,
      currency: "IDR",
      status: "PENDING",
      payment_url: `https://pay.example.com/${orderId}`,
    }
    payments.push(payment)
    return copyRecord(payment)
  }

  verifyFoodPayment(paymentId: string): StoreResult {
    const payments = this.recordsFor("payments")
    const payment = payments.find((item) => item.payment_id === paymentId)
    if (!payment) {
      return error("payment_not_found", "food payment was not found")
    }
    if (payment.status !== "PENDING") {
      return error("invalid_transition", "food payment is not pending")
    }

    payment.status = "PAID"
    payment.paid_at = "2026-08-30T06:50:00Z"
    const orderId = fieldString(payment, "order_id")
    const order = this.recordsFor("orders").find(
      (item) => item.order_id === orderId,
    )
    if (order) {
      order.payment_status = "PAID"
      order.status = "ON_DELIVERY"
    }
    return copyRecord(payment)
  }

  doctorSchedule(): ToolData {
    const schedules = this.recordsFor("schedules").map(copyRecord)
    return {
      doctor_id: schedules[0]?.doctor_id ?? "doc-001",
      doctor_name: schedules[0]?.doctor_name ?? "Doctor",
      schedules,
    }
  }

  reserveDoctorAppointment(scheduleId: string): StoreResult {
    const schedule = this.recordsFor("schedules").find(
      (item) => item.schedule_id === scheduleId,
    )
    if (!schedule) {
      return error("schedule_not_found", "doctor schedule was not found")
    }
    if (schedule.available !== true) {
      return error("slot_unavailable", "doctor schedule is not available")
    }

    const reservations = this.recordsFor("reservations")
    const reservation = {
      reservation_id: this.nextId("RES"),
      doctor_id: schedule.doctor_id,
      schedule_id: scheduleId,
      date: schedule.date,
      time: schedule.time,
      status: "RESERVED",
      expires_at: "2026-08-30T10:02:58Z",
    }
    reservations.push(reservation)
    schedule.available = false
    schedule.quota_remaining = 0
    return copyRecord(reservation)
  }

  confirmDoctorReservation(reservationId: string): StoreResult {
    const reservation = this.recordsFor("reservations").find(
      (item) => item.reservation_id === reservationId,
    )
    if (!reservation) {
      return error("reservation_not_found", "doctor reservation was not found")
    }
    if (reservation.status !== "RESERVED") {
      return error("invalid_transition", "doctor reservation is not pending")
    }

    const appointments = this.recordsFor("appointments")
    const appointment = {
      booking_id: this.nextId("BKGD"),
      status: "CONFIRMED",
      doctor_id: reservation.doctor_id,
      schedule_id: reservation.schedule_id,
      date: reservation.date,
      time: reservation.time,
      payment_status: "PENDING",
    }
    appointments.push(appointment)
    reservation.status = "CONFIRMED"
    return copyRecord(appointment)
  }

  cancelDoctorAppointment(bookingId: string): StoreResult {
    const appointments = this.recordsFor("appointments")
    const appointment = appointments.find(
      (item) => item.booking_id === bookingId,
    )
    if (!appointment) {
      return error("booking_not_found", "doctor appointment was not found")
    }
    if (appointment.status !== "CONFIRMED") {
      return error("invalid_transition", "doctor appointment is not confirmed")
    }

    appointment.status = "CANCELLED"
    const schedule = this.recordsFor("schedules").find(
      (item) => item.schedule_id === appointment.schedule_id,
    )
    if (schedule) {
      schedule.available = true
      schedule.quota_remaining = 1
    }
    return copyRecord(appointment)
  }

  bookDoctorAppointment(scheduleId: string): StoreResult {
    const schedule = this.recordsFor("schedules").find(
      (item) => item.schedule_id === scheduleId,
    )
    if (!schedule) {
      return error("schedule_not_found", "doctor schedule was not found")
    }
    if (schedule.available !== true) {
      return error("slot_unavailable", "doctor schedule is not available")
    }

    const appointments = this.recordsFor("appointments")
    const appointment = {
      booking_id: this.nextId("BKGD"),
      status: "CONFIRMED",
      doctor_id: schedule.doctor_id,
      schedule_id: scheduleId,
      date: schedule.date,
      time: schedule.time,
      payment_status: "PENDING",
    }
    appointments.push(appointment)
    schedule.available = false
    schedule.quota_remaining = 0
    return copyRecord(appointment)
  }

  createDoctorPayment(bookingId: string): StoreResult {
    const appointment = this.recordsFor("appointments").find(
      (item) => item.booking_id === bookingId,
    )
    if (!appointment) {
      return error("booking_not_found", "doctor appointment was not found")
    }
    const payments = this.recordsFor("payments")
    if (payments.some((item) => item.booking_id === bookingId)) {
      return error(
        "payment_exists",
        "a payment already exists for this appointment",
      )
    }

    const payment = {
      payment_id: this.nextId("PAY-DOC"),
      booking_id: bookingId,
      amount: 200000,
      currency: "IDR",
      status: "PENDING",
      payment_url: `https://pay.example.com/${bookingId}`,
    }
    payments.push(payment)
    return copyRecord(payment)
  }

  verifyDoctorPayment(paymentId: string): StoreResult {
    const payments = this.recordsFor("payments")
    const payment = payments.find((item) => item.payment_id === paymentId)
    if (!payment) {
      return error("payment_not_found", "doctor payment was not found")
    }
    if (payment.status !== "PENDING") {
      return error("invalid_transition", "doctor payment is not pending")
    }

    payment.status = "PAID"
    payment.paid_at = "2026-08-30T07:10:00Z"
    const appointment = this.recordsFor("appointments").find(
      (item) => item.booking_id === payment.booking_id,
    )
    if (appointment) {
      appointment.payment_status = "PAID"
    }
    return copyRecord(payment)
  }

  sendDoctorNotification(bookingId: string, message: string): StoreResult {
    const appointment = this.recordsFor("appointments").find(
      (item) => item.booking_id === bookingId,
    )
    if (!appointment) {
      return error("booking_not_found", "doctor appointment was not found")
    }

    const notifications = this.recordsFor("notifications")
    const notification = {
      notification_id: this.nextId("NOTIF-DOC"),
      booking_id: bookingId,
      channel: "whatsapp",
      status: "DELIVERED",
      message,
    }
    notifications.push(notification)
    return copyRecord(notification)
  }
}

export class PadelScenarioStore extends ProviderScenarioStore {}
