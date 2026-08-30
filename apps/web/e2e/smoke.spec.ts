import { expect, test } from "@playwright/test"
import { GOLDEN_IDENTITIES } from "./fixtures/golden-journeys"
import { signInAs } from "./support/auth"
import { runFixtureCommand } from "./support/fixture-command"
import { installLocalNetworkGuard } from "./support/network-guard"

test("starts a browser, forwards authenticated BFF traffic, and verifies fixtures", async ({
  page,
}) => {
  const violations = installLocalNetworkGuard(page)
  let currentUserForwarded = false
  page.on("request", (request) => {
    if (request.url().includes("/api/proxy/auth/me")) {
      currentUserForwarded = true
    }
  })

  await signInAs(page, GOLDEN_IDENTITIES.editor)
  expect(currentUserForwarded).toBe(true)
  expect(runFixtureCommand("verify").checks).toContain("safe JSON fields")
  expect(violations()).toEqual([])
})
