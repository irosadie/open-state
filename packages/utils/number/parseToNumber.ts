const parseToNumber = (val: string) => {
  return Number(val.replace(/[^0-9]/g, "")) || 0
}

export default parseToNumber
