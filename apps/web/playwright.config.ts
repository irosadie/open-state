import { resolve } from "node:path"
import { defineConfig, devices } from "@playwright/test"

const webPort = Number(process.env.E2E_WEB_PORT ?? 3020)
const apiPort = Number(process.env.E2E_API_PORT ?? 8021)
const baseURL = process.env.E2E_BASE_URL ?? `http://127.0.0.1:${webPort}`
const apiURL = process.env.E2E_API_URL ?? `http://127.0.0.1:${apiPort}`
const repoRoot = resolve(process.cwd(), "../..")
const databaseURL =
  process.env.DATABASE_URL ??
  "postgresql://postgres:postgres@127.0.0.1:55437/openstate_e2e"
const testSecret =
  process.env.E2E_TEST_SECRET ?? "openstate-e2e-only-secret-0123456789-abcdef"

const serverEnv = {
  ...process.env,
  NODE_ENV: "test",
  DATABASE_URL: databaseURL,
  JWT_SECRET: testSecret,
  MCP_API_KEY_PEPPER: testSecret,
  NEXTAUTH_SECRET: testSecret,
  NEXTAUTH_URL: baseURL,
  NEXT_PUBLIC_APP_URL: baseURL,
  NEXT_PUBLIC_API_URL: apiURL,
  NEXT_PUBLIC_TENANT_ID:
    process.env.NEXT_PUBLIC_TENANT_ID ?? "00000000-0000-0000-0000-0000000000a1",
  NEXT_DIST_DIR: ".next-e2e",
  API_URL: apiURL,
  PORT: String(apiPort),
  METRICS_ENABLED: "false",
}

export default defineConfig({
  testDir: resolve(process.cwd(), "e2e"),
  globalSetup: resolve(process.cwd(), "e2e/support/global-setup.ts"),
  outputDir: resolve(process.cwd(), "test-results/e2e"),
  fullyParallel: false,
  forbidOnly: Boolean(process.env.CI),
  retries: process.env.CI ? 1 : 0,
  workers: 1,
  timeout: 60_000,
  expect: { timeout: 10_000 },
  reporter: [
    ["list"],
    ["json", { outputFile: "test-results/e2e/results.json" }],
  ],
  use: {
    baseURL,
    screenshot: "only-on-failure",
    video: "retain-on-failure",
    trace: "retain-on-failure",
    actionTimeout: 10_000,
    navigationTimeout: 30_000,
  },
  projects: [{ name: "chromium", use: { ...devices["Desktop Chrome"] } }],
  webServer: [
    {
      command: "exec go run ./cmd/server/main.go",
      cwd: resolve(repoRoot, "apps/api"),
      url: `${apiURL}/health`,
      timeout: 120_000,
      reuseExistingServer: !process.env.CI,
      env: serverEnv,
    },
    {
      command: `exec bunx next dev --port ${webPort}`,
      cwd: resolve(repoRoot, "apps/web"),
      url: baseURL,
      timeout: 120_000,
      reuseExistingServer: !process.env.CI,
      env: serverEnv,
    },
  ],
})
