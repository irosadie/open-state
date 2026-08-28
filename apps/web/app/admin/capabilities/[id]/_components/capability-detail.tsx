"use client"

import { StatusBadge } from "$/components/status-badge"
import {
  getCapabilityStatusLabel,
  getProviderTypeLabel,
} from "@openstate/schemas"
import type { CapabilityResponse } from "@openstate/types"

const statusVariant: Record<
  CapabilityResponse["status"],
  "success" | "warning" | "danger"
> = {
  ACTIVE: "success",
  INACTIVE: "warning",
  DISABLED: "danger",
}

const jsonText = (value?: Record<string, unknown> | null) => {
  if (!value) return "-"

  try {
    return JSON.stringify(value, null, 2)
  } catch {
    return "-"
  }
}

type CapabilityDetailProps = {
  capability: CapabilityResponse
}

export function CapabilityDetail({ capability }: CapabilityDetailProps) {
  return (
    <div className="space-y-4 px-1">
      <DetailRow label="ID" value={capability.id} />
      <DetailRow
        label="Provider"
        value={getProviderTypeLabel(capability.providerType)}
      />
      <DetailRow label="Status">
        <StatusBadge
          variant={statusVariant[capability.status]}
          activeLabel={getCapabilityStatusLabel(capability.status)}
          inactiveLabel={getCapabilityStatusLabel(capability.status)}
        />
      </DetailRow>
      <DetailRow label="Version" value={String(capability.version)} />
      <DetailRow label="Description" value={capability.description || "-"} />

      <div>
        <p className="text-xs font-semibold uppercase tracking-wide text-gray-500">
          Input schema
        </p>
        <pre className="mt-1 max-h-40 overflow-auto rounded-md bg-gray-50 p-3 text-xs text-gray-700">
          {jsonText(capability.inputSchema)}
        </pre>
      </div>

      <div>
        <p className="text-xs font-semibold uppercase tracking-wide text-gray-500">
          Output schema
        </p>
        <pre className="mt-1 max-h-40 overflow-auto rounded-md bg-gray-50 p-3 text-xs text-gray-700">
          {jsonText(capability.outputSchema)}
        </pre>
      </div>

      <DetailRow
        label="Credential reference"
        value={capability.credentialReference || "-"}
        hint="Only the reference is shown; secret values are never displayed."
      />

      <DetailRow label="Created at" value={capability.createdAt} />
      <DetailRow label="Updated at" value={capability.updatedAt} />
    </div>
  )
}

function DetailRow({
  label,
  value,
  hint,
  children,
}: {
  label: string
  value?: string
  hint?: string
  children?: React.ReactNode
}) {
  return (
    <div>
      <p className="text-xs font-semibold uppercase tracking-wide text-gray-500">
        {label}
      </p>
      <div className="mt-0.5 text-sm text-gray-800">
        {children ?? value ?? "-"}
      </div>
      {hint ? <p className="mt-0.5 text-xs text-gray-400">{hint}</p> : null}
    </div>
  )
}
