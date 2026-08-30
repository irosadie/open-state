import { Suspense } from "react"
import RuntimeInstancesPageContent from "./runtime-instances-page-content"

export const metadata = {
  title: "Runtime Inspector",
  description: "Inspect persisted workflow runtime instances",
}

export default function RuntimeInstancesPage() {
  return (
    <Suspense fallback={null}>
      <RuntimeInstancesPageContent />
    </Suspense>
  )
}
