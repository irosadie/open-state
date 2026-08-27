"use client"

import { Avatar } from "$/components/avatar"
import { cn } from "$/utils/cn"
import { ChevronUp, LogOut } from "lucide-react"
import { signOut, useSession } from "next-auth/react"
import { useRouter } from "next/navigation"
import { useEffect, useRef, useState } from "react"

interface UserMenuProps {
  collapsed?: boolean
}

export function UserMenu({ collapsed = false }: UserMenuProps) {
  const { data: session } = useSession()
  const router = useRouter()
  const [isOpen, setIsOpen] = useState(false)
  const menuRef = useRef<HTMLDivElement>(null)

  const handleLogout = async () => {
    setIsOpen(false)
    await signOut({ redirect: false })
    router.push("/login")
    router.refresh()
  }

  const userName = session?.user?.name || "Starter User"
  const userEmail = session?.user?.email || "starter@local.dev"

  // Close menu when clicking outside
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (menuRef.current && !menuRef.current.contains(event.target as Node)) {
        setIsOpen(false)
      }
    }

    document.addEventListener("mousedown", handleClickOutside)
    return () => document.removeEventListener("mousedown", handleClickOutside)
  }, [])

  if (collapsed) {
    return (
      <div className="flex flex-col items-center gap-2 relative" ref={menuRef}>
        <button
          type="button"
          onClick={() => setIsOpen(!isOpen)}
          className="flex items-center justify-center w-10 h-10 rounded-full bg-primary-100 text-primary-700 hover:bg-primary-200 transition-colors"
        >
          <span className="text-sm font-medium">
            {userName.charAt(0).toUpperCase()}
          </span>
        </button>

        {isOpen && (
          <div className="absolute bottom-full left-1/2 -translate-x-1/2 mb-2 w-40 bg-white rounded-lg shadow-lg border border-gray-200 py-1 z-50">
            <div className="px-3 py-2 border-b border-gray-100">
              <p className="text-xs font-medium text-gray-900 truncate">
                {userName}
              </p>
              <p className="text-[10px] text-gray-500 truncate">{userEmail}</p>
            </div>
            <button
              type="button"
              onClick={handleLogout}
              className="flex items-center gap-2 w-full px-3 py-1.5 text-xs text-danger-600 hover:bg-danger-50"
            >
              <LogOut className="h-3 w-3" />
              Logout
            </button>
          </div>
        )}
      </div>
    )
  }

  return (
    <div className="relative" ref={menuRef}>
      <button
        type="button"
        onClick={() => setIsOpen(!isOpen)}
        className={cn(
          "flex items-center gap-3 w-full p-2 rounded-lg transition-colors",
          isOpen ? "bg-gray-100" : "hover:bg-gray-50",
        )}
      >
        <Avatar name={userName} className="h-8 w-8" />
        <div className="flex-1 min-w-0 text-left">
          <p className="text-sm font-medium text-gray-900 truncate">
            {userName}
          </p>
          <p className="text-xs text-gray-500 truncate">{userEmail}</p>
        </div>
        <ChevronUp
          className={cn(
            "h-4 w-4 text-gray-400 transition-transform",
            isOpen ? "rotate-180" : "",
          )}
        />
      </button>

      {isOpen && (
        <div className="absolute bottom-full left-0 right-0 mb-2 bg-white rounded-lg shadow-lg border border-gray-200 py-1 z-50">
          <div className="px-3 py-2 border-b border-gray-100">
            <p className="text-sm font-medium text-gray-900">{userName}</p>
            <p className="text-xs text-gray-500">{userEmail}</p>
          </div>
          <button
            type="button"
            onClick={handleLogout}
            className="flex items-center gap-2 w-full px-3 py-2 text-sm text-danger-600 hover:bg-danger-50"
          >
            <LogOut className="h-4 w-4" />
            Logout
          </button>
        </div>
      )}
    </div>
  )
}

export default UserMenu
