"use client"

import { cn } from "$/utils/cn"
import type { LucideIcon } from "lucide-react"
import type React from "react"
import { Button } from "../button"

interface EmptyStateProps {
  icon?: LucideIcon
  title: string
  description?: string
  actionButton?: {
    label: string
    onClick: () => void
    icon?: React.ReactElement
  }
  className?: string
}

export const EmptyState: React.FC<EmptyStateProps> = ({
  icon: Icon,
  title,
  description,
  actionButton,
  className,
}) => {
  return (
    <div className={cn("text-center py-12", className)}>
      {Icon && <Icon className="w-12 h-12 text-gray-300 mx-auto mb-4" />}
      <h3 className="text-lg font-medium text-gray-900 mb-1">{title}</h3>
      {description && (
        <p className="text-sm text-gray-500 mb-4">{description}</p>
      )}
      {actionButton && (
        <div className="inline-flex">
          <Button
            intent="primary"
            size="medium"
            textOnly
            leftIcon={actionButton.icon}
            onClick={actionButton.onClick}
            rounded={"large"}
          >
            {actionButton.label}
          </Button>
        </div>
      )}
    </div>
  )
}

export default EmptyState
