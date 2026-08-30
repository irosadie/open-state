type ErrorWithStatus = {
  status?: unknown
  statusCode?: unknown
}

export type AuthRecoveryAction = "login" | "forbidden" | "none"

export const getApiErrorStatus = (error: unknown) => {
  if (!error || typeof error !== "object") return undefined

  const candidate = error as ErrorWithStatus
  const status = candidate.status ?? candidate.statusCode

  return typeof status === "number" ? status : undefined
}

export const getAuthRecoveryAction = (
  status: number | undefined,
): AuthRecoveryAction => {
  if (status === 401) return "login"
  if (status === 403) return "forbidden"
  return "none"
}
