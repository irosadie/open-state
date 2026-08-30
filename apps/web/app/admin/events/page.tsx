import { Suspense } from "react"

import EventsPageContent from "./events-page-content"

export const metadata = {
  title: "Event browser",
  description: "Browse immutable tenant events",
}

export default function EventsPage() {
  return (
    <Suspense fallback={null}>
      <EventsPageContent />
    </Suspense>
  )
}
