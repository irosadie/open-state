const pointConverter = (point: number) => {
  if (point <= 0) {
    return "Sangat Buruk"
  }
  if (point <= 20) {
    return "Buruk"
  }
  if (point <= 60) {
    return "Lumayan"
  }
  if (point <= 80) {
    return "Baik"
  }
  if (point <= 100) {
    return "Sangat Baik"
  }
}

export default pointConverter
