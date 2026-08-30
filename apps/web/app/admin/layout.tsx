import type { ReactNode } from "react"

import { AdminConsoleShell } from "$/components/admin-console"

export default function AdminLayout({ children }: { children: ReactNode }) {
  return <AdminConsoleShell>{children}</AdminConsoleShell>
}
