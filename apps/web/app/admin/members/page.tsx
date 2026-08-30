import { Suspense } from "react"

import MembersPageContent from "./members-page-content"

export const metadata = {
  title: "Members & roles",
  description: "Manage tenant membership and roles",
}

export default function MembersPage() {
  return (
    <Suspense fallback={null}>
      <MembersPageContent />
    </Suspense>
  )
}
