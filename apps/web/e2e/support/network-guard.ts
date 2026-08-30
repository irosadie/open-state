import type { Page } from "@playwright/test"

const LOCAL_HOSTS = new Set(["127.0.0.1", "localhost"])

export function installLocalNetworkGuard(page: Page): () => string[] {
  const violations: string[] = []
  void page.route("**/*", async (route) => {
    const url = new URL(route.request().url())
    if (LOCAL_HOSTS.has(url.hostname)) {
      await route.continue()
      return
    }

    violations.push(`${url.protocol}//${url.hostname}`)
    await route.abort("blockedbyclient")
  })
  return () => violations
}
