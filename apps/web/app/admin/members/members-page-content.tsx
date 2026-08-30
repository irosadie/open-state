"use client"

import { useState } from "react"

import { Button } from "$/components/button"
import { ContentTitle } from "$/components/content-title"
import { LoadingSpinner } from "$/components/loading-spinner"
import { PanelCard } from "$/components/panel-card"
import {
  useAdminMemberRemove,
  useAdminMemberRoleUpdate,
  useAdminMembers,
} from "$/hooks/transactions/use-admin"
import { useAuthorization } from "$/providers/authorization-provider"
import { extractErrorMessage } from "$/utils/api-error"
import {
  getTenantRoleLabel,
  tenantRoleLabels,
  tenantRoles,
  updateMembershipRoleSchema,
} from "@openstate/schemas"
import type { TenantRole } from "@openstate/types"

type Notice = { type: "success" | "error"; text: string }

export default function MembersPageContent() {
  const authorization = useAuthorization()
  const canRead = authorization.hasPermission("user:read")
  const canUpdate = authorization.hasPermission("user:update")
  const canRemove = authorization.hasPermission("user:delete")
  const [search, setSearch] = useState("")
  const [page, setPage] = useState(1)
  const [notice, setNotice] = useState<Notice | null>(null)
  const members = useAdminMembers({
    search,
    page,
    pageSize: 20,
    enabled: authorization.status === "ready" && canRead,
  })
  const roleUpdate = useAdminMemberRoleUpdate()
  const remove = useAdminMemberRemove()

  if (!canRead) {
    return (
      <div className="space-y-6 p-8">
        <ContentTitle title="Members & roles" />
        <div className="rounded-md bg-amber-50 px-4 py-3 text-sm text-amber-800">
          You are not authorized to view tenant members.
        </div>
      </div>
    )
  }
  if (members.isLoading) return <LoadingSpinner />

  const submitRole = (
    userId: string,
    role: string,
    currentRole: TenantRole,
  ) => {
    if (role === currentRole) return
    const parsed = updateMembershipRoleSchema.safeParse({ role })
    if (!parsed.success) {
      setNotice({ type: "error", text: "Select a valid tenant role." })
      return
    }
    if (
      !window.confirm(
        `Change this member's role to ${getTenantRoleLabel(parsed.data.role)}?`,
      )
    )
      return
    setNotice(null)
    roleUpdate.mutate(
      { userId, role: parsed.data.role },
      {
        onSuccess: () =>
          setNotice({ type: "success", text: "Member role updated." }),
        onError: (error) =>
          setNotice({
            type: "error",
            text:
              extractErrorMessage(error) ?? "Member role could not be updated.",
          }),
      },
    )
  }

  const removeMember = (userId: string, name: string) => {
    if (!window.confirm(`Remove ${name || "this member"} from the tenant?`))
      return
    setNotice(null)
    remove.mutate(userId, {
      onSuccess: () => setNotice({ type: "success", text: "Member removed." }),
      onError: (error) =>
        setNotice({
          type: "error",
          text: extractErrorMessage(error) ?? "Member could not be removed.",
        }),
    })
  }

  return (
    <div className="space-y-6 p-8">
      <ContentTitle title="Members & roles" />
      {notice ? (
        <output
          className={`rounded-md px-4 py-3 text-sm ${notice.type === "success" ? "bg-green-50 text-green-700" : "bg-red-50 text-red-700"}`}
        >
          {notice.text}
        </output>
      ) : null}
      <PanelCard
        title="Tenant members"
        description="Memberships are always limited to the current tenant."
      >
        <div className="mb-6 flex flex-wrap items-end gap-3">
          <label
            className="min-w-64 flex-1 text-sm font-medium text-slate-700"
            htmlFor="member-search"
          >
            Search
            <input
              id="member-search"
              value={search}
              onChange={(event) => {
                setSearch(event.target.value)
                setPage(1)
              }}
              placeholder="Name or email"
              className="mt-1 block w-full rounded-md border border-slate-300 px-3 py-2 font-normal text-slate-900"
            />
          </label>
        </div>
        {members.isError ? (
          <div className="rounded-md bg-red-50 px-4 py-3 text-sm text-red-700">
            {extractErrorMessage(members.error) ??
              "Members could not be loaded."}
          </div>
        ) : (
          <div className="overflow-x-auto rounded-lg border border-slate-200">
            <table className="min-w-full divide-y divide-slate-200 text-left text-sm">
              <thead className="bg-slate-50 text-xs uppercase text-slate-500">
                <tr>
                  <th className="px-4 py-3">Member</th>
                  <th className="px-4 py-3">Role</th>
                  <th className="px-4 py-3">Status</th>
                  <th className="px-4 py-3" />
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-200 bg-white">
                {(members.data?.data ?? []).map((member) => (
                  <tr key={member.userId}>
                    <td className="px-4 py-3">
                      <div className="font-medium text-slate-900">
                        {member.name || member.email}
                      </div>
                      <div className="text-xs text-slate-500">
                        {member.email}
                      </div>
                    </td>
                    <td className="px-4 py-3">
                      <select
                        aria-label={`Role for ${member.name || member.email}`}
                        value={member.role}
                        disabled={
                          !canUpdate || roleUpdate.isPending || remove.isPending
                        }
                        onChange={(event) =>
                          submitRole(
                            member.userId,
                            event.target.value,
                            member.role,
                          )
                        }
                        className="rounded-md border border-slate-300 px-2 py-1 text-slate-900"
                      >
                        {tenantRoles.map((role) => (
                          <option key={role} value={role}>
                            {tenantRoleLabels.find(
                              (item) => item.value === role,
                            )?.label ?? role}
                          </option>
                        ))}
                      </select>
                    </td>
                    <td className="px-4 py-3 text-slate-600">
                      {member.status}
                    </td>
                    <td className="px-4 py-3 text-right">
                      {canRemove ? (
                        <Button
                          intent="clean"
                          onClick={() =>
                            removeMember(
                              member.userId,
                              member.name || member.email,
                            )
                          }
                          loading={remove.isPending}
                        >
                          Remove
                        </Button>
                      ) : null}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
            {(members.data?.data ?? []).length === 0 ? (
              <p className="px-4 py-8 text-center text-sm text-slate-500">
                No members match this search.
              </p>
            ) : null}
          </div>
        )}
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
            disabled={!members.data?.hasNext}
            onClick={() => setPage((value) => value + 1)}
          >
            Next
          </Button>
        </div>
      </PanelCard>
    </div>
  )
}
