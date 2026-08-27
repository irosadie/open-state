"use client"

import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "$/components/dropdown"
import { cn } from "$/utils/cn"
import { Loader2, MoreHorizontal } from "lucide-react"
import type * as React from "react"
import { Button } from "../button"

export interface ActionsDropdownProps {
  actions: {
    label: string
    onClick: () => void
    className?: string
    destructive?: boolean
    warning?: boolean
    loading?: boolean
  }[]
  className?: string
  buttonClassName?: string
}

export const ActionsDropdown: React.FC<ActionsDropdownProps> = ({
  actions,
  className,
  buttonClassName,
}) => {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          intent="clean"
          className={cn(
            "p-2 hover:bg-gray-100 rounded-full transition-colors outline-0",
            buttonClassName,
          )}
        >
          <MoreHorizontal className="w-4 h-4 text-gray-500" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent
        align="end"
        className={cn("min-w-30 border-gray-200", className)}
      >
        {actions.map((action) => (
          <DropdownMenuItem
            key={action.label}
            onClick={action.loading ? undefined : action.onClick}
            className={cn(
              "cursor-pointer",
              action.destructive
                ? "text-danger-500 focus:text-danger-500"
                : action.warning
                  ? "text-warning-500 focus:text-warning-500"
                  : null,
              action.loading && "cursor-not-allowed opacity-50",
            )}
          >
            {action.loading ? (
              <>
                <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                Deleting...
              </>
            ) : (
              action.label
            )}
          </DropdownMenuItem>
        ))}
      </DropdownMenuContent>
    </DropdownMenu>
  )
}

export default ActionsDropdown
