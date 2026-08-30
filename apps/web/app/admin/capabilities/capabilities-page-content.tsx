"use client"

import { useCallback, useState } from "react"

import { PermissionGate } from "$/components/auth-guard/permission-gate"
import { Button } from "$/components/button"
import { ContentTitle } from "$/components/content-title"
import { PanelCard } from "$/components/panel-card"
import { Select } from "$/components/select"
import { useAuthorization } from "$/providers/authorization-provider"
import { PlusIcon } from "lucide-react"

import {
  useCapabilitiesCreate,
  useCapabilitiesDelete,
  useCapabilitiesList,
} from "$/hooks/transactions/use-capability"
import { extractErrorMessage } from "$/utils/api-error"
import { capabilityStatuses, providerTypes } from "@openstate/schemas"
import type { CreateCapabilitySchemaProps } from "@openstate/schemas"
import type {
  CapabilityProviderType,
  CapabilityResponse,
  CapabilityStatus,
} from "@openstate/types"
import { CapabilityConfirmDialog } from "./_components/capability-confirm-dialog"
import { CapabilityFormDialog } from "./_components/capability-form"
import { CapabilitiesTable } from "./_components/capability-table"

const providerFilterOptions = providerTypes.map((value) => ({
  value,
  label: value,
}))
const statusFilterOptions = capabilityStatuses.map((value) => ({
  value,
  label: value,
}))

export default function CapabilitiesPageContent() {
  const authorization = useAuthorization()
  const [providerType, setProviderType] = useState<
    CapabilityProviderType | undefined
  >(undefined)
  const [status, setStatus] = useState<CapabilityStatus | undefined>(undefined)
  const [isFormOpen, setIsFormOpen] = useState(false)
  const [notice, setNotice] = useState<{
    type: "success" | "error"
    text: string
  } | null>(null)
  const [pendingDisable, setPendingDisable] =
    useState<CapabilityResponse | null>(null)

  const {
    data: capabilities,
    isLoading,
    isError,
    refetch,
  } = useCapabilitiesList({
    providerType,
    status,
    enabled:
      authorization.status === "ready" &&
      authorization.hasPermission("capability:read"),
  })

  const { mutateAsync: createMutateAsync, isPending: isCreatePending } =
    useCapabilitiesCreate()
  const { mutate: deleteMutate, isPending: isDeletePending } =
    useCapabilitiesDelete()
  const canDelete = authorization.hasPermission("capability:delete")

  const handleCreate = useCallback(
    async (payload: CreateCapabilitySchemaProps) => {
      await createMutateAsync(payload, {
        onSuccess: () => {
          setNotice({ type: "success", text: "Capability created" })
          setIsFormOpen(false)
        },
        onError: (error) => {
          setNotice({
            type: "error",
            text: extractErrorMessage(error) || "Failed to create capability",
          })
        },
        onSettled: () => {
          void refetch()
        },
      })
    },
    [createMutateAsync, refetch],
  )

  const handleConfirmDisable = useCallback(() => {
    if (!pendingDisable) return

    deleteMutate(pendingDisable.id, {
      onSuccess: () => {
        setNotice({ type: "success", text: "Capability disabled" })
      },
      onError: (error) => {
        setNotice({
          type: "error",
          text: extractErrorMessage(error) || "Failed to disable capability",
        })
      },
      onSettled: () => {
        setPendingDisable(null)
        void refetch()
      },
    })
  }, [deleteMutate, pendingDisable, refetch])

  return (
    <div className="space-y-6 p-8">
      <ContentTitle title="Capabilities" />

      {notice ? (
        <div
          className={`rounded-md px-4 py-3 text-sm ${
            notice.type === "success"
              ? "bg-green-50 text-green-700"
              : "bg-red-50 text-red-700"
          }`}
        >
          {notice.text}
        </div>
      ) : null}

      <PanelCard
        title="Capability Registry"
        description="Manage the tenant capability registry (PRD §59)"
        action={
          <PermissionGate action="capability:create">
            <Button
              intent="primary"
              leftIcon={<PlusIcon size={16} />}
              onClick={() => setIsFormOpen(true)}
            >
              New Capability
            </Button>
          </PermissionGate>
        }
      >
        <div className="flex flex-wrap items-center gap-4 px-6 pt-4">
          <div className="w-48">
            <Select<{ value: string; label: string }>
              label="Provider type"
              placeholder="All providers"
              options={providerFilterOptions}
              value={
                providerType
                  ? { value: providerType, label: providerType }
                  : undefined
              }
              getOptionLabel={(o) => o.label}
              getOptionValue={(o) => o.value}
              isClearable
              onChange={(option) =>
                setProviderType(
                  (option as { value?: CapabilityProviderType } | null)?.value,
                )
              }
            />
          </div>
          <div className="w-48">
            <Select<{ value: string; label: string }>
              label="Status"
              placeholder="All statuses"
              options={statusFilterOptions}
              value={status ? { value: status, label: status } : undefined}
              getOptionLabel={(o) => o.label}
              getOptionValue={(o) => o.value}
              isClearable
              onChange={(option) =>
                setStatus(
                  (option as { value?: CapabilityStatus } | null)?.value,
                )
              }
            />
          </div>
        </div>

        {isError ? (
          <div className="px-6 py-8 text-center text-sm text-red-600">
            Failed to load capabilities.{" "}
            <Button intent="clean" onClick={() => void refetch()}>
              Retry
            </Button>
          </div>
        ) : (
          <CapabilitiesTable
            data={capabilities || []}
            isLoading={isLoading}
            onDisable={
              canDelete
                ? (id) => {
                    const cap = (capabilities || []).find((c) => c.id === id)

                    if (cap) {
                      setPendingDisable(cap)
                    }
                  }
                : undefined
            }
          />
        )}
      </PanelCard>

      <CapabilityFormDialog
        open={isFormOpen}
        onOpenChange={setIsFormOpen}
        isPending={isCreatePending}
        onSubmit={handleCreate}
      />

      <PermissionGate action="capability:delete">
        <CapabilityConfirmDialog
          open={!!pendingDisable}
          isPending={isDeletePending}
          title="Disable capability"
          description={`Disabling "${pendingDisable?.name ?? ""}" also removes its bindings. Continue?`}
          confirmLabel="Yes, disable"
          onCancel={() => setPendingDisable(null)}
          onConfirm={handleConfirmDisable}
        />
      </PermissionGate>
    </div>
  )
}
