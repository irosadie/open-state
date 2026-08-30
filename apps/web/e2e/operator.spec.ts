import { expect, test } from "@playwright/test"
import { GOLDEN_IDENTITIES, GOLDEN_IDS } from "./fixtures/golden-journeys"
import { signInAs } from "./support/auth"
import { runFixtureCommand } from "./support/fixture-command"
import { installLocalNetworkGuard } from "./support/network-guard"

test.beforeEach(() => {
  runFixtureCommand("seed")
})

test("Operator discovers only tenant-A runtime instances and filters them", async ({
  page,
}) => {
  const violations = installLocalNetworkGuard(page)
  await signInAs(page, GOLDEN_IDENTITIES.operator, "/admin/runtime-instances")

  const rows = page.locator('[data-testid^="runtime-instance-row-"]')
  await expect(page.getByTestId("runtime-inspector-root")).toBeVisible()
  await expect(rows).toHaveCount(3)
  await expect(
    page.getByTestId(`runtime-instance-row-${GOLDEN_IDS.sentinelInstance}`),
  ).toHaveCount(0)

  await page.getByTestId("runtime-status-filter").selectOption("RUNNING")
  await expect(rows).toHaveCount(1)
  await expect(
    page.getByTestId(`runtime-instance-row-${GOLDEN_IDS.runningInstance}`),
  ).toBeVisible()

  await page.getByTestId("runtime-status-filter").selectOption("")
  await page.getByTestId("runtime-correlation-filter").fill("golden-suspended")
  await expect(rows).toHaveCount(1)
  await expect(
    page.getByTestId(`runtime-instance-row-${GOLDEN_IDS.suspendedInstance}`),
  ).toBeVisible()
  expect(violations()).toEqual([])
})

test("Operator sees safe runtime detail and sanitized Debug View metadata", async ({
  page,
}) => {
  const violations = installLocalNetworkGuard(page)
  await signInAs(page, GOLDEN_IDENTITIES.operator)
  await page.goto(`/admin/runtime-instances/${GOLDEN_IDS.runningInstance}`)

  await expect(page.getByTestId("runtime-instance-detail")).toBeVisible()
  await expect(page.getByTestId("runtime-workflow-summary")).toContainText(
    "golden-runtime",
  )
  await expect(page.getByTestId("runtime-workflow-summary")).toContainText("v1")
  await expect(page.getByTestId("runtime-current-state")).toContainText("START")
  await expect(page.getByTestId("runtime-context")).toContainText(
    "SYNTHETIC-001",
  )

  const timeline = page.getByTestId("runtime-timeline").locator("li")
  await expect(timeline).toHaveCount(2)
  await expect(page.getByTestId("runtime-timeline")).toContainText(
    "fixture.started",
  )

  const debug = page.getByTestId("runtime-debug-view")
  await expect(debug).toContainText("local-stub")
  await expect(debug).toContainText("MCP_ACTIVITY")
  await expect(debug).toContainText("SUCCEEDED")
  await expect(debug).not.toContainText(
    /raw_prompt|raw_response|rag_document|authorization|credential/i,
  )
  expect(violations()).toEqual([])
})

test("Operator confirms suspend, resume, and retry and verifies audits", async ({
  page,
}) => {
  const violations = installLocalNetworkGuard(page)
  const commands = [
    {
      id: GOLDEN_IDS.runningInstance,
      button: "runtime-command-suspend",
      result: "SUSPENDED: SUSPENDED",
      status: "SUSPENDED",
      verification: "verify-runtime-suspend",
    },
    {
      id: GOLDEN_IDS.suspendedInstance,
      button: "runtime-command-resume",
      result: "RESUMED: RUNNING",
      status: "RUNNING",
      verification: "verify-runtime-resume",
    },
    {
      id: GOLDEN_IDS.failedInstance,
      button: "runtime-command-retry",
      result: "RETRIED: RUNNING",
      status: "RUNNING",
      verification: "verify-runtime-retry",
    },
  ] as const

  await signInAs(page, GOLDEN_IDENTITIES.operator)
  for (const command of commands) {
    await page.goto(`/admin/runtime-instances/${command.id}`)
    await expect(page.getByTestId(command.button)).toBeVisible()
    page.once("dialog", async (dialog) => {
      await dialog.accept()
    })
    await page.getByTestId(command.button).click()
    await expect(page.getByTestId("runtime-command-result")).toHaveText(
      command.result,
    )
    await expect(page.getByTestId("runtime-lifecycle-status")).toContainText(
      command.status,
    )
    expect(runFixtureCommand(command.verification).checks).toContain(
      command.verification === "verify-runtime-suspend"
        ? "suspend state and audit"
        : command.verification === "verify-runtime-resume"
          ? "resume state and audit"
          : "retry state and audit",
    )
  }
  expect(violations()).toEqual([])
})

test("denies management routes without protected data", async ({ page }) => {
  const violations = installLocalNetworkGuard(page)
  const protectedRequests: string[] = []
  page.on("request", (request) => {
    if (request.url().includes("/api/proxy/admin/tenant")) {
      protectedRequests.push(request.url())
    }
  })

  await signInAs(page, GOLDEN_IDENTITIES.operator)
  await page.goto("/admin/tenant")
  await expect(page.getByTestId("access-denied")).toBeVisible()
  expect(protectedRequests).toEqual([])

  expect(violations()).toEqual([])
})

test("Viewer cannot use lifecycle commands or protected Debug View data", async ({
  page,
}) => {
  const violations = installLocalNetworkGuard(page)
  const protectedRequests: string[] = []
  page.on("request", (request) => {
    if (request.url().includes("/api/proxy/admin/instances")) {
      protectedRequests.push(request.url())
    }
  })

  await signInAs(
    page,
    GOLDEN_IDENTITIES.viewer,
    `/admin/runtime-instances/${GOLDEN_IDS.runningInstance}`,
  )
  await expect(page.getByTestId("runtime-instance-detail")).toBeVisible()
  await expect(page.getByTestId("runtime-command-suspend")).toHaveCount(0)
  await expect(page.getByTestId("runtime-debug-access-denied")).toContainText(
    "not authorized",
  )
  expect(protectedRequests).toEqual([])
  expect(violations()).toEqual([])
})

test("does not expose a sentinel tenant instance", async ({ page }) => {
  const violations = installLocalNetworkGuard(page)
  await signInAs(page, GOLDEN_IDENTITIES.operator)
  await page.goto(`/admin/runtime-instances/${GOLDEN_IDS.sentinelInstance}`)

  await expect(page.getByTestId("runtime-instance-error")).toBeVisible()
  await expect(page.getByTestId("runtime-instance-error")).not.toContainText(
    "sentinel-only",
  )
  await expect(page.getByTestId("runtime-command-suspend")).toHaveCount(0)
  expect(violations()).toEqual([])
})
