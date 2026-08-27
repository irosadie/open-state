"use client"

import { Fragment } from "react"
import { Button, type ButtonProps } from "../button"

export type PaginationProps = {
  limit: number
  page: number
  totalData: number
  onPageClick: (page: number) => void
  intent?: ButtonProps["intent"]
  size?: ButtonProps["size"]
}

const Pagination = (props: PaginationProps) => {
  const {
    limit,
    page = 1,
    totalData,
    onPageClick,
    intent = "primary",
    size = "small",
  } = props

  const totalPages = Math.ceil(totalData / limit)

  const isDisabledPrevious = page <= 1
  const isDisabledNext = page >= totalPages

  const renderPageButtons = () => {
    if (totalPages <= 8) {
      return Array.from({ length: totalPages }, (_, index) => index + 1)
    }
    let pages: number[] = []

    if (page <= 4) {
      pages = Array.from({ length: 6 }, (_, index) => index + 1)
    } else if (page >= totalPages - 3) {
      pages = Array.from({ length: 6 }, (_, index) => totalPages - 5 + index)
    } else {
      pages = Array.from({ length: 6 }, (_, index) => page - 2 + index)
    }
    if (!pages.includes(1)) {
      pages.unshift(1, 2)
    }
    if (!pages.includes(totalPages)) {
      pages.push(totalPages - 1, totalPages)
    }
    return pages
  }

  return (
    <div className="flex gap-2">
      <Button
        disabled={isDisabledPrevious}
        onClick={(e) => {
          if (!isDisabledPrevious) {
            onPageClick(page - 1)
            return
          }
          e.preventDefault()
        }}
        intent={intent}
        size={size}
        bordered
      >
        Previous
      </Button>
      <div className="flex items-center gap-2">
        {renderPageButtons().map((pageNumber) => {
          const isNumberDisabled = page === pageNumber

          return (
            <Fragment key={pageNumber}>
              <Button
                key={pageNumber}
                intent={intent}
                textOnly={!isNumberDisabled}
                size={size}
                onClick={() => onPageClick(pageNumber)}
                disabled={isNumberDisabled}
                className="opacity-100 transition-all duration-200"
              >
                {pageNumber}
              </Button>
            </Fragment>
          )
        })}
      </div>
      <Button
        className="capitalize"
        disabled={isDisabledNext}
        onClick={(e) => {
          if (!isDisabledNext) {
            onPageClick(page + 1)
            return
          }
          e.preventDefault()
        }}
        size={size}
        intent={intent}
        bordered
      >
        Next
      </Button>
    </div>
  )
}

export default Pagination
