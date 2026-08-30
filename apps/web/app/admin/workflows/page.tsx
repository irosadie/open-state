import { Suspense } from "react"

import WorkflowsPageContent from "./workflows-page-content"

export const metadata = {
  title: "Workflow inventory",
  description: "Browse tenant workflows and open them in Builder",
}

export default function WorkflowsPage() {
  return (
    <Suspense fallback={null}>
      <WorkflowsPageContent />
    </Suspense>
  )
}
