import type { FullConfig } from "@playwright/test"
import { runFixtureCommand } from "./fixture-command"
import { validateGoldenManifests } from "./fixture-safety"

export default function globalSetup(_config: FullConfig): void {
  validateGoldenManifests()
  runFixtureCommand("verify")
}
