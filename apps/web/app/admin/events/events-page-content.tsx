"use client"

import Link from "next/link"
import { useState } from "react"

import { Button } from "$/components/button"
import { ContentTitle } from "$/components/content-title"
import { LoadingSpinner } from "$/components/loading-spinner"
import { PanelCard } from "$/components/panel-card"
import { useAdminEvent, useAdminEvents } from "$/hooks/transactions/use-admin"
import { useAuthorization } from "$/providers/authorization-provider"
import { extractErrorMessage } from "$/utils/api-error"

const formatTime = (value: string) => {
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? value : date.toLocaleString()
}

export default function EventsPageContent() {
  const authorization = useAuthorization()
  const canRead = authorization.hasPermission("instance:read")
  const [workflowInstanceId, setWorkflowInstanceId] = useState("")
  const [type, setType] = useState("")
  const [source, setSource] = useState("")
  const [correlationId, setCorrelationId] = useState("")
  const [page, setPage] = useState(1)
  const [selectedEventId, setSelectedEventId] = useState("")
  const events = useAdminEvents({
    workflowInstanceId: workflowInstanceId || undefined,
    type: type || undefined,
    source: source || undefined,
    correlationId: correlationId || undefined,
    page,
    pageSize: 20,
    enabled: authorization.status === "ready" && canRead,
  })
  const detail = useAdminEvent(selectedEventId, canRead)

  if (!canRead) {
    return (
      <div className="space-y-6 p-8">
        <ContentTitle title="Event browser" />
        <div className="rounded-md bg-amber-50 px-4 py-3 text-sm text-amber-800">
          You are not authorized to browse events.
        </div>
      </div>
    )
  }
  if (events.isLoading) return <LoadingSpinner />

  return (
    <div className="space-y-6 p-8">
      <ContentTitle title="Event browser" />
      <PanelCard
        title="Immutable events"
        description="Tenant-scoped event history is read-only. There are no edit, delete, replay, or injection actions here."
      >
        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
          <Filter
            label="Workflow instance"
            value={workflowInstanceId}
            onChange={(value) => {
              setWorkflowInstanceId(value)
              setPage(1)
            }}
          />
          <Filter
            label="Type"
            value={type}
            onChange={(value) => {
              setType(value)
              setPage(1)
            }}
          />
          <Filter
            label="Source"
            value={source}
            onChange={(value) => {
              setSource(value)
              setPage(1)
            }}
          />
          <Filter
            label="Correlation id"
            value={correlationId}
            onChange={(value) => {
              setCorrelationId(value)
              setPage(1)
            }}
          />
        </div>
        {events.isError ? (
          <div className="mt-6 rounded-md bg-red-50 px-4 py-3 text-sm text-red-700">
            {extractErrorMessage(events.error) ?? "Events could not be loaded."}
          </div>
        ) : (
          <>
            <div className="mt-6 overflow-x-auto rounded-lg border border-slate-200">
              <table className="min-w-full divide-y divide-slate-200 text-left text-sm">
                <thead className="bg-slate-50 text-xs uppercase text-slate-500">
                  <tr>
                    <th className="px-4 py-3">Event</th>
                    <th className="px-4 py-3">Type / source</th>
                    <th className="px-4 py-3">Instance</th>
                    <th className="px-4 py-3">Timestamp</th>
                    <th className="px-4 py-3" />
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-200 bg-white">
                  {(events.data?.data ?? []).map((event) => (
                    <tr key={event.id}>
                      <td className="px-4 py-3">
                        <button
                          type="button"
                          className="font-mono text-xs text-slate-900 underline"
                          onClick={() => setSelectedEventId(event.id)}
                        >
                          {event.eventId}
                        </button>
                        <div className="text-xs text-slate-500">
                          Sequence {event.sequence}
                        </div>
                      </td>
                      <td className="px-4 py-3">
                        <div className="font-medium text-slate-900">
                          {event.type}
                        </div>
                        <div className="text-xs text-slate-500">
                          {event.source}
                        </div>
                      </td>
                      <td className="px-4 py-3">
                        {event.workflowInstanceId ? (
                          <Link
                            className="text-slate-700 underline"
                            href={`/admin/runtime-instances/${event.workflowInstanceId}`}
                          >
                            {event.workflowInstanceId}
                          </Link>
                        ) : (
                          <span className="text-slate-500">—</span>
                        )}
                      </td>
                      <td className="px-4 py-3 text-slate-600">
                        {formatTime(event.timestamp)}
                      </td>
                      <td className="px-4 py-3 text-right">
                        <Button
                          intent="clean"
                          onClick={() => setSelectedEventId(event.id)}
                        >
                          View detail
                        </Button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
              {(events.data?.data ?? []).length === 0 ? (
                <p className="px-4 py-8 text-center text-sm text-slate-500">
                  No events match these filters.
                </p>
              ) : null}
            </div>
            <div className="mt-6 flex items-center justify-end gap-3 text-sm text-slate-600">
              <Button
                intent="clean"
                disabled={page <= 1}
                onClick={() => setPage((value) => Math.max(1, value - 1))}
              >
                Previous
              </Button>
              <span>Page {page}</span>
              <Button
                intent="clean"
                disabled={!events.data?.hasNext}
                onClick={() => setPage((value) => value + 1)}
              >
                Next
              </Button>
            </div>
          </>
        )}
      </PanelCard>
      {selectedEventId ? (
        <PanelCard
          title="Event detail"
          action={
            <Button intent="clean" onClick={() => setSelectedEventId("")}>
              Close
            </Button>
          }
        >
          {detail.isLoading ? (
            <LoadingSpinner />
          ) : detail.isError || !detail.data ? (
            <p className="text-sm text-red-700">
              {extractErrorMessage(detail.error) ??
                "Event detail could not be loaded."}
            </p>
          ) : (
            <EventDetail event={detail.data} />
          )}
        </PanelCard>
      ) : null}
    </div>
  )
}

function Filter({
  label,
  value,
  onChange,
}: { label: string; value: string; onChange: (value: string) => void }) {
  return (
    <label className="block text-sm font-medium text-slate-700">
      {label}
      <input
        value={value}
        onChange={(event) => onChange(event.target.value)}
        className="mt-1 block w-full rounded-md border border-slate-300 px-3 py-2 font-normal text-slate-900"
      />
    </label>
  )
}

function EventDetail({
  event,
}: { event: NonNullable<ReturnType<typeof useAdminEvent>["data"]> }) {
  return (
    <div className="grid gap-6 lg:grid-cols-2">
      <dl className="space-y-3 text-sm">
        <DetailRow label="Event id" value={event.eventId} />
        <DetailRow label="Type" value={event.type} />
        <DetailRow label="Source" value={event.source} />
        <DetailRow label="Sequence" value={String(event.sequence)} />
        <DetailRow label="Correlation" value={event.correlationId ?? "—"} />
        <DetailRow label="Causation" value={event.causationId ?? "—"} />
      </dl>
      <div>
        <h3 className="text-sm font-medium text-slate-700">Payload</h3>
        <pre className="mt-2 max-h-96 overflow-auto rounded-md bg-slate-950 p-4 text-xs text-slate-100">
          {JSON.stringify(event.payload, null, 2)}
        </pre>
        <div className="mt-3 flex flex-wrap gap-3 text-sm">
          {event.workflowInstanceId ? (
            <Link
              className="text-slate-700 underline"
              href={`/admin/runtime-instances/${event.workflowInstanceId}`}
            >
              Open Runtime Inspector
            </Link>
          ) : null}
          {event.correlationId ? (
            <Link className="text-slate-700 underline" href="/admin/audit">
              Open audit context
            </Link>
          ) : null}
        </div>
      </div>
    </div>
  )
}

function DetailRow({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <dt className="text-slate-500">{label}</dt>
      <dd className="break-all font-medium text-slate-900">{value}</dd>
    </div>
  )
}
