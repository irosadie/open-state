import { Suspense } from "react"

import TenantPageContent from "./tenant-page-content"

export const metadata = {
  title: "Tenant settings",
  description: "Manage the current tenant profile",
}

export default function TenantPage() {
  return (
    <Suspense fallback={null}>
      <TenantPageContent />
    </Suspense>
  )
}
