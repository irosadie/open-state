import { AuthorizationBoundary } from "$/components/auth-guard/authorization-boundary"
import { AuthorizationProvider } from "$/providers/authorization-provider"
import QueryProvider from "$/providers/query-provider"
import SessionProviderWrapper from "$/providers/session-provider"
import type { Metadata } from "next"
import "./globals.css"

export const metadata: Metadata = {
  title: "OpenState",
  description:
    "OpenState — Enterprise Conversation State Orchestration Platform",
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
          <QueryProvider>
            <AuthorizationProvider>
              <AuthorizationBoundary>{children}</AuthorizationBoundary>
            </AuthorizationProvider>
          </QueryProvider>
        </SessionProviderWrapper>
      </body>
    </html>
  )
}
