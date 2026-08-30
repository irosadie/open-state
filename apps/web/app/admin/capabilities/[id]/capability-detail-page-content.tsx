"use client"

import Link from "next/link"
import { useState } from "react"

import { PermissionGate } from "$/components/auth-guard/permission-gate"
import { Button } from "$/components/button"
import { ContentTitle } from "$/components/content-title"
import { LoadingSpinner } from "$/components/loading-spinner"
import { PanelCard } from "$/components/panel-card"
import { useAuthorization } from "$/providers/authorization-provider"
import { ArrowLeftIcon } from "lucide-react"

import {
  useCapabilitiesGet,
  useCapabilitiesUpdate,
} from "$/hooks/transactions/use-capability"
import { extractErrorMessage } from "$/utils/api-error"
import type { UpdateCapabilitySchemaProps } from "@openstate/schemas"
import { BindingsPanel } from "./_components/bindings-panel"
import { CapabilityDetail } from "./_components/capability-detail"
import { CapabilityEditForm } from "./_components/capability-edit-form"
import { TestInvocationPanel } from "./_components/test-invocation-panel"

type CapabilityDetailPageContentProps = {
  id: string
}

export default function CapabilityDetailPageContent({
  id,
}: CapabilityDetailPageContentProps) {
  const authorization = useAuthorization()
  const [isEditing, setIsEditing] = useState(false)
  const [notice, setNotice] = useState<{
    type: "success" | "error"
    text: string
  } | null>(null)

  const {
    data: capability,
    isLoading,
    isError,
    refetch,
  } = useCapabilitiesGet({
    id,
    enabled:
      authorization.status === "ready" &&
      authorization.hasPermission("capability:read"),
  })
  const { mutateAsync: updateMutateAsync, isPending: isUpdatePending } =
    useCapabilitiesUpdate()

  const handleSave = async (payload: UpdateCapabilitySchemaProps) => {
    await updateMutateAsync(
      { id, payload },
      {
        onSuccess: () => {
          setNotice({ type: "success", text: "Capability updated" })
          setIsEditing(false)
        },
        onError: (error) => {
          setNotice({
            type: "error",
            text: extractErrorMessage(error) || "Failed to update capability",
          })
        },
        onSettled: () => {
          void refetch()
        },
      },
    )
  }

  if (isLoading) {
    return <LoadingSpinner />
  }

  if (isError || !capability) {
    return (
      <div className="space-y-6 p-8">
        <ContentTitle title="Capability" />
        <div className="rounded-md bg-red-50 px-4 py-3 text-sm text-red-700">
          Capability not found.
        </div>
        <Link href="/admin/capabilities">
          <Button intent="secondary" leftIcon={<ArrowLeftIcon size={16} />}>
            Back to capabilities
          </Button>
        </Link>
      </div>
    )
  }

  return (
    <div className="space-y-6 p-8">
      <ContentTitle
        title={capability.name}
        breadcrumbData={[
          { title: "Capabilities", link: "/admin/capabilities" },
        ]}
      />

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

      <div className="grid gap-6 lg:grid-cols-2">
        <PanelCard
          title="Capability"
          action={
            !isEditing ? (
              <PermissionGate action="capability:update">
                <Button intent="secondary" onClick={() => setIsEditing(true)}>
                  Edit
                </Button>
              </PermissionGate>
            ) : undefined
          }
        >
          {isEditing ? (
            <CapabilityEditForm
              capability={capability}
              isPending={isUpdatePending}
              onCancel={() => setIsEditing(false)}
              onSave={handleSave}
            />
          ) : (
            <CapabilityDetail capability={capability} />
          )}
        </PanelCard>

        <BindingsPanel
          capabilityId={capability.id}
          enabled={authorization.hasPermission("binding:read")}
        />
      </div>

      <TestInvocationPanel
        capabilityId={capability.id}
        inputSchema={capability.inputSchema}
      />
    </div>
  )
}
