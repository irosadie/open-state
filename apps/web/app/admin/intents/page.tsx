import { Suspense } from "react"

import IntentsPageContent from "./intents-page-content"

export const metadata = {
  title: "Intent catalog",
  description: "Review canonical intents and their mapped workflows",
}

export default function IntentsPage() {
  return (
    <Suspense fallback={null}>
      <IntentsPageContent />
    </Suspense>
  )
}
