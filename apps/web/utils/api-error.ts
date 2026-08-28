type ClassifiedError = {
  kind?: string
  code?: string
  message?: string
}

type ApiErrorLike = {
  message?: string
  error?: string | ClassifiedError
  errors?: string | ClassifiedError
}

// Extract a human-readable message from the various error shapes the API may
// return (the generic `{"error": "..."}` shape, the classified capability error
// `{"error": {kind, code, message}}`, or a `{message}` shape). Never a raw
// provider error.
export const extractErrorMessage = (err: unknown): string | undefined => {
  if (typeof err === "string") return err

  if (err && typeof err === "object") {
    const e = err as ApiErrorLike

    if (typeof e.message === "string" && e.message) return e.message

    if (typeof e.error === "string" && e.error) return e.error

    if (typeof e.errors === "string" && e.errors) return e.errors

    if (typeof e.error === "object" && e.error?.message) return e.error.message

    if (typeof e.errors === "object" && e.errors?.message)
      return e.errors.message
  }

  return undefined
}

// Extract the classified capability error (kind/code/message, PRD §87) from an
// API error, if present. Used by the test-invocation panel to surface the
// classified kind/code to operators.
export const extractClassifiedError = (
  err: unknown,
): ClassifiedError | undefined => {
  if (err && typeof err === "object") {
    const e = err as ApiErrorLike

    const classified = e.error ?? e.errors

    if (classified && typeof classified === "object") {
      return classified
    }
  }

  return undefined
}
