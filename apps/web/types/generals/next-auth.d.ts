// types/next-auth.d.ts
import type { DefaultSession } from "next-auth"

declare module "next-auth" {
  interface Session {
    user: {
      id: string
      name?: string | null
      email?: string | null
      photo?: string
      companyId: number
    } & DefaultSession["user"]
    accessToken?: string
    refreshToken?: string
    accessTokenExpires?: number
    error?: string
  }

  interface User {
    id: string
    name: string
    email: string
    photo?: string
    companyId: number
    accessToken?: string
    refreshToken?: string
    accessTokenExpires?: number
  }
}

declare module "next-auth/jwt" {
  interface JWT {
    id?: string
    name?: string
    email?: string
    photo?: string
    companyId?: number
    accessToken?: string
    refreshToken?: string
    accessTokenExpires?: number
    error?: string
  }
}
