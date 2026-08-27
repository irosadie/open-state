"use client"

import { cn } from "$/utils/cn"
import {
  type Cell,
  type ColumnDef,
  type Header,
  type HeaderGroup,
  type Row,
  flexRender,
  getCoreRowModel,
  useReactTable,
} from "@tanstack/react-table"
import type { VariantProps } from "class-variance-authority"
import { motion as Motion } from "framer-motion"
import { SquirrelIcon } from "lucide-react"
import type React from "react"
import { Fragment, useEffect, useState } from "react"
import { Pagination, type PaginationProps } from "../pagination"
import {
  expandedRowBorder,
  thBg,
  theadBg,
  theadRowBorder,
  trow,
} from "./table.style"

enum StageProps {
  LOADING = "LOADING",
  EMPTY = "EMPTY",
  LOADED = "LOADED",
}

const htmlToText = (value: string) => {
  if (typeof window !== "undefined" && "DOMParser" in window) {
    const document = new window.DOMParser().parseFromString(value, "text/html")

    return document.body.textContent ?? ""
  }

  return value.replace(/<[^>]*>/g, "")
}

type ColumnMeta = {
  className?: string
}

export type ColumnDefWithMeta<TData> = ColumnDef<TData> & {
  meta?: ColumnMeta
  order?: number
  visible?: boolean
  moveable?: boolean
  hideable?: boolean
}

type TableProps<T> = {
  columns: ColumnDef<T>[]
  data: T[]
  isShowPagination?: boolean
  isLoading?: boolean
  pagination?: Omit<PaginationProps, "intent" | "size">
  intent?: VariantProps<typeof theadBg>["intent"]
  wrapperClassName?: string
  className?: string
  theadClassName?: string
  theadRowClassName?: string
  tbodyClassName?: string
  trClassName?: string
  thClassName?: string
  tdClassName?: string
  paginationWrapperClassName?: string
  expandedId?: string
  renderTable?: (props: { children: React.ReactNode }) => React.ReactNode
  renderTHead?: (props: { children: React.ReactNode }) => React.ReactNode
  renderTHeadRow?: (props: {
    children: React.ReactNode
    headerGroup: HeaderGroup<T>
    index: number
  }) => React.ReactNode
  renderTRow?: (props: {
    children: React.ReactNode
    row: Row<T>
    onClick: () => void
    expanded?: boolean
  }) => React.ReactNode
  renderTH?: (props: {
    children: React.ReactNode
    header: Header<T, unknown>
    index: number
  }) => React.ReactNode
  renderTBody?: (props: { children: React.ReactNode }) => React.ReactNode
  renderTD?: (props: {
    children: React.ReactNode
    cell: Cell<T, unknown>
  }) => React.ReactNode
  renderExpandedRow?: (row: Row<T>) => React.ReactNode
  renderPagination?: (
    props: Omit<PaginationProps, "intent" | "size">,
  ) => React.ReactNode
  renderFooter?: () => React.ReactNode
  renderLoading?: () => React.ReactNode
  renderNoData?: () => React.ReactNode
  onRowClick?: (row: Row<T>) => void
}

