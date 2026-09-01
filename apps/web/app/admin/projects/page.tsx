import { Suspense } from "react"

import ProjectsPageContent from "./projects-page-content"

export const metadata = {
  title: "Project inventory",
  description: "Choose the project that scopes intents, workflows, and states",
}

export default function ProjectsPage() {
  return (
    <Suspense fallback={null}>
      <ProjectsPageContent />
    </Suspense>
  )
}
