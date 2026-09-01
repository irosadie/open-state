"use client"

import { type FormEvent, useState } from "react"

import { Button } from "$/components/button"
import { ContentTitle } from "$/components/content-title"
import { Input } from "$/components/input"
import { LoadingSpinner } from "$/components/loading-spinner"
import { PanelCard } from "$/components/panel-card"
import { Select } from "$/components/select"
import {
  useAdminAPIKeyCreate,
  useAdminAPIKeyRevoke,
  useAdminAPIKeys,
} from "$/hooks/transactions/use-admin"
import { useProjectsList } from "$/hooks/transactions/use-project"
import { useAuthorization } from "$/providers/authorization-provider"
import { extractErrorMessage } from "$/utils/api-error"
import {
  createStateMCPAPIKeySchema,
  mcpAPIKeyScopeLabels,
} from "@openstate/schemas"
import type {
  CreateStateMCPAPIKeyResponse,
  MCPAPIKeyScope,
  StateMCPAPIKeyResponse,
} from "@openstate/types"
import { KeyRoundIcon } from "lucide-react"

type Notice = { type: "success" | "error"; text: string }

type FormState = {
  name: string
  projectIds: string[]
  defaultProjectId: string
  scopes: MCPAPIKeyScope[]
  expiresAt: string
}

type ProjectOption = { label: string; value: string }

type FormErrors = Partial<Record<keyof FormState, string>>

const initialForm: FormState = {
  name: "",
  projectIds: [],
  defaultProjectId: "",
  scopes: ["state:read"],
  expiresAt: "",
}