const Table = <T,>(props: TableProps<T>) => {
  const {
    columns,
    data,
    wrapperClassName,
    className,
    isShowPagination = true,
    tdClassName,
    theadClassName,
    theadRowClassName,
    thClassName,
    tbodyClassName,
    pagination,
    isLoading,
    intent,
    expandedId,
    renderTable = ({ children }) => (
      <table className={cn("w-full", className)}>{children}</table>
    ),
    renderTHead = ({ children }) => (
      <thead className={cn(theadBg({ intent }), theadClassName)}>
        {children}
      </thead>
    ),
    renderTHeadRow = ({ children, headerGroup }) => (
      <tr
        key={headerGroup.id}
        className={cn(theadRowBorder({ intent }), theadRowClassName)}
      >
        {children}
      </tr>
    ),
    renderTRow = ({ children, onClick, row, expanded = false }) => (
      <tr
        key={row.id}
        onClick={onClick}
        onKeyDown={(event) => {
          if (event.key === "Enter" || event.key === " ") {
            event.preventDefault()
            onClick()
          }
        }}
        tabIndex={0}
        className={cn(trow({ intent, hasExpandedRow: expanded }), className)}
      >
        {children}
      </tr>
    ),
    renderTH = ({ children, header }) => {
      const meta = header.column.columnDef.meta as ColumnMeta

      return (
        <th
          key={header.column.getIndex()}
          className={cn(thBg({ intent }), thClassName, meta?.className)}
        >
          {children}
        </th>
      )
    },
    renderTBody = ({ children }) => (
      <tbody className={tbodyClassName}>{children}</tbody>
    ),
    renderTD = ({ cell, children }) => {
      const meta = cell.column.columnDef.meta as { className?: string }

      const isCellValueObject =
        typeof cell.getValue() === "object" && cell.getValue() !== null

      if (!isCellValueObject) {
        return (
          <td key={cell.id} className={cn("p-2", tdClassName, meta?.className)}>
            {children}
          </td>
        )
      }

      const value = cell.getValue() as string
      const isHtmlContent = /<[^>]*>/.test(value)

      return (
        <td key={cell.id} className={cn("p-2", tdClassName, meta?.className)}>
          {isHtmlContent ? htmlToText(value) : value}
        </td>
      )
    },
    renderExpandedRow,
    renderFooter = () => null,
    renderLoading = () => <TableLoadingContent lineNumber={4} />,
    renderNoData = () => <TableNoDataContent />,
    renderPagination = ({ limit, onPageClick, page, totalData }) => (
      <Pagination
        limit={limit}
        page={page}
        totalData={totalData}
        onPageClick={onPageClick}
        intent={"primary"}
      />
    ),
    onRowClick = () => null,
  } = props

  const {
    limit = 0,
    onPageClick = () => null,
    page = 0,
    totalData = 0,
  } = pagination || {}
  const [expandedRow, setExpandedRow] = useState<string | null>(
    expandedId ?? null,
  )

  const table = useReactTable<T>({
    data,
    columns,
    getCoreRowModel: getCoreRowModel(),
  })

  const toggleExpand = (rowId: string) => {
    onRowClick(table.getRow(rowId))
    if (renderExpandedRow) {
      setExpandedRow(expandedRow === rowId ? null : rowId)
    }
  }

  const mainContentStage = isLoading
    ? StageProps.LOADING
    : data.length === 0
      ? StageProps.EMPTY
      : StageProps.LOADED

  useEffect(() => {
    setExpandedRow(expandedId ?? null)
  }, [expandedId])

  return (
    <Fragment>
      <div className={cn("overflow-x-scroll w-full pb-4", wrapperClassName)}>
        {renderTable({
          children: (
            <Fragment>
              {renderTHead({
                children: table.getHeaderGroups().map((headerGroup, index) =>
                  renderTHeadRow({
                    index,
                    headerGroup,
                    children: headerGroup.headers.map((header, index) =>
                      renderTH({
                        index,
                        header,
                        children: flexRender(
                          header.column.columnDef.header,
                          header.getContext(),
                        ),
                      }),
                    ),
                  }),
                ),
              })}
              {
                <Fragment>
                  {
                    {
                      [StageProps.LOADING]: renderTBody({
                        children: (
                          <tr>
                            <td colSpan={columns.length}>{renderLoading()}</td>
                          </tr>
                        ),
                      }),
                      [StageProps.EMPTY]: renderTBody({
                        children: (
                          <tr>
                            <td colSpan={columns.length}>{renderNoData()}</td>
                          </tr>
                        ),
                      }),
                      [StageProps.LOADED]: renderTBody({
                        children: table.getRowModel().rows.map((row) => (
                          <Fragment key={row.id}>
                            {renderTRow({
                              row,
                              onClick: () => toggleExpand(row.id),
                              expanded:
                                expandedRow === row.id || !!renderExpandedRow,
                              children: row.getVisibleCells().map((cell) =>
                                renderTD({
                                  cell,
                                  children: flexRender(
                                    cell.column.columnDef.cell,
                                    cell.getContext(),
                                  ),
                                }),
                              ),
                            })}
                            {expandedRow === row.id && renderExpandedRow && (
                              <Fragment>
                                <tr />
                                <tr>
                                  <td
                                    colSpan={columns.length}
                                    className={cn(
                                      expandedRowBorder({ intent }),
                                    )}
                                  >
                                    <Motion.div
                                      initial={{ height: 0, opacity: 0 }}
                                      animate={{ height: "auto", opacity: 1 }}
                                      exit={{ height: 0, opacity: 0 }}
                                      transition={{
                                        duration: 0.3,
                                        ease: "easeInOut",
                                      }}
                                    >
                                      {renderExpandedRow(row)}
                                    </Motion.div>
                                  </td>
                                </tr>
                              </Fragment>
                            )}
                          </Fragment>
                        )),
                      }),
                    }[mainContentStage]
                  }
                </Fragment>
              }
              {renderFooter()}
            </Fragment>
          ),
        })}
      </div>
      {isShowPagination && totalData >= 1 && (
        <div className="flex justify-end items-center my-4">
          {renderPagination({ limit, onPageClick, page, totalData })}
        </div>
      )}
    </Fragment>
  )
}

type TableLoadingContentProps = {
  lineNumber?: number
}

const TableLoadingContent = (props: TableLoadingContentProps) => {
  const { lineNumber = 4 } = props
  const weightSkeleton = [
    "max-w-[30%]",
    "max-w-[90%]",
    "max-w-[60%]",
    "max-w-[80%]",
  ]
  const skeletonLines = Array.from({ length: lineNumber }, (_, index) => {
    const randomChooseIndex = Math.floor(Math.random() * weightSkeleton.length)

    return {
      id: `line-${index + 1}`,
      className: weightSkeleton[index] ?? weightSkeleton[randomChooseIndex],
    }
  })

  return (
    <div className="gap-3 flex flex-col my-4">
      {skeletonLines.map((line) => (
        <div
          key={line.id}
          className={cn(
            "animate-pulse bg-main-100 h-6 rounded",
            line.className,
          )}
        />
      ))}
    </div>
  )
}

const TableNoDataContent = () => {
  return (
    <div className="w-full flex flex-col items-center justify-center py-20">
      <SquirrelIcon className="h-20 w-20 text-main-200" />
      <div className="text-center text-main-500 text-sm">
        There is no data yet!
      </div>
    </div>
  )
}

export type { TableProps, ColumnDef }
export default Table
