import { Suspense } from "react"
import LoginContent from "./login-content"

export default function LoginPage() {
  return (
    <Suspense
      fallback={
        <main className="mx-auto flex min-h-screen w-full max-w-md items-center px-6 py-12 text-sm text-slate-600">
          Loading sign in...
        </main>
      }
    >
      <LoginContent />
    </Suspense>
  )
}
