const toPositiveNumber = (value: string) => {
  const parsedValue = Number(value)

  if (!Number.isFinite(parsedValue) || parsedValue <= 0) {
    return 0
  }

  return parsedValue
}

export default toPositiveNumber
