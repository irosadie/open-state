const masking = (string: string, start = 4, end = 4) => {
  return `${string.substring(start, end)}****${string.substring(string.length - start, string.length)}`
}

export default masking
