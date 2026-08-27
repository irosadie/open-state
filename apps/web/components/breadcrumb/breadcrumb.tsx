"use client"

import Link from "next/link"
import type { FC } from "react"

type BreadcrumbDataProps = {
  title: string
  link?: string
}

export type BreadcrumbProps = {
  data: BreadcrumbDataProps[]
}

const Breadcrumb: FC<BreadcrumbProps> = (props) => {
  const { data } = props

  const renderBreadcrumb = data.map(({ link, title }, index) => (
    <li
      key={link ?? title}
      className={index !== data.length - 1 ? "underline" : ""}
    >
      <span className={index > 0 ? 'before:content-["/"]' : ""}>
        {link ? <Link href={link}>{title}</Link> : title}
      </span>
    </li>
  ))

  return <ul className="flex gap-1 text-xs">{renderBreadcrumb}</ul>
}

export default Breadcrumb
