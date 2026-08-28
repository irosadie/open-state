import { Suspense } from "react"
import AuditPageContent from "./audit-page-content"

export const metadata = {
  title: "Audit Log",
  description: "Browse the tenant audit trail (PRD 50)",
}

export default function AuditPage() {
  return (
    <Suspense fallback={null}>
      <AuditPageContent />
    </Suspense>
  )
}
