import { expect, test } from "@playwright/test"
import { GOLDEN_IDENTITIES, GOLDEN_IDS } from "./fixtures/golden-journeys"
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

  const intentResponse = await page.evaluate(async () => {
    const response = await fetch("/api/proxy/intents", {
      headers: {
        "X-Tenant-ID": "00000000-0000-0000-0000-0000000000a1",
      },
    })
    return {
      status: response.status,
      payload: (await response.json()) as {
        data?: Array<{
          key: string
          examples: string[]
          workflowId: string
        }>
      },
    }
  })
  expect(intentResponse.status).toBe(200)
  expect(intentResponse.payload.data).toEqual(
    expect.arrayContaining([
      expect.objectContaining({
        key: "BOOKING_PADEL",
        examples: expect.arrayContaining(["saya mau order lapangan"]),
        workflowId: GOLDEN_IDS.builderWorkflow,
      }),
    ]),
  )

  const apiKeyResponse = await page.evaluate(async () => {
    const response = await fetch("/api/proxy/api-keys", {
      headers: {
        "X-Tenant-ID": "00000000-0000-0000-0000-0000000000a1",
      },
    })
    return response.status
  })
  expect(apiKeyResponse).toBe(403)

  await page.goto("/admin/intents")
  await expect(page.getByText("BOOKING_PADEL")).toBeVisible()
  await expect(page.getByText(/saya mau order lapangan/)).toBeVisible()
  await expect(
    page.getByRole("link", { name: "Open Builder" }),
  ).toHaveAttribute("href", `/state-builder/${GOLDEN_IDS.builderWorkflow}`)
  expect(runFixtureCommand("verify").checks).toContain("safe JSON fields")
  expect(violations()).toEqual([])
})
