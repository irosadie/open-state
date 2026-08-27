import QueryProvider from "$/providers/query-provider"
import SessionProviderWrapper from "$/providers/session-provider"
import type { Metadata } from "next"
import "./globals.css"

export const metadata: Metadata = {
  title: "vibecoding-starter",
  description:
    "Monorepo boilerplate with Next.js, Hono, worker, and agent flow",
}

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode
}>) {
  return (
    <html lang="en">
      <body>
        <SessionProviderWrapper>
          <QueryProvider>{children}</QueryProvider>
        </SessionProviderWrapper>
      </body>
    </html>
  )
}
