export type ErrorResponse<T = unknown> = {
  success: false
  message?: string
  errors: T
}
