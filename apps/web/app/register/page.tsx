import { Suspense } from "react"
import RegisterContent from "./register-content"

export default function RegisterPage() {
  return (
    <Suspense
      fallback={
        <main className="mx-auto flex min-h-screen w-full max-w-md items-center px-6 py-12 text-sm text-slate-600">
          Loading register form...
        </main>
      }
    >
      <RegisterContent />
    </Suspense>
  )
}
