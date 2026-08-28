"use client"

import { useCallback, useState } from "react"

import { Button } from "$/components/button"
import { ContentTitle } from "$/components/content-title"
import { PanelCard } from "$/components/panel-card"
import { Select } from "$/components/select"
import { useAuditList } from "$/hooks/transactions/use-audit"
import { auditActionLabels, auditActions } from "@openstate/schemas"
import { ChevronLeftIcon, ChevronRightIcon } from "lucide-react"
import { AuditTable } from "./_components/audit-table"

const pageSize = 20

const actionFilterOptions = auditActions.map((value) => ({
  value,
  label: auditActionLabels.find((l) => l.value === value)?.label ?? value,
}))

export default function AuditPageContent() {
  const [action, setAction] = useState<string | undefined>(undefined)
  const [resourceType, setResourceType] = useState<string | undefined>(
    undefined,
  )
  const [actor, setActor] = useState<string | undefined>(undefined)
  const [page, setPage] = useState(1)

  const { data, isLoading, isError, refetch } = useAuditList({
    action,
    resourceType,
    actor,
    page,
    pageSize,
  })

  const entries = data?.data || []
  const hasNext = data?.hasNext ?? false
  const isFirstPage = page <= 1

  const handleApplyFilters = useCallback(() => {
    setPage(1)
  }, [])

  const handleNext = useCallback(() => {
    if (hasNext) setPage((p) => p + 1)
  }, [hasNext])

  const handlePrev = useCallback(() => {
    setPage((p) => Math.max(1, p - 1))
  }, [])

  return (
    <div className="space-y-6 p-8">
      <ContentTitle title="Audit Log" />

      <PanelCard
        title="Audit Trail"
        description="Tenant-scoped, append-only audit trail (PRD 50)"
      >
        <div className="flex flex-wrap items-end gap-4 px-6 pt-4">
          <div className="w-56">
            <Select<{ value: string; label: string }>
              label="Action"
              placeholder="All actions"
              options={actionFilterOptions}
              value={action ? { value: action, label: action } : undefined}
              getOptionLabel={(o) => o.label}
              getOptionValue={(o) => o.value}
              isClearable
              onChange={(option) =>
                setAction((option as { value?: string } | null)?.value)
              }
            />
          </div>
          <div className="w-48">
            <Select<{ value: string; label: string }>
              label="Resource type"
              placeholder="All resources"
              options={[
                { value: "workflow", label: "Workflow" },
                { value: "capability", label: "Capability" },
                { value: "binding", label: "Binding" },
                { value: "role_assignment", label: "Role assignment" },
              ]}
              value={
                resourceType
                  ? { value: resourceType, label: resourceType }
                  : undefined
              }
              getOptionLabel={(o) => o.label}
              getOptionValue={(o) => o.value}
              isClearable
              onChange={(option) =>
                setResourceType((option as { value?: string } | null)?.value)
              }
            />
          </div>
          <div className="flex-1 min-w-40">
            <input
              type="text"
              placeholder="Filter by actor (user id)"
              value={actor ?? ""}
              onChange={(e) => setActor(e.target.value || undefined)}
              className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm"
            />
          </div>
          <Button intent="secondary" onClick={handleApplyFilters}>
            Apply filters
          </Button>
        </div>

        {isError ? (
          <div className="px-6 py-8 text-center text-sm text-red-600">
            Failed to load audit entries.{" "}
            <Button intent="clean" onClick={() => void refetch()}>
              Retry
            </Button>
          </div>
        ) : (
          <AuditTable data={entries} isLoading={isLoading} />
        )}

        <div className="flex items-center justify-end gap-2 px-6 py-4">
          <Button
            intent="clean"
            leftIcon={<ChevronLeftIcon size={16} />}
            disabled={isFirstPage}
            onClick={handlePrev}
          >
            Prev
          </Button>
          <span className="text-sm text-gray-600">Page {page}</span>
          <Button
            intent="clean"
            rightIcon={<ChevronRightIcon size={16} />}
            disabled={!hasNext}
            onClick={handleNext}
          >
            Next
          </Button>
        </div>
      </PanelCard>
    </div>
  )
}
