"use client"

import {
  type ReactNode,
  createContext,
  useCallback,
  useContext,
  useState,
} from "react"

type PanelContextType = {
  isSidebarCollapsed: boolean
  toggleSidebar: () => void
  setSidebarCollapsed: (collapsed: boolean) => void
  sidebarTitle: string
  setSidebarTitle: (title: string) => void
}

const PanelContext = createContext<PanelContextType | undefined>(undefined)

export function PanelProvider({ children }: { children: ReactNode }) {
  const [isSidebarCollapsed, setIsSidebarCollapsed] = useState(false)
  const [sidebarTitle, setSidebarTitle] = useState("Dashboard")

  const toggleSidebar = useCallback(() => {
    setIsSidebarCollapsed((prev) => !prev)
  }, [])

  const setSidebarCollapsed = useCallback((collapsed: boolean) => {
    setIsSidebarCollapsed(collapsed)
  }, [])

  return (
    <PanelContext.Provider
      value={{
        isSidebarCollapsed,
        toggleSidebar,
        setSidebarCollapsed,
        sidebarTitle,
        setSidebarTitle,
      }}
    >
      {children}
    </PanelContext.Provider>
  )
}

export function usePanel() {
  const context = useContext(PanelContext)

  if (context === undefined) {
    throw new Error("usePanel must be used within a PanelProvider")
  }
  return context
}
