import { cleanup } from "@testing-library/react"
import { afterEach, vi } from "vitest"

afterEach(() => {
  cleanup()
})

// Mock environment variables
vi.mock("$/configs/env", () => ({
  env: {
    appBaseUrl: "http://localhost:3020",
    apiBaseUrl: "/api/proxy",
    authBaseUrl: "/api/auth",
    defaultAuthRedirectPath: "/",
  },
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

// Setup jsdom globals if not available
if (typeof document !== "undefined") {
  // Already in browser/jsdom
} else if (typeof global !== "undefined") {
  // Node.js environment - jsdom will be available
}

// Mock NextAuth
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

// Mock react-hot-toast
vi.mock("react-hot-toast", () => ({
  toast: {
    success: vi.fn(),
    error: vi.fn(),
    loading: vi.fn(),
  },
  Toaster: vi.fn(),
}))

// Setup global test utilities
global.ResizeObserver = vi.fn().mockImplementation(() => ({
  observe: vi.fn(),
  unobserve: vi.fn(),
  disconnect: vi.fn(),
}))

global.MutationObserver = vi.fn().mockImplementation(() => ({
  observe: vi.fn(),
  unobserve: vi.fn(),
  disconnect: vi.fn(),
}))
