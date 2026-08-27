export type SuccessResponse<T> = {
  success: true
  message: string
  data: T
  meta?: unknown
}
