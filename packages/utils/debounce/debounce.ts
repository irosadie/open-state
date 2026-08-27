type DebounceOptions = {
  leading?: boolean
}

const debounce = <F extends (...args: never[]) => void>(
  func: F,
  delay: number,
  options: DebounceOptions = {},
) => {
  let timeoutId: ReturnType<typeof setTimeout> | null = null
  let lastCallTime: number | null = null

  const debounced = (...args: Parameters<F>) => {
    const now = Date.now()

    if (options.leading && !lastCallTime) {
      func(...args)
      lastCallTime = now
    }

    if (timeoutId) {
      clearTimeout(timeoutId)
    }

    timeoutId = setTimeout(() => {
      if (!options.leading || (lastCallTime && now - lastCallTime >= delay)) {
        func(...args)
      }
      lastCallTime = null
      timeoutId = null
    }, delay)
  }

  debounced.cancel = () => {
    if (timeoutId) {
      clearTimeout(timeoutId)
    }
    timeoutId = null
    lastCallTime = null
  }

  return debounced as typeof debounced & { cancel: () => void }
}

export default debounce
