import { Suspense } from "react"

import APIKeysPageContent from "./api-keys-page-content"

export const metadata = {
  title: "State MCP API Keys",
  description: "Manage machine credentials for State MCP clients",
}

export default function APIKeysPage() {
  return (
    <Suspense fallback={null}>
      <APIKeysPageContent />
    </Suspense>
  )
}
