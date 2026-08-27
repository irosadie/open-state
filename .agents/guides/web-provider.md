# Guide: Web Provider (`apps/web/providers/`)

## Folder Contract

✅ Allowed:
- React Context provider wrappers
- Global state that needs to be shared between components (theme, auth session, toast, etc.)
- Wrap library providers (ReactQueryProvider, ThemeProvider, etc.)

❌ Forbidden:
- Business logic
- Direct API calls
- UI rendering unrelated to providing context

---

## Conventions

### Root Provider

```tsx
// providers/app-providers.tsx
'use client'

import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { SessionProvider } from 'next-auth/react'
import { Toaster } from 'react-hot-toast'
import { useState } from 'react'
import type { ReactNode } from 'react'

interface AppProvidersProps {
  children: ReactNode
}

export function AppProviders({ children }: AppProvidersProps) {
  const [queryClient] = useState(() => new QueryClient({
    defaultOptions: {
      queries: { retry: 1, staleTime: 30_000 },
    },
  }))

  return (
    <SessionProvider>
      <QueryClientProvider client={queryClient}>
        {children}
        <Toaster position="top-right" />
      </QueryClientProvider>
    </SessionProvider>
  )
}
```

### Custom Context Provider

```tsx
// providers/sidebar-provider.tsx
'use client'

import { createContext, useContext, useState } from 'react'
import type { ReactNode } from 'react'

interface SidebarContextValue {
  isOpen: boolean
  toggle: () => void
}

const SidebarContext = createContext<SidebarContextValue | null>(null)

export function SidebarProvider({ children }: { children: ReactNode }) {
  const [isOpen, setIsOpen] = useState(true)
  return (
    <SidebarContext.Provider value={{ isOpen, toggle: () => setIsOpen((v) => !v) }}>
      {children}
    </SidebarContext.Provider>
  )
}

export function useSidebar() {
  const ctx = useContext(SidebarContext)
  if (!ctx) throw new Error('useSidebar must be used within SidebarProvider')
  return ctx
}
```

---

## Additional Rules

- Root provider (`app-providers.tsx`) is imported in `app/layout.tsx`
- Custom context hook (`useSidebar`) always exports alongside its provider
- Files must end with newline
