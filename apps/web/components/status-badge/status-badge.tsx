"use client"

import { cn } from "$/utils/cn"
import type React from "react"

export type StatusBadgeVariant =
  | "success"
  | "warning"
  | "danger"
  | "info"
  | "neutral"
  | "primary"

interface StatusBadgeProps {
  status?: boolean | string
  activeLabel?: string
  inactiveLabel?: string
  activeClassName?: string
  inactiveClassName?: string
  className?: string
  variant?: StatusBadgeVariant
  children?: React.ReactNode
}

export const StatusBadge: React.FC<StatusBadgeProps> = ({
  status,
  activeLabel = "Active",
  inactiveLabel = "Inactive",
  activeClassName = "bg-success-100 text-success-700",
  inactiveClassName = "bg-gray-100 text-gray-500",
  className,
  variant,
  children,
}) => {
  if (variant) {
    const variantStyles: Record<StatusBadgeVariant, string> = {
      success: "bg-green-100 text-green-800",
      warning: "bg-yellow-100 text-yellow-800",
      danger: "bg-red-100 text-red-800",
      info: "bg-blue-100 text-blue-800",
      neutral: "bg-gray-100 text-gray-800",
      primary: "bg-primary-100 text-primary-800",
    }

    return (
      <span
        className={cn(
          "inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium",
          variantStyles[variant],
          className,
        )}
      >
        {children || (typeof status === "string" ? status : null)}
      </span>
    )
  }

  const isActive =
    typeof status === "boolean"
      ? status
      : status === "active" || status === "true"

  return (
    <span
      className={cn(
        "inline-flex rounded-full px-2.5 py-0.5 text-xs font-medium",
        isActive ? activeClassName : inactiveClassName,
        className,
      )}
    >
      {isActive ? activeLabel : inactiveLabel}
    </span>
  )
}

export default StatusBadge
