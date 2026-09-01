import { startProviderMock } from "./app"

const defaultScenarioPath = new URL("../fixtures/padel.json", import.meta.url)
  .pathname
const port = Number.parseInt(process.env.MCP_PROVIDER_MOCK_PORT ?? "8031", 10)
const scenarioPath =
  process.env.MCP_PROVIDER_MOCK_SCENARIO ?? defaultScenarioPath

if (!Number.isInteger(port) || port < 1 || port > 65535) {
  process.stderr.write("MCP_PROVIDER_MOCK_PORT must be a valid port number\n")
  process.exitCode = 1
} else {
  startProviderMock({ port, scenarioPath })
    .then((app) => {
      process.stdout.write(`${app.providerName} listening on ${app.url}/mcp\n`)
    })
    .catch((error) => {
      const message =
        error instanceof Error ? error.message : "unable to start provider mock"
      process.stderr.write(`provider mock startup failed: ${message}\n`)
      process.exitCode = 1
    })
}
