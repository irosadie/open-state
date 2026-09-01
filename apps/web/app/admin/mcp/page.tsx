import { Suspense } from "react"
import MCPConnectionsPageContent from "./mcp-connections-page-content"

export const metadata = {
  title: "MCP Connections",
  description: "Manage project-owned external MCP connections",
}

export default function MCPConnectionsPage() {
  return (
    <Suspense fallback={null}>
      <MCPConnectionsPageContent />
    </Suspense>
  )
}
