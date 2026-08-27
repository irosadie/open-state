const MINUTES_PER_DAY = 60 * 24

const daysToMinutes = (days: string) => {
  const parsedDays = Number(days)

  if (!Number.isFinite(parsedDays) || parsedDays <= 0) {
    return 0
  }

  return Math.round(parsedDays * MINUTES_PER_DAY)
}

export default daysToMinutes
