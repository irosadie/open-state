import type { AuthUserResponse } from "./auth-response"

export type AuthSessionResponse = {
  expiresAt: string | null
}

export type AuthCurrentUserResponse = {
  user: AuthUserResponse
  session: AuthSessionResponse
}
