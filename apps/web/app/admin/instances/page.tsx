import { Suspense } from "react"

import InstancesPageContent from "./instances-page-content"

export const metadata = {
  title: "Instance operations",
  description: "Manage eligible tenant runtime instances",
}

export default function InstancesPage() {
  return (
    <Suspense fallback={null}>
      <InstancesPageContent />
    </Suspense>
  )
}
