"use client"

import { Button } from "$/components/button"

type AccessDeniedProps = {
  title?: string
  description?: string
  onRetry?: () => void
}

export function AccessDenied({
  title = "Access denied",
  description = "You do not have permission to view this page.",
  onRetry,
}: AccessDeniedProps) {
  return (
    <main
      className="flex min-h-screen items-center justify-center px-6 py-12"
      data-testid="access-denied"
    >
      <section className="w-full max-w-md rounded-2xl border border-slate-200 bg-white p-8 text-center shadow-sm">
        <h1 className="text-xl font-semibold text-slate-900">{title}</h1>
        <p className="mt-2 text-sm text-slate-600">{description}</p>
        {onRetry ? (
          <Button className="mt-6" intent="secondary" onClick={onRetry}>
            Try again
          </Button>
        ) : null}
      </section>
    </main>
  )
}
