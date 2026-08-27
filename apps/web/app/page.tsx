import { Suspense } from "react"
import HomeContent from "./home-content"

export default function HomePage() {
  return (
    <Suspense
      fallback={
        <main className="mx-auto min-h-screen w-full max-w-5xl px-6 py-12 text-sm text-slate-600">
          Loading boilerplate overview...
        </main>
      }
    >
      <HomeContent />
    </Suspense>
  )
}
