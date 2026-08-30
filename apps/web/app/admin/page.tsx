import { Suspense } from "react"

import AdminPageContent from "./admin-page-content"

export const metadata = {
  title: "Admin Console",
  description: "Tenant administration and operations console",
}

export default function AdminPage() {
  return (
    <Suspense fallback={null}>
      <AdminPageContent />
    </Suspense>
  )
}
