"use client"

import { cn } from "$/utils/cn"
import { TrendingDown, TrendingUp } from "lucide-react"
import type React from "react"
import type { FC } from "react"

export type StatCardProps = {
  title: string
  value: string | number
  change?: string
  trend?: "up" | "down"
  icon?: React.ReactNode
  iconColor?: string
  className?: string
}

const StatCard: FC<StatCardProps> = (props) => {
  const {
    title,
    value,
    change,
    trend,
    icon,
    iconColor = "bg-gray-100",
    className,
  } = props

  return (
    <div
      className={cn(
        "rounded-xl border border-gray-200 bg-white p-6 shadow-sm transition-shadow hover:shadow-md",
        className,
      )}
    >
      <div className="flex items-center justify-between">
        <div
          className={cn(
            "flex h-12 w-12 items-center justify-center rounded-lg",
            iconColor,
          )}
        >
          {icon}
        </div>
        {change && (
          <div
            className={cn(
              "flex items-center gap-1 rounded-full px-2 py-1 text-xs font-medium",
              trend === "up"
                ? "bg-success-50 text-success-600"
                : "bg-danger-50 text-danger-600",
            )}
          >
            {trend === "up" ? (
              <TrendingUp className="h-3 w-3" />
            ) : (
              <TrendingDown className="h-3 w-3" />
            )}
            {change}
          </div>
        )}
      </div>
      <div className="mt-4">
        <p className="text-sm text-gray-500">{title}</p>
        <p className="text-2xl font-bold text-gray-900 mt-1">{value}</p>
      </div>
    </div>
  )
}

export default StatCard
