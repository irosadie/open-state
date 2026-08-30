"use client"

import Link from "next/link"
import { useMemo } from "react"

import {
  ActionsDropdown,
  type ActionsDropdownProps,
} from "$/components/actions-dropdown"
import { EmptyState } from "$/components/empty-state"
import { StatusBadge } from "$/components/status-badge"
import { type ColumnDef, Table } from "$/components/table"
import {
  getCapabilityStatusLabel,
  getProviderTypeLabel,
} from "@openstate/schemas"
import type { CapabilityResponse } from "@openstate/types"
import { BoxesIcon } from "lucide-react"

type CapabilitiesTableProps = {
  data: CapabilityResponse[]
  isLoading?: boolean
  onDisable?: (id: string) => void
}

const statusVariant: Record<
  CapabilityResponse["status"],
  "success" | "warning" | "danger"
> = {
  ACTIVE: "success",
  INACTIVE: "warning",
  DISABLED: "danger",
}

export function CapabilitiesTable({
  data,
  isLoading,
  onDisable,
}: CapabilitiesTableProps) {
  const columns: ColumnDef<CapabilityResponse>[] = useMemo(
    () => [
      {
        accessorKey: "name",
        header: "Name",
        cell: (info) => (
          <Link
            href={`/admin/capabilities/${info.row.original.id}`}
            className="text-sm font-medium text-blue-700 hover:underline"
          >
            {info.getValue() as string}
          </Link>
        ),
      },
      {
        accessorKey: "providerType",
        header: "Provider",
        cell: (info) => (
          <span className="text-sm text-gray-600">
            {getProviderTypeLabel(
              info.getValue() as CapabilityResponse["providerType"],
            )}
          </span>
        ),
      },
      {
        accessorKey: "status",
        header: "Status",
        cell: (info) => {
          const status = info.getValue() as CapabilityResponse["status"]

          return (
            <StatusBadge
              variant={statusVariant[status]}
              activeLabel={getCapabilityStatusLabel(status)}
              inactiveLabel={getCapabilityStatusLabel(status)}
            />
          )
        },
      },
      {
        accessorKey: "version",
        header: "Version",
        cell: (info) => (
          <span className="text-sm text-gray-600">
            {info.getValue() as number}
          </span>
        ),
      },
      {
        id: "actions",
        header: "",
        cell: (info) => {
          const actions: ActionsDropdownProps["actions"] = [
            {
              label: "View",
              onClick: () => {
                window.location.href = `/admin/capabilities/${info.row.original.id}`
              },
            },
          ]

          if (onDisable) {
            actions.push({
              label: "Disable",
              onClick: () => onDisable(info.row.original.id),
              destructive: true,
            })
          }

          return <ActionsDropdown actions={actions} />
        },
      },
    ],
    [onDisable],
  )

  if (!isLoading && data.length === 0) {
    return (
      <EmptyState
        icon={BoxesIcon}
        title="No capabilities"
        description="Register a capability to get started."
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
