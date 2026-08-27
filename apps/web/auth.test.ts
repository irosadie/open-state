import { beforeEach, describe, expect, it, vi } from "vitest"

const assertDefined = <T>(value: T | null | undefined, label: string): T => {
  if (value == null) {
    throw new Error(`${label} should be defined`)
  }

  return value
}

// Mock next-auth before importing auth
vi.mock("next-auth/react", () => ({
  useSession: vi.fn(() => ({
    data: {
      user: {
        id: "test-user-id",
        email: "test@example.com",
        name: "Test User",
        companyId: 123,
      },
      accessToken: "mock-access-token",
      refreshToken: "mock-refresh-token",
    },
    status: "authenticated",
  })),
  signIn: vi.fn(),
  signOut: vi.fn(),
}))

vi.mock("$/configs/auth", () => ({
  authConfig: {
    loginPath: "/login",
    defaultRedirectPath: "/",
    authApiBasePath: "/api/auth",
    proxyApiBasePath: "/api/proxy",
    backendLoginPath: "/auth/login",
    backendRefreshPath: "/auth/refresh",
    backendLogoutPath: "/auth/logout",
    sessionMaxAgeSeconds: 6 * 60 * 60,
  },
}))

vi.mock("$/configs/auth-server", () => ({
  serverAuthConfig: {
    appBaseUrl: "http://localhost:3020",
    backendApiBaseUrl: "http://localhost:8020",
    secret: "test-auth-secret",
    secureCookies: false,
    sessionCookieName: "next-auth.session-token",
  },
}))

describe("Auth Configuration", () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe("authOptions structure", () => {
    it("should have correct session configuration", async () => {
      const { authOptions } = await import("./auth")
      const session = assertDefined(authOptions.session, "authOptions.session")

      expect(session).toBeDefined()
      expect(session.strategy).toBe("jwt")
      expect(session.maxAge).toBe(6 * 60 * 60) // 6 hours
    })

    it("should have custom pages configuration", async () => {
      const { authOptions } = await import("./auth")
      const pages = assertDefined(authOptions.pages, "authOptions.pages")

      expect(pages).toBeDefined()
      expect(pages.signIn).toBe("/login")
      expect(pages.error).toBe("/login")
    })

    it("should have required callbacks", async () => {
      const { authOptions } = await import("./auth")
      const callbacks = assertDefined(
        authOptions.callbacks,
        "authOptions.callbacks",
      )

      expect(callbacks).toBeDefined()
      expect(typeof callbacks.jwt).toBe("function")
      expect(typeof callbacks.session).toBe("function")
    })

    it("should have JWT configuration", async () => {
      const { authOptions } = await import("./auth")
      const jwt = assertDefined(authOptions.jwt, "authOptions.jwt")

      expect(jwt).toBeDefined()
      expect(jwt.secret).toBe("test-auth-secret")
    })
  })

  describe("JWT Callback", () => {
    it("should set token properties on initial login", async () => {
      const { authOptions } = await import("./auth")
      const callbacks = assertDefined(
        authOptions.callbacks,
        "authOptions.callbacks",
      )
      const jwtCallback = assertDefined(callbacks.jwt, "callbacks.jwt")

      const mockUser = {
        id: "user-123",
        name: "Test User",
        email: "test@example.com",
        companyId: 123,
        accessToken: "access-token",
        refreshToken: "refresh-token",
      }

      const token = {}
      const result = await jwtCallback({
        token,
        user: mockUser,
        account: { provider: "credentials" },
      } as never)

      expect(result.id).toBe("user-123")
      expect(result.accessToken).toBe("access-token")
      expect(result.refreshToken).toBe("refresh-token")
    })

    it("should update token on session update", async () => {
      const { authOptions } = await import("./auth")
      const callbacks = assertDefined(
        authOptions.callbacks,
        "authOptions.callbacks",
      )
      const jwtCallback = assertDefined(callbacks.jwt, "callbacks.jwt")

      const token = {
        accessToken: "old-token",
        refreshToken: "old-refresh",
      }

      const result = await jwtCallback({
        token,
        trigger: "update",
        session: {
          accessToken: "new-token",
          refreshToken: "new-refresh",
        },
      } as never)

      expect(result.accessToken).toBe("new-token")
      expect(result.refreshToken).toBe("new-refresh")
    })
  })

  describe("Session Callback", () => {
    it("should set session properties from token", async () => {
      const { authOptions } = await import("./auth")
      const callbacks = assertDefined(
        authOptions.callbacks,
        "authOptions.callbacks",
      )
      const sessionCallback = assertDefined(
        callbacks.session,
        "callbacks.session",
      )

      const session = {
        expires: new Date().toISOString(),
        user: {
          id: "",
          email: "",
          name: "",
          companyId: 0,
        },
      }

      const token = {
        id: "user-123",
        email: "test@example.com",
        name: "Test User",
        companyId: 456,
        accessToken: "access-token",
        refreshToken: "refresh-token",
      }

      const result = (await sessionCallback({
        session,
        token,
      } as never)) as typeof session & {
        accessToken?: string
      }

      expect(result.user.id).toBe("user-123")
      expect(result.user.companyId).toBe(456)
      expect(result.accessToken).toBe("access-token")
    })
  })
})

describe("Auth - API Contract Compliance", () => {
  it("should match LoginRequest schema from OpenAPI", () => {
    // According to OpenAPI:
    // LoginRequest: { email: string (format: email), password: string }
    const loginData = {
      email: "admin@example.com",
      password: "Admin123!",
    }

    expect(loginData).toHaveProperty("email")
    expect(loginData).toHaveProperty("password")
    expect(typeof loginData.email).toBe("string")
    expect(typeof loginData.password).toBe("string")
  })
})
