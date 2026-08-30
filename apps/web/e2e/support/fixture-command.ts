import { execFileSync } from "node:child_process"
import { resolve } from "node:path"
import { fixturePassword } from "../fixtures/golden-journeys"

type FixtureCommandResult = {
  mode: string
  checks: string[]
}

const repoRoot = resolve(process.cwd(), "../..")
const apiRoot = resolve(repoRoot, "apps/api")

export function runFixtureCommand(
  mode: string,
  extraEnv: Record<string, string> = {},
): FixtureCommandResult {
  const raw = execFileSync(
    "go",
    ["run", "./cmd/e2e-fixtures", "--mode", mode],
    {
      cwd: apiRoot,
      env: {
        ...process.env,
        ...extraEnv,
        E2E_FIXTURES: "1",
        E2E_FIXTURE_PASSWORD: fixturePassword(),
        DATABASE_URL:
          process.env.DATABASE_URL ??
          "postgresql://postgres:postgres@127.0.0.1:55437/openstate_e2e",
      },
      encoding: "utf8",
      stdio: ["ignore", "pipe", "pipe"],
    },
  )

  return JSON.parse(raw) as FixtureCommandResult
}
