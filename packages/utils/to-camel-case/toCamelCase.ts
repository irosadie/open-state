const toCamelCase = <T>(obj: T): T => {
  if (Array.isArray(obj)) {
    return obj.map((v) => toCamelCase(v)) as T
  }

  if (obj && typeof obj === "object") {
    return Object.keys(obj).reduce(
      (result, key) => {
        const camelCaseKey = key.replace(/([-_][a-z])/gi, (match) =>
          match.toUpperCase().replace("-", "").replace("_", ""),
        )
        ;(result as Record<string, unknown>)[camelCaseKey] = toCamelCase(
          (obj as Record<string, unknown>)[key],
        )
        return result
      },
      {} as Record<string, unknown>,
    ) as T
  }

  return obj
}

export default toCamelCase
