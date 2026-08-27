"use client"

import type React from "react"
import type { FC } from "react"
import { Breadcrumb } from "../breadcrumb"
import type { BreadcrumbProps } from "../breadcrumb/breadcrumb"

export type ContentTitleProps = {
  title: string
  icon?: React.ReactNode
  breadcrumbData?: BreadcrumbProps["data"]
}

const ContentTitle: FC<ContentTitleProps> = (props) => {
  const { title, icon, breadcrumbData } = props

  return (
    <div className="flex justify-between">
      <span className="flex gap-3 items-start">
        {icon ?? null}
        <h2 className="text-3xl font-semibold">{title}</h2>
      </span>
      {breadcrumbData ? (
        <div className="flex items-center">
          <Breadcrumb data={breadcrumbData} />
        </div>
      ) : null}
    </div>
  )
}

export default ContentTitle
