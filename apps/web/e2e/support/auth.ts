import { type Page, expect } from "@playwright/test"
import type { GoldenIdentity } from "../fixtures/golden-journeys"
import { fixturePassword } from "../fixtures/golden-journeys"

export async function signInAs(
  page: Page,
  identity: GoldenIdentity,
  callbackPath = "/",
): Promise<void> {
  await page.goto(`/login?callbackUrl=${encodeURIComponent(callbackPath)}`)
  await page.getByTestId("login-email").fill(identity.email)
  await page.getByTestId("login-password").fill(fixturePassword())
  await page.getByTestId("login-submit").click()
  await expect(page).toHaveURL(
    new RegExp(`${callbackPath.replace("/", "\\/")}$`),
  )
}
