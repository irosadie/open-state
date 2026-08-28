import { Suspense } from "react"
import CapabilitiesPageContent from "./capabilities-page-content"

export const metadata = {
  title: "Capabilities",
  description: "Manage the capability registry, bindings, and sandbox tests",
}

export default function CapabilitiesPage() {
  return (
    <Suspense fallback={null}>
      <CapabilitiesPageContent />
    </Suspense>
  )
}
