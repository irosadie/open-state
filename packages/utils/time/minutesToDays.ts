const MINUTES_PER_DAY = 60 * 24

const minutesToDays = (minutes: number) => {
  if (!minutes) {
    return ""
  }

  return (minutes / MINUTES_PER_DAY).toString()
}

export default minutesToDays
