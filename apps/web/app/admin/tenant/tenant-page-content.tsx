"use client"

import { type FormEvent, useEffect, useState } from "react"

import { AdminFlowGuide } from "$/components/admin-console"
import { Button } from "$/components/button"
import { ContentTitle } from "$/components/content-title"
import { LoadingSpinner } from "$/components/loading-spinner"
import { PanelCard } from "$/components/panel-card"
import {
  useAdminTenant,
  useAdminTenantUpdate,
} from "$/hooks/transactions/use-admin"
import { useAuthorization } from "$/providers/authorization-provider"
import { extractErrorMessage } from "$/utils/api-error"
import { updateTenantSchema } from "@openstate/schemas"

type Notice = { type: "success" | "error"; text: string }

export default function TenantPageContent() {
  const authorization = useAuthorization()
  const canRead = authorization.hasPermission("tenant:read")
  const canUpdate = authorization.hasPermission("tenant:update")
  const tenant = useAdminTenant(authorization.status === "ready" && canRead)
  const update = useAdminTenantUpdate()
  const [form, setForm] = useState({ name: "", slug: "", description: "" })
  const [notice, setNotice] = useState<Notice | null>(null)

  useEffect(() => {
    if (!tenant.data) return
    setForm({
      name: tenant.data.name,
      slug: tenant.data.slug,
      description: tenant.data.description,
    })
  }, [tenant.data])

  if (!canRead) {
    return (
      <div className="space-y-6 p-8">
        <ContentTitle title="Tenant settings" />
        <div className="rounded-md bg-amber-50 px-4 py-3 text-sm text-amber-800">
          You are not authorized to view tenant settings.
        </div>
      </div>
    )
  }

  if (tenant.isLoading) return <LoadingSpinner />

  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    const parsed = updateTenantSchema.safeParse(form)
    if (!parsed.success) {
      setNotice({
        type: "error",
        text: parsed.error.issues[0]?.message ?? "Check the form values.",
      })
      return
    }
    if (!window.confirm("Update the tenant profile?")) return

    setNotice(null)
    update.mutate(parsed.data, {
      onSuccess: () =>
        setNotice({ type: "success", text: "Tenant profile updated." }),
      onError: (error) =>
        setNotice({
          type: "error",
          text:
            extractErrorMessage(error) ??
            "Tenant profile could not be updated.",
        }),
    })
  }

  return (
    <div className="space-y-6 p-8">
      <ContentTitle title="Tenant settings" />
      <AdminFlowGuide currentStep="tenant" />
      {notice ? (
        <output
          className={`rounded-md px-4 py-3 text-sm ${
            notice.type === "success"
              ? "bg-green-50 text-green-700"
              : "bg-red-50 text-red-700"
          }`}
        >
          {notice.text}
        </output>
      ) : null}
      {tenant.isError ? (
        <div className="rounded-md bg-red-50 px-4 py-3 text-sm text-red-700">
          {extractErrorMessage(tenant.error) ??
            "Tenant profile could not be loaded."}
        </div>
      ) : (
        <PanelCard
          title="Current tenant"
          description="Changes apply to this tenant only."
        >
          <form className="max-w-2xl space-y-5" onSubmit={handleSubmit}>
            <label
              className="block text-sm font-medium text-slate-700"
              htmlFor="tenant-name"
            >
              Name
              <input
                id="tenant-name"
                value={form.name}
                onChange={(event) =>
                  setForm((value) => ({ ...value, name: event.target.value }))
                }
                disabled={!canUpdate || update.isPending}
                className="mt-1 block w-full rounded-md border border-slate-300 px-3 py-2 font-normal text-slate-900"
              />
            </label>
            <label
              className="block text-sm font-medium text-slate-700"
              htmlFor="tenant-slug"
            >
              Slug
              <input
                id="tenant-slug"
                value={form.slug}
                onChange={(event) =>
                  setForm((value) => ({ ...value, slug: event.target.value }))
                }
                disabled={!canUpdate || update.isPending}
                className="mt-1 block w-full rounded-md border border-slate-300 px-3 py-2 font-normal text-slate-900"
              />
            </label>
            <label
              className="block text-sm font-medium text-slate-700"
              htmlFor="tenant-description"
            >
              Description
              <textarea
                id="tenant-description"
                rows={4}
                value={form.description}
                onChange={(event) =>
                  setForm((value) => ({
                    ...value,
                    description: event.target.value,
                  }))
                }
                disabled={!canUpdate || update.isPending}
                className="mt-1 block w-full rounded-md border border-slate-300 px-3 py-2 font-normal text-slate-900"
              />
            </label>
            {canUpdate ? (
              <Button type="submit" intent="primary" loading={update.isPending}>
                Save changes
              </Button>
            ) : (
              <p className="text-sm text-slate-500">
                You have read-only access to this tenant.
              </p>
            )}
          </form>
        </PanelCard>
      )}
    </div>
  )
}
