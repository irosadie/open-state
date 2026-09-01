const stateUrl = process.env.STATE_MCP_URL ?? "http://127.0.0.1:8030"
const providerUrl = process.env.PROVIDER_MCP_URL ?? "http://127.0.0.1:8031"

async function readReady(url) {
  const response = await fetch(`${url}/health/ready`)
  const body = await response.json()
  if (!response.ok) throw new Error(`${url} returned HTTP ${response.status}`)
  return body
}

try {
  const state = await readReady(stateUrl)
  if (state.server !== "openstate") {
    throw new Error(`State MCP identity mismatch: expected openstate, got ${state.server ?? "unknown"}`)
  }

  const provider = await readReady(providerUrl)
  if (typeof provider.provider !== "string" || provider.provider.length === 0) {
    throw new Error("Provider MCP identity is missing")
  }

  console.log(`MCP runtime OK: State MCP (${state.server}) + Provider MCP (${provider.provider})`)
} catch (error) {
  console.error(`MCP runtime check failed: ${error instanceof Error ? error.message : String(error)}`)
  process.exitCode = 1
}
