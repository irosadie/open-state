type PERIOD = "ALL" | "WEEK" | "MONTH" | "YEAR"
type TYPE = "ROLLING" | "CALENDAR"
type DIRECTION = "FORWARD" | "BACKWARD"

type DateRangeParams =
  | { period: "ALL"; direction: DIRECTION; startDate?: Date }
  | {
      period: Exclude<PERIOD, "ALL">
      value: number
      type: TYPE
      direction: DIRECTION
      startDate?: Date
    }

const dateRange = (params: DateRangeParams): { start: Date; end: Date } => {
  const now = params.startDate || new Date() // Use provided startDate, or current date
  let start: Date
  let end: Date

  if (params.period === "ALL") {
    start = new Date(now.getFullYear(), 0, 1) // Start of the current year
    if (params.direction === "FORWARD") {
      end = new Date(9999, 11, 31, 23, 59, 59) // End of the year, until December 31, 9999
    } else {
      start = new Date(1945, 0, 1) // Start from the year when computers were widely available (around 1945)
      end = new Date(now.getFullYear(), 11, 31, 23, 59, 59) // End of the current year
    }
    return { start, end }
  }

  const { period, value, type, direction } = params

  switch (period) {
    case "YEAR":
      if (type === "ROLLING") {
        start = new Date(now)
        end = new Date(now)
        if (direction === "FORWARD") {
          end.setFullYear(now.getFullYear() + value)
        } else {
          start.setFullYear(now.getFullYear() - value)
        }
      } else {
        // CALENDAR type
        if (direction === "FORWARD") {
          start = new Date(now.getFullYear(), 0, 1) // Start of the current year
          end = new Date(now) // End is today
          if (value > 1) {
            start.setFullYear(now.getFullYear() - (value - 1)) // Adjust to start from January 1 of (current year - (value - 1))
          }
        } else {
          // BACKWARD
          start = new Date(now.getFullYear() - value, 0, 1) // Start from January 1 of calculated year
          end = new Date(now) // End is today
        }
      }
      break

    case "MONTH":
      if (type === "ROLLING") {
        start = new Date(now)
        end = new Date(now)
        if (direction === "FORWARD") {
          end.setMonth(now.getMonth() + value)
        } else {
          start.setMonth(now.getMonth() - value)
        }
      } else {
        // CALENDAR type
        if (direction === "FORWARD") {
          start = new Date(now.getFullYear(), now.getMonth(), 1) // Start of the current month
          end = new Date(now) // End is today
          if (value > 1) {
            start.setMonth(now.getMonth() - (value - 1)) // Adjust to start from the 1st of (current month - (value - 1))
          }
        } else {
          // BACKWARD
          start = new Date(now.getFullYear(), now.getMonth() - value, 1) // Start from the 1st of calculated month
          end = new Date(now) // End is today
        }
      }
      break

    case "WEEK":
      if (type === "ROLLING") {
        start = new Date(now)
        end = new Date(now)
        if (direction === "FORWARD") {
          end.setDate(now.getDate() + value * 7)
        } else {
          start.setDate(now.getDate() - value * 7)
        }
      } else {
        // CALENDAR type
        const dayOfWeek = now.getDay()

        if (direction === "FORWARD") {
          start = new Date(now)
          start.setDate(now.getDate() - dayOfWeek) // Start of the current week (previous Sunday)
          end = new Date(now) // End is today
          if (value > 1) {
            start.setDate(start.getDate() - (value - 1) * 7) // Adjust to start of (current week - (value - 1))
          }
        } else {
          // BACKWARD
          start = new Date(now)
          start.setDate(now.getDate() - dayOfWeek - value * 7) // Start of calculated week
          end = new Date(now) // End is today
        }
      }
      break

    default:
      throw new Error("Invalid period")
  }

  return { start, end }
}

export default dateRange
