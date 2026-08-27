export type AuthUserResponse = {
  id: string
  email: string
  name: string
  companyId?: number | null
  photo?: string | null
  role?: string | null
}

export type AuthTokensResponse = {
  accessToken: string
  refreshToken: string
  expiresIn: number
}

export type AuthLoginResponse = {
  user: AuthUserResponse
  tokens: AuthTokensResponse
}

export type AuthRegisterResponse = AuthLoginResponse

export type AuthCurrentUserResponse = AuthUserResponse
