"use client"

import { cn } from "$/utils/cn"
import type { FC, ReactNode } from "react"

export type PanelCardProps = {
  title?: string
  description?: string
  children: ReactNode
  action?: ReactNode
  className?: string
  noPadding?: boolean
}

const PanelCard: FC<PanelCardProps> = (props) => {
  const {
    title,
    description,
    children,
    action,
    className,
    noPadding = false,
  } = props

  return (
    <div
      className={cn(
        "rounded-xl border border-gray-200 bg-white shadow-xs",
        className,
      )}
    >
      {(title || action) && (
        <div className="flex items-center justify-between border-b border-gray-200 px-6 py-4">
          <div>
            {title && (
              <h2 className="text-lg font-semibold text-gray-900">{title}</h2>
            )}
            {description && (
              <p className="text-sm text-gray-500 mt-0.5">{description}</p>
            )}
          </div>
          {action && <div>{action}</div>}
        </div>
      )}
      <div className={cn(!noPadding && "p-6")}>{children}</div>
    </div>
  )
}

export default PanelCard