const formatDate = (value: string | null) => {
  if (!value) return "Never"

  return new Intl.DateTimeFormat(undefined, {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(new Date(value))
}

const statusClassName = (key: StateMCPAPIKeyResponse) =>
  key.revokedAt
    ? "bg-slate-100 text-slate-600"
    : "bg-emerald-100 text-emerald-700"

const statusLabel = (key: StateMCPAPIKeyResponse) =>
  key.revokedAt ? "Revoked" : "Active"

export default function APIKeysPageContent() {
  const authorization = useAuthorization()
  const canRead = authorization.hasPermission("api_key:read")
  const canCreate = authorization.hasPermission("api_key:create")
  const canRevoke = authorization.hasPermission("api_key:revoke")
  const [form, setForm] = useState<FormState>(initialForm)
  const [formErrors, setFormErrors] = useState<FormErrors>({})
  const [notice, setNotice] = useState<Notice | null>(null)
  const [newKey, setNewKey] = useState<string | null>(null)
  const [copied, setCopied] = useState(false)
  const keys = useAdminAPIKeys(authorization.status === "ready" && canRead)
  const create = useAdminAPIKeyCreate()
  const revoke = useAdminAPIKeyRevoke()
  const projects = useProjectsList({
    enabled: authorization.status === "ready" && canCreate,
  })
  const projectOptions: ProjectOption[] = (projects.data ?? []).map(
    (project) => ({
      label: `${project.name} (${project.slug}) — ${project.id}`,
      value: project.id,
    }),
  )
  const selectedProjectOptions = form.projectIds.map(
    (value) =>
      projectOptions.find((option) => option.value === value) ?? {
        label: value,
        value,
      },
  )
  const defaultProjectOption = selectedProjectOptions.find(
    (option) => option.value === form.defaultProjectId,
  )

  if (!canRead) {
    return (
      <div className="space-y-6 p-8">
        <ContentTitle title="State MCP API Keys" />
        <div className="rounded-md bg-amber-50 px-4 py-3 text-sm text-amber-800">
          You are not authorized to manage State MCP API keys.
        </div>
      </div>
    )
  }

  if (keys.isLoading || projects.isLoading) return <LoadingSpinner />

  const updateForm = <K extends keyof FormState>(
    field: K,
    value: FormState[K],
  ) => {
    setForm((current) => ({ ...current, [field]: value }))
    setFormErrors((current) => ({ ...current, [field]: undefined }))
  }

  const toggleScope = (scope: MCPAPIKeyScope) => {
    const scopes = form.scopes.includes(scope)
      ? form.scopes.filter((item) => item !== scope)
      : [...form.scopes, scope]

    updateForm("scopes", scopes)
  }

  const handleCreate = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    setNotice(null)

    const projectIds = form.projectIds
    const defaultProjectId = form.defaultProjectId || projectIds[0]

    if (defaultProjectId && !projectIds.includes(defaultProjectId)) {
      setFormErrors({
        defaultProjectId: "Default project must be one of the project IDs.",
      })
      return
    }

    let expiresAt: string | undefined

    if (form.expiresAt) {
      const expiresDate = new Date(form.expiresAt)

      if (Number.isNaN(expiresDate.getTime())) {
        setFormErrors({ expiresAt: "Enter a valid expiration date." })
        return
      }

      expiresAt = expiresDate.toISOString()
    }

    const parsed = createStateMCPAPIKeySchema.safeParse({
      name: form.name,
      projectIds,
      defaultProjectId,
      scopes: form.scopes,
      expiresAt,
    })

    if (!parsed.success) {
      const errors: FormErrors = {}

      for (const issue of parsed.error.issues) {
        const field = issue.path[0]

        if (
          field === "name" ||
          field === "projectIds" ||
          field === "defaultProjectId" ||
          field === "scopes" ||
          field === "expiresAt"
        ) {
          errors[field] = issue.message
        }
      }

      setFormErrors(errors)
      return
    }

    try {
      const result: CreateStateMCPAPIKeyResponse = await create.mutateAsync(
        parsed.data,
      )
      setNewKey(result.key)
      setCopied(false)
      setForm(initialForm)
      setFormErrors({})
      setNotice({
        type: "success",
        text: "API key created. Copy the secret before dismissing this message.",
      })
    } catch (error) {
      setNotice({
        type: "error",
        text: extractErrorMessage(error) ?? "API key could not be created.",
      })
    }
  }

  const handleCopy = async () => {
    if (!newKey) return

    if (!navigator.clipboard) {
      setNotice({ type: "error", text: "Clipboard access is unavailable." })
      return
    }

    try {
      await navigator.clipboard.writeText(newKey)
      setCopied(true)
    } catch {
      setNotice({ type: "error", text: "The key could not be copied." })
    }
  }

  const handleRevoke = (key: StateMCPAPIKeyResponse) => {
    if (
      key.revokedAt ||
      !window.confirm(
        `Revoke the API key "${key.name}"? This cannot be undone.`,
      )
    ) {
      return
    }

    setNotice(null)
    revoke.mutate(key.id, {
      onSuccess: () => setNotice({ type: "success", text: "API key revoked." }),
      onError: (error) =>
        setNotice({
          type: "error",
          text: extractErrorMessage(error) ?? "API key could not be revoked.",
        }),
    })
  }

  return (
    <div className="space-y-6 p-8">
      <ContentTitle title="State MCP API Keys" />
      <p className="max-w-3xl text-sm text-slate-600">
        Create machine credentials for clients that connect to State MCP at
        <span className="font-mono"> /mcp</span>. The raw secret is shown only
        once after creation.
      </p>

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

      {newKey ? (
        <PanelCard
          title="Your new API key"
          description="Save this secret now. It cannot be retrieved after you dismiss this panel."
        >
          <div className="space-y-4">
            <div className="rounded-md border border-amber-200 bg-amber-50 p-4 text-sm text-amber-950">
              <p className="font-semibold">One-time secret</p>
              <code className="mt-2 block break-all rounded bg-white p-3 font-mono text-xs">
                {newKey}
              </code>
              <p className="mt-3 text-xs text-amber-800">
                Use this as a Bearer token for the State MCP endpoint at
                <span className="font-mono"> http://localhost:8030/mcp</span>.
              </p>
            </div>
            <div className="flex flex-wrap gap-3">
              <Button
                type="button"
                intent="primary"
                onClick={() => void handleCopy()}
              >
                {copied ? "Copied" : "Copy key"}
              </Button>
              <Button
                type="button"
                intent="secondary"
                onClick={() => {
                  setNewKey(null)
                  setCopied(false)
                }}
              >
                Dismiss secret
              </Button>
            </div>
          </div>
        </PanelCard>
      ) : null}

      {canCreate ? (
        <PanelCard
          title="Create API key"
          description="The key will be limited to the selected projects and scopes."
        >
          {projects.isError ? (
            <div className="mb-5 flex items-center justify-between gap-4 rounded-md bg-red-50 px-4 py-3 text-sm text-red-700">
              <span>
                {extractErrorMessage(projects.error) ??
                  "Projects could not be loaded."}
              </span>
              <Button intent="clean" onClick={() => void projects.refetch()}>
                Retry
              </Button>
            </div>
          ) : null}
          <form className="max-w-2xl space-y-5" onSubmit={handleCreate}>
            <Input
              id="api-key-name"
              label="Name"
              required
              placeholder="Production assistant"
              value={form.name}
              error={formErrors.name}
              disabled={create.isPending}
              onChange={(event) => updateForm("name", event.target.value)}
            />
            <Select<ProjectOption, true>
              id="api-key-projects"
              label="Project IDs"
              required
              isMulti
              closeMenuOnSelect={false}
              options={projectOptions}
              value={selectedProjectOptions}
              placeholder="Select one or more projects"
              hint="Projects are limited to the current tenant."
              error={formErrors.projectIds}
              disabled={create.isPending}
              noOptionsMessage={() =>
                projectOptions.length
                  ? "No matching projects."
                  : "No projects found for this tenant."
              }
              getOptionLabel={(option) => option.label}
              getOptionValue={(option) => option.value}
              onChange={(options) => {
                const projectIds = Array.from(
                  new Set(
                    options
                      .map((option) => option.value.trim())
                      .filter(Boolean),
                  ),
                )
                setForm((current) => ({
                  ...current,
                  projectIds,
                  defaultProjectId: projectIds.includes(
                    current.defaultProjectId,
                  )
                    ? current.defaultProjectId
                    : "",
                }))
                setFormErrors((current) => ({
                  ...current,
                  projectIds: undefined,
                  defaultProjectId: undefined,
                }))
              }}
            />
            <Select<ProjectOption>
              id="api-key-default-project"
              label="Default project ID"
              placeholder="Leave empty to use the first project ID"
              hint="The MCP client can omit project when this default is set."
              options={selectedProjectOptions}
              value={defaultProjectOption}
              error={formErrors.defaultProjectId}
              disabled={create.isPending || projectOptions.length === 0}
              isClearable
              getOptionLabel={(option) => option.label}
              getOptionValue={(option) => option.value}
              onChange={(option) =>
                updateForm(
                  "defaultProjectId",
                  (option as ProjectOption | null)?.value ?? "",
                )
              }
            />
            <Input
              id="api-key-expires-at"
              label="Expires at"
              type="datetime-local"
              hint="Optional. Leave empty for a key without an expiration."
              value={form.expiresAt}
              error={formErrors.expiresAt}
              disabled={create.isPending}
              onChange={(event) => updateForm("expiresAt", event.target.value)}
            />
            <fieldset>
              <legend className="text-sm font-medium text-slate-700">
                Scopes <span className="text-danger-500">*</span>
              </legend>
              <div className="mt-2 grid gap-2 sm:grid-cols-3">
                {mcpAPIKeyScopeLabels.map((scope) => (
                  <label
                    key={scope.value}
                    className="flex items-center gap-2 rounded-md border border-slate-200 px-3 py-2 text-sm text-slate-700"
                  >
                    <input
                      type="checkbox"
                      checked={form.scopes.includes(scope.value)}
                      disabled={create.isPending}
                      onChange={() => toggleScope(scope.value)}
                    />
                    {scope.label}
                  </label>
                ))}
              </div>
              {formErrors.scopes ? (
                <p className="mt-1 text-xs text-danger-500">
                  {formErrors.scopes}
                </p>
              ) : null}
            </fieldset>
            <Button type="submit" intent="primary" loading={create.isPending}>
              Create API key
            </Button>
          </form>
        </PanelCard>
      ) : (
        <PanelCard title="Create API key">
          <p className="text-sm text-slate-600">
            You have read-only access to API-key metadata.
          </p>
        </PanelCard>
      )}

      <PanelCard
        title="Existing keys"
        description="Only safe metadata is shown here. The raw secret is never returned again."
      >
        {keys.isError ? (
          <div className="flex items-center justify-between gap-4 rounded-md bg-red-50 px-4 py-3 text-sm text-red-700">
            <span>
              {extractErrorMessage(keys.error) ??
                "API keys could not be loaded."}
            </span>
            <Button intent="clean" onClick={() => void keys.refetch()}>
              Retry
            </Button>
          </div>
        ) : keys.data?.length ? (
          <div className="overflow-x-auto rounded-lg border border-slate-200">
            <table className="min-w-full divide-y divide-slate-200 text-left text-sm">
              <thead className="bg-slate-50 text-xs uppercase text-slate-500">
                <tr>
                  <th className="px-4 py-3">Name</th>
                  <th className="px-4 py-3">Scopes</th>
                  <th className="px-4 py-3">Projects</th>
                  <th className="px-4 py-3">Usage</th>
                  <th className="px-4 py-3">Status</th>
                  <th className="px-4 py-3" />
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-200 bg-white">
                {keys.data.map((key) => (
                  <tr key={key.id}>
                    <td className="px-4 py-4 align-top">
                      <div className="flex items-start gap-2">
                        <KeyRoundIcon
                          className="mt-0.5 shrink-0 text-slate-500"
                          size={16}
                          aria-hidden="true"
                        />
                        <div>
                          <p className="font-medium text-slate-900">
                            {key.name}
                          </p>
                          <p className="mt-1 font-mono text-xs text-slate-500">
                            {key.prefix}_••••••
                          </p>
                        </div>
                      </div>
                    </td>
                    <td className="px-4 py-4 align-top text-slate-600">
                      <ul className="space-y-1">
                        {key.scopes.map((scope) => (
                          <li key={scope} className="font-mono text-xs">
                            {scope}
                          </li>
                        ))}
                      </ul>
                    </td>
                    <td className="max-w-xs px-4 py-4 align-top text-xs text-slate-600">
                      <p>{key.projectIds.length} project(s)</p>
                      <p className="mt-1 break-all font-mono">
                        {key.projectIds.join(", ")}
                      </p>
                      {key.defaultProjectId ? (
                        <p className="mt-1 text-slate-500">
                          Default:{" "}
                          <span className="font-mono">
                            {key.defaultProjectId}
                          </span>
                        </p>
                      ) : null}
                    </td>
                    <td className="px-4 py-4 align-top text-xs text-slate-600">
                      <p>Last used: {formatDate(key.lastUsedAt)}</p>
                      <p className="mt-1">
                        Expires: {formatDate(key.expiresAt)}
                      </p>
                    </td>
                    <td className="px-4 py-4 align-top">
                      <span
                        className={`inline-flex rounded-full px-2.5 py-1 text-xs font-medium ${statusClassName(key)}`}
                      >
                        {statusLabel(key)}
                      </span>
                    </td>
                    <td className="px-4 py-4 text-right align-top">
                      {canRevoke && !key.revokedAt ? (
                        <Button
                          type="button"
                          intent="clean"
                          loading={revoke.isPending}
                          onClick={() => handleRevoke(key)}
                        >
                          Revoke
                        </Button>
                      ) : null}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : (
          <div className="rounded-md border border-dashed border-slate-300 px-4 py-10 text-center">
            <p className="font-medium text-slate-900">No API keys yet</p>
            <p className="mt-1 text-sm text-slate-500">
              Create a key to connect an LLM client to State MCP.
            </p>
          </div>
        )}
      </PanelCard>
    </div>
  )
}
