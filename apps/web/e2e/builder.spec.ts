import { expect, test } from "@playwright/test"
import { GOLDEN_IDENTITIES, GOLDEN_IDS } from "./fixtures/golden-journeys"
import { signInAs } from "./support/auth"
import { runFixtureCommand } from "./support/fixture-command"
import { installLocalNetworkGuard } from "./support/network-guard"

test.beforeEach(() => {
  runFixtureCommand("seed")
})

test("Editor saves, reloads, publishes, and compares a golden Builder graph", async ({
  page,
}) => {
  const violations = installLocalNetworkGuard(page)
  await signInAs(
    page,
    GOLDEN_IDENTITIES.editor,
    `/state-builder/${GOLDEN_IDS.builderWorkflow}`,
  )

  await expect(page.getByTestId("builder-root")).toBeVisible()
  await page
    .getByTestId("workflow-node-42000000-0000-0000-0000-000000000002")
    .click()
  await page.getByTestId("workflow-node-name").fill("INTAKE_EDITED")
  await page.getByTestId("builder-save").click()
  await expect(page.getByTestId("builder-save-status")).toContainText(
    "Tersimpan",
  )
  expect(runFixtureCommand("verify-builder-draft").checks).toContain(
    "builder draft persisted",
  )

  await page.reload()
  await expect(
    page.getByTestId("workflow-node-42000000-0000-0000-0000-000000000002"),
  ).toContainText("INTAKE_EDITED")
  await page.getByTestId("builder-publish").click()
  await expect(page.getByText(/Workflow berhasil dipublish/i)).toBeVisible()
  expect(runFixtureCommand("verify-builder-published").checks).toContain(
    "builder publish persisted",
  )

  await expect(page.getByTestId("version-history-panel")).toBeVisible()
  await expect(
    page.getByTestId("version-list").locator("text=v3"),
  ).toBeVisible()
  await expect(page.getByTestId("version-diff")).toContainText("name")
  expect(violations()).toEqual([])
})

test("prevents invalid publish without sending a publish mutation", async ({
  page,
}) => {
  await signInAs(
    page,
    GOLDEN_IDENTITIES.editor,
    `/state-builder/${GOLDEN_IDS.invalidWorkflow}`,
  )
  let publishRequests = 0
  page.on("request", (request) => {
    if (
      request.method() === "POST" &&
      request
        .url()
        .includes(`/api/proxy/workflows/${GOLDEN_IDS.invalidWorkflow}/publish`)
    ) {
      publishRequests += 1
    }
  })

  await expect(page.getByTestId("builder-validation")).toBeVisible()
  await page.getByTestId("builder-publish").click()
  await expect(page.getByText(/belum valid/i)).toBeVisible()
  expect(publishRequests).toBe(0)
})

test("keeps the local graph visible after a stale save conflict", async ({
  page,
}) => {
  await signInAs(
    page,
    GOLDEN_IDENTITIES.editor,
    `/state-builder/${GOLDEN_IDS.builderWorkflow}`,
  )
  await page
    .getByTestId("workflow-node-42000000-0000-0000-0000-000000000002")
    .click()
  await page.getByTestId("workflow-node-name").fill("LOCAL_EDIT")
  runFixtureCommand("bump-builder-version")
  await page.getByTestId("builder-save").click()
  await expect(page.getByTestId("builder-save-status")).toContainText("Konflik")
  await expect(
    page.getByTestId("workflow-node-42000000-0000-0000-0000-000000000002"),
  ).toContainText("LOCAL_EDIT")
})
