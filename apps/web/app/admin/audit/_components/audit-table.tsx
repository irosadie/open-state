"use client"

import { useMemo } from "react"

import { EmptyState } from "$/components/empty-state"
import { type ColumnDef, Table } from "$/components/table"
import { getAuditActionLabel } from "@openstate/schemas"
import type { AuditEntryResponse } from "@openstate/types"
import { ScrollTextIcon } from "lucide-react"

type AuditTableProps = {
  data: AuditEntryResponse[]
  isLoading?: boolean
}

// Formats an ISO timestamp into a readable local datetime string.
const formatTime = (iso: string) => {
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso
  return d.toLocaleString()
}

export function AuditTable({ data, isLoading }: AuditTableProps) {
  const columns: ColumnDef<AuditEntryResponse>[] = useMemo(
    () => [
      {
        accessorKey: "action",
        header: "Action",
        cell: (info) => (
          <span className="text-sm font-medium text-gray-800">
            {getAuditActionLabel(info.getValue() as string)}
          </span>
        ),
      },
      {
        accessorKey: "actor",
        header: "Actor",
        cell: (info) => (
          <span className="text-sm text-gray-600">
            {info.getValue() as string}
          </span>
        ),
      },
      {
        accessorKey: "resourceType",
        header: "Resource",
        cell: (info) => (
          <span className="text-sm text-gray-600">
            {info.getValue() as string}:{info.row.original.resourceId}
          </span>
        ),
      },
      {
        accessorKey: "occurredAt",
        header: "Occurred at",
        cell: (info) => (
          <span className="text-sm text-gray-600">
            {formatTime(info.getValue() as string)}
          </span>
        ),
      },
    ],
    [],
  )

  if (!isLoading && data.length === 0) {
    return (
      <EmptyState
        icon={ScrollTextIcon}
        title="No audit entries"
        description="No audit trail entries match the current filters."
      />
    )
  }

  return (
    <div className="px-6 py-4">
      <Table
        columns={columns}
        data={data}
        isLoading={isLoading}
        isShowPagination={false}
      />
    </div>
  )
}
