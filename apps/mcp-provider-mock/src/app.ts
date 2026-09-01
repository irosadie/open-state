import { type ProviderMock, createProviderMock } from "./provider-mock"
import { type ProviderScenario, loadScenario } from "./scenario"

export type ProviderMockApp = {
  providerName: string
  server: ReturnType<typeof Bun.serve>
  scenario: ProviderScenario
  stop: () => void
  url: string
}

export async function startProviderMock(options: {
  port?: number
  scenarioPath: string
}): Promise<ProviderMockApp> {
  const scenario = await loadScenario(options.scenarioPath)
  const provider = await createProviderMock(scenario)
  const server = Bun.serve({
    fetch(request) {
      return routeRequest(request, provider)
    },
    port: options.port ?? 8031,
  })
  const url = `http://127.0.0.1:${server.port}`

  return {
    providerName: provider.providerName,
    scenario,
    server,
    stop: () => server.stop(true),
    url,
  }
}

function routeRequest(
  request: Request,
  provider: ProviderMock,
): Promise<Response> | Response {
  const path = new URL(request.url).pathname

  if (path === "/health/live") {
    return Response.json({ status: "ok" })
  }

  if (path === "/health/ready") {
    return Response.json({ status: "ready", provider: provider.providerName })
  }

  if (path === "/mcp") {
    return provider.handleMcpRequest(request)
  }

  return Response.json({ message: "not found" }, { status: 404 })
}
