import { authConfig } from "$/configs/auth"
import { serverAuthConfig } from "$/configs/auth-server"
import { loginSchema } from "@vibecoding-starter/schemas"
import type { AuthLoginResponse } from "@vibecoding-starter/types"
import axios from "axios"
import type { NextAuthOptions, User } from "next-auth"
import CredentialsProvider from "next-auth/providers/credentials"

type AuthEnvelope<T> = T | { data: T }

const loginProxyUrl = `${serverAuthConfig.appBaseUrl}${authConfig.proxyApiBasePath}${authConfig.backendLoginPath}`

const unwrapData = <T>(payload: AuthEnvelope<T>): T => {
  if (
    typeof payload === "object" &&
    payload !== null &&
    "data" in payload &&
    payload.data
  ) {
    return payload.data as T
  }

  return payload as T
}

export const authOptions: NextAuthOptions = {
  debug: !serverAuthConfig.secureCookies,
  providers: [
    CredentialsProvider({
      name: "credentials",
      id: "credentials",
      credentials: {
        email: { label: "Email", type: "text" },
        password: { label: "Password", type: "password" },
      },
      authorize: async (credentials) => {
        const parsedCredentials = loginSchema.safeParse(credentials)

        if (!parsedCredentials.success) {
          throw new Error(
            parsedCredentials.error.issues[0]?.message ?? "Invalid credentials",
          )
        }

        try {
          const { data } = await axios.post<AuthEnvelope<AuthLoginResponse>>(
            loginProxyUrl,
            parsedCredentials.data,
            {
              headers: {
                "Content-Type": "application/json",
              },
            },
          )

          const result = unwrapData(data)

          if (!result.user || !result.tokens) {
            throw new Error("Invalid response from auth server")
          }

          const accessTokenExpires = Date.now() + result.tokens.expiresIn * 1000

          const authUser: User = {
            id: result.user.id,
            email: result.user.email,
            name: result.user.name,
            photo: result.user.photo ?? undefined,
            companyId: result.user.companyId ?? 0,
            accessToken: result.tokens.accessToken,
            refreshToken: result.tokens.refreshToken,
            accessTokenExpires,
            image: result.user.photo ?? undefined,
          }

          return authUser
        } catch (error: unknown) {
          if (axios.isAxiosError(error)) {
            const errorMessage =
              (error.response?.data as { message?: string })?.message ||
              "Login failed"

            throw new Error(errorMessage)
          }

          if (error instanceof Error) {
            throw new Error(error.message)
          }

          throw new Error("An unexpected error occurred")
        }
      },
    }),
  ],
  callbacks: {
    async jwt({ token, user, account, trigger, session }) {
      if (user && account) {
        token.id = user.id
        token.name = user.name
        token.email = user.email
        token.photo = user.photo
        token.companyId = user.companyId
        token.accessToken = user.accessToken
        token.refreshToken = user.refreshToken
        token.accessTokenExpires = user.accessTokenExpires
      }

      if (trigger === "update" && session) {
        if (session.accessToken) {
          token.accessToken = session.accessToken
        }
        if (session.refreshToken) {
          token.refreshToken = session.refreshToken
        }
        if (session.accessTokenExpires) {
          token.accessTokenExpires = session.accessTokenExpires
        }
        if (session.error) {
          token.error = session.error
        }
      }

      return token
    },
    async session({ session, token }) {
      session.accessToken = token.accessToken
      session.refreshToken = token.refreshToken
      session.accessTokenExpires = token.accessTokenExpires
      session.error = token.error
      session.user = {
        ...session.user,
        id: token.id as string,
        email: token.email ?? undefined,
        name: token.name ?? undefined,
        photo: token.photo ?? undefined,
        companyId: (token.companyId as number | undefined) ?? 0,
      }

      return session
    },
  },
  secret: serverAuthConfig.secret,
  session: {
    strategy: "jwt",
    maxAge: authConfig.sessionMaxAgeSeconds,
  },
  jwt: {
    secret: serverAuthConfig.secret,
  },
  pages: {
    signIn: authConfig.loginPath,
    error: authConfig.loginPath,
  },
}
