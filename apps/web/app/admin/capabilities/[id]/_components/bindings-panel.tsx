"use client"

import { useState } from "react"

import { PermissionGate } from "$/components/auth-guard/permission-gate"
import { Button } from "$/components/button"
import { EmptyState } from "$/components/empty-state"
import { Input } from "$/components/input"
import { LoadingSpinner } from "$/components/loading-spinner"
import { PanelCard } from "$/components/panel-card"
import { Select } from "$/components/select"
import { StatusBadge } from "$/components/status-badge"
import { LinkIcon, Trash2Icon } from "lucide-react"

import {
  useCapabilitiesCreateBinding,
  useCapabilitiesDeleteBinding,
  useCapabilitiesListBindings,
} from "$/hooks/transactions/use-capability"
import { extractErrorMessage } from "$/utils/api-error"
import {
  type BindingSchemaProps,
  bindingPermissionLabels,
  bindingSchema,
  bindingScopeTypeLabels,
  getBindingPermissionLabel,
  getBindingScopeTypeLabel,
} from "@openstate/schemas"
import type {
  BindingPermission,
  BindingScopeType,
  CapabilityBindingResponse,
} from "@openstate/types"
import type { ZodError } from "zod"

const permissionVariant: Record<BindingPermission, "success" | "danger"> = {
  ALLOW: "success",
  DENY: "danger",
}

type BindingsPanelProps = {
  capabilityId: string
  enabled?: boolean
}

type FieldErrors = Partial<Record<keyof BindingSchemaProps, string>>

export function BindingsPanel({
  capabilityId,
  enabled = true,
}: BindingsPanelProps) {
  const [scopeType, setScopeType] = useState<BindingScopeType | undefined>(
    undefined,
  )
  const [scopeId, setScopeId] = useState("")
  const [permission, setPermission] = useState<BindingPermission>("ALLOW")
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({})
  const [notice, setNotice] = useState<string | null>(null)

  const {
    data: bindings,
    isLoading,
    refetch,
  } = useCapabilitiesListBindings({
    capabilityId,
    enabled,
  })
  const { mutateAsync: createMutateAsync, isPending: isCreatePending } =
    useCapabilitiesCreateBinding()
  const { mutate: deleteMutate } = useCapabilitiesDeleteBinding()

  const handleCreate = async () => {
    const parsed = bindingSchema.safeParse({
      scopeType,
      scopeId,
      permission,
    })

    if (!parsed.success) {
      const errors: FieldErrors = {}

      for (const issue of (parsed.error as ZodError).issues) {
        const key = issue.path[0] as keyof BindingSchemaProps

        if (key) {
          errors[key] = issue.message
        }
      }

      setFieldErrors(errors)

      return
    }

    await createMutateAsync(
      { capabilityId, payload: parsed.data },
      {
        onSuccess: () => {
          setScopeId("")
          setFieldErrors({})
          setNotice(null)
        },
        onError: (error) => {
          setNotice(extractErrorMessage(error) || "Failed to create binding")
        },
        onSettled: () => {
          void refetch()
        },
      },
    )
  }

  const handleDelete = (binding: CapabilityBindingResponse) => {
    deleteMutate(
      { bindingId: binding.id, capabilityId },
      {
        onError: (error) => {
          setNotice(extractErrorMessage(error) || "Failed to remove binding")
        },
        onSettled: () => {
          void refetch()
        },
      },
    )
  }

  return (
    <PanelCard
      title="Bindings"
      description="Bind this capability to tenant, workflow, or state scopes (PRD §60)"
    >
      <div className="space-y-4 px-6 py-4">
        {notice ? (
          <div className="rounded-md bg-red-50 px-4 py-3 text-sm text-red-700">
            {notice}
          </div>
        ) : null}

        <div className="grid grid-cols-2 gap-3">
          <Select<(typeof bindingScopeTypeLabels)[number]>
            label="Scope type"
            required
            options={bindingScopeTypeLabels}
            getOptionLabel={(o) => o.label}
            getOptionValue={(o) => o.value}
            onChange={(option) =>
              setScopeType(
                (option as (typeof bindingScopeTypeLabels)[number] | null)
                  ?.value as BindingScopeType | undefined,
              )
            }
            error={fieldErrors.scopeType}
          />

          <Select<(typeof bindingPermissionLabels)[number]>
            label="Permission"
            options={bindingPermissionLabels}
            value={
              bindingPermissionLabels.find((l) => l.value === permission) ??
              undefined
            }
            getOptionLabel={(o) => o.label}
            getOptionValue={(o) => o.value}
            onChange={(option) =>
              setPermission(
                (option as (typeof bindingPermissionLabels)[number])
                  .value as BindingPermission,
              )
            }
          />
        </div>

        <div className="flex gap-2">
          <Input
            label="Scope ID"
            required
            placeholder="workflow-id or state-id"
            value={scopeId}
            error={fieldErrors.scopeId}
            onChange={(e) => setScopeId(e.target.value)}
          />
          <div className="flex items-end pb-0.5">
            <PermissionGate action="binding:create">
              <Button
                intent="primary"
                loading={isCreatePending}
                onClick={() => void handleCreate()}
              >
                Bind
              </Button>
            </PermissionGate>
          </div>
        </div>

        {isLoading ? (
          <LoadingSpinner />
        ) : !bindings || bindings.length === 0 ? (
          <EmptyState
            icon={LinkIcon}
            title="No bindings"
            description="Bind this capability to a scope to control where it is available."
          />
        ) : (
          <ul className="divide-y divide-gray-100">
            {bindings.map((binding) => (
              <li
                key={binding.id}
                className="flex items-center justify-between py-2"
              >
                <div className="flex items-center gap-3">
                  <StatusBadge
                    variant={permissionVariant[binding.permission]}
                    activeLabel={getBindingPermissionLabel(binding.permission)}
                    inactiveLabel={getBindingPermissionLabel(
                      binding.permission,
                    )}
                  />
                  <span className="text-sm font-medium text-gray-700">
                    {getBindingScopeTypeLabel(binding.scopeType)}
                  </span>
                  <span className="font-mono text-xs text-gray-500">
                    {binding.scopeId}
                  </span>
                </div>
                <PermissionGate action="binding:delete">
                  <Button
                    intent="clean"
                    leftIcon={<Trash2Icon size={15} className="text-red-500" />}
                    onClick={() => handleDelete(binding)}
                    aria-label={`Remove binding ${binding.id}`}
                  />
                </PermissionGate>
              </li>
            ))}
          </ul>
        )}
      </div>
    </PanelCard>
  )
}
