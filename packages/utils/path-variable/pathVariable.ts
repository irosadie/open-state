const pathVariable = <T extends string>(
  url: T,
  params: PathParams<T>,
): string => {
  let result = url

  for (const key of Object.keys(params)) {
    result = result.replace(
      `:${key}`,
      String(params[key as keyof typeof params]),
    ) as T
  }

  return result
}

type ExtractPathParams<T extends string> =
  T extends `${infer _Start}:${infer Param}/${infer Rest}`
    ? Param | ExtractPathParams<`/${Rest}`>
    : T extends `${infer _Start}:${infer Param}`
      ? Param
      : never

// Map path parameters to an object type
type PathParams<T extends string> = ExtractPathParams<T> extends string
  ? Record<ExtractPathParams<T>, string | number>
  : Record<string, never>

export default pathVariable
