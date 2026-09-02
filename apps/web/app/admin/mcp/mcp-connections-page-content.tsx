"use client"

import Link from "next/link"
import { useRouter, useSearchParams } from "next/navigation"
import { useState } from "react"

import { Button } from "$/components/button"
import { ContentTitle } from "$/components/content-title"
import { Input } from "$/components/input"
import { LoadingSpinner } from "$/components/loading-spinner"
import { PanelCard } from "$/components/panel-card"
import { Select } from "$/components/select"
import { Textarea } from "$/components/textarea"
import {
  useCreateMCPConnection,
  useDeleteMCPConnection,
  useDiagnoseMCPConnection,
  useDisableMCPConnection,
  useDisconnectMCPOAuth,
  useEnableMCPConnection,
  useListMCPConnections,
  useListMCPTools,
  useRefreshMCPTools,
  useResetMCPConnectionHealth,
  useSetMCPToolEnabled,
  useStartMCPOAuth,
  useTestMCPConnection,
  useUpdateMCPConnection,
} from "$/hooks/transactions/use-mcp-connection"
import { useProjectsList } from "$/hooks/transactions/use-project"
import { useAuthorization } from "$/providers/authorization-provider"
import { extractErrorMessage } from "$/utils/api-error"
import {
  createMCPConnectionSchema,
  mcpConnectionAuthTypes,
  mcpConnectionTransports,
} from "@openstate/schemas"
import type {
  MCPConnectionAuthType,
  MCPConnectionResponse,
  MCPConnectionTransport,
} from "@openstate/types"

type FormState = {
  name: string
  alias: string
  transport: MCPConnectionTransport
  endpoint: string
  stdioProfile: string
  stdioArgs: string
  authType: MCPConnectionAuthType
  credentialReference: string
  credentialValue: string
  oauthAuthorizationEndpoint: string
  oauthTokenEndpoint: string
  oauthClientId: string
  oauthClientSecretValue: string
  oauthScopes: string
  oauthRedirectUri: string
}

const emptyForm: FormState = {
  name: "",
  alias: "",
  transport: "streamable_http",
  endpoint: "",
  stdioProfile: "",
  stdioArgs: "",
  authType: "none",
  credentialReference: "",
  credentialValue: "",
  oauthAuthorizationEndpoint: "",
  oauthTokenEndpoint: "",
  oauthClientId: "",
  oauthClientSecretValue: "",
  oauthScopes: "",
  oauthRedirectUri: "",
}

const transportOptions = mcpConnectionTransports.map((value) => ({
  value,
  label: value === "streamable_http" ? "Streamable HTTP" : value.toUpperCase(),
}))
const authOptions = mcpConnectionAuthTypes.map((value) => ({
  value,
  label: value === "none" ? "No authentication" : value.toUpperCase(),
}))

const schemaSummary = (schema: Record<string, unknown>) => {
  const properties = schema.properties
  if (
    !properties ||
    typeof properties !== "object" ||
    Array.isArray(properties)
  )
    return "No input fields"
  const names = Object.keys(properties)
  return names.length ? names.join(", ") : "No input fields"
}

const shortFingerprint = (fingerprint: string) =>
  fingerprint.length > 16 ? `${fingerprint.slice(0, 16)}…` : fingerprint

export default function MCPConnectionsPageContent() {
  const authorization = useAuthorization()
  const router = useRouter()
  const selectedProjectId = useSearchParams().get("projectId") || ""
  const canRead = authorization.hasPermission("mcp_connection:read")
  const canManage = authorization.hasPermission("mcp_connection:create")
  const projects = useProjectsList({
    enabled: authorization.status === "ready" && canRead,
  })
  const selectedProject = projects.data?.find(
    (project) => project.id === selectedProjectId,
  )
  const connections = useListMCPConnections(
    selectedProjectId,
    authorization.status === "ready" && canRead,
  )
  const create = useCreateMCPConnection()
  const update = useUpdateMCPConnection()
  const enable = useEnableMCPConnection()
  const disable = useDisableMCPConnection()
  const test = useTestMCPConnection()
  const diagnose = useDiagnoseMCPConnection()
  const resetHealth = useResetMCPConnectionHealth()
  const startOAuth = useStartMCPOAuth()
  const disconnectOAuth = useDisconnectMCPOAuth()
  const remove = useDeleteMCPConnection()
  const refreshTools = useRefreshMCPTools()
  const setToolEnabled = useSetMCPToolEnabled()
  const [editing, setEditing] = useState<MCPConnectionResponse | null>(null)
  const [catalogConnectionId, setCatalogConnectionId] = useState<string>()
  const [form, setForm] = useState<FormState>(emptyForm)
  const [notice, setNotice] = useState<{
    type: "success" | "error"
    text: string
  } | null>(null)
  const [formErrors, setFormErrors] = useState<Record<string, string>>({})
  const canDiscover = authorization.hasPermission("mcp_connection:discover")
  const canManageTools = authorization.hasPermission(
    "mcp_connection:tool:update",
  )
  const catalog = useListMCPTools(
    selectedProjectId,
    catalogConnectionId,
    authorization.status === "ready" && canRead,
  )

  if (!canRead) {
    return (
      <div className="space-y-6 p-8">
        <ContentTitle title="MCP Connections" />
        <div className="rounded-md bg-amber-50 px-4 py-3 text-sm text-amber-800">
          You are not authorized to view MCP connections.
        </div>
      </div>
    )
  }

  const openCreate = () => {
    setEditing(null)
    setForm(emptyForm)
    setFormErrors({})
  }

  const openEdit = (connection: MCPConnectionResponse) => {
    setEditing(connection)
    setForm({
      name: connection.name,
      alias: connection.alias,
      transport: connection.transport,
      endpoint: connection.endpoint ?? "",
      stdioProfile: connection.stdioProfile ?? "",
      stdioArgs: connection.stdioArgs.join(", "),
      authType: connection.authType,
      credentialReference: "",
      credentialValue: "",
      oauthAuthorizationEndpoint: connection.oauthAuthorizationEndpoint ?? "",
      oauthTokenEndpoint: connection.oauthTokenEndpoint ?? "",
      oauthClientId: connection.oauthClientId ?? "",
      oauthClientSecretValue: "",
      oauthScopes: connection.oauthScopes.join(", "),
      oauthRedirectUri: connection.oauthRedirectUri ?? "",
    })
    setFormErrors({})
  }

  const save = async () => {
    if (!selectedProjectId) return
    const payload = {
      name: form.name,
      alias: form.alias,
      transport: form.transport,
      endpoint: form.endpoint || undefined,
      stdioProfile: form.stdioProfile || undefined,
      stdioArgs: form.stdioArgs
        .split(",")
        .map((arg) => arg.trim())
        .filter(Boolean),
      authType: form.authType,
      credentialReference: form.credentialReference || undefined,
      credentialValue:
        form.authType === "bearer"
          ? form.credentialValue || undefined
          : undefined,
      oauthAuthorizationEndpoint:
        form.authType === "oauth"
          ? form.oauthAuthorizationEndpoint || undefined
          : undefined,
      oauthTokenEndpoint:
        form.authType === "oauth"
          ? form.oauthTokenEndpoint || undefined
          : undefined,
      oauthClientId:
        form.authType === "oauth" ? form.oauthClientId || undefined : undefined,
      oauthClientSecretValue:
        form.authType === "oauth"
          ? form.oauthClientSecretValue || undefined
          : undefined,
      oauthScopes:
        form.authType === "oauth"
          ? form.oauthScopes
              .split(",")
              .map((scope) => scope.trim())
              .filter(Boolean)
          : undefined,
      oauthRedirectUri:
        form.authType === "oauth"
          ? form.oauthRedirectUri || undefined
          : undefined,
    }
    const parsed = createMCPConnectionSchema.safeParse(payload)
    if (!parsed.success) {
      const errors: Record<string, string> = {}
      for (const issue of parsed.error.issues)
        errors[String(issue.path[0])] = issue.message
      setFormErrors(errors)
      return
    }
    setFormErrors({})
    try {
      if (editing)
        await update.mutateAsync({
          projectId: selectedProjectId,
          id: editing.id,
          payload: parsed.data,
        })
      else
        await create.mutateAsync({
          projectId: selectedProjectId,
          payload: parsed.data,
        })
      setNotice({
        type: "success",
        text: editing ? "MCP connection updated." : "MCP connection created.",
      })
      setEditing(null)
      setForm(emptyForm)
    } catch (error) {
      setNotice({
        type: "error",
        text:
          extractErrorMessage(error) || "Could not save the MCP connection.",
      })
    }
  }

  const runAction = async (
    action:
      | "enable"
      | "disable"
      | "test"
      | "diagnose"
      | "resetHealth"
      | "delete",
    connection: MCPConnectionResponse,
  ) => {
    if (!selectedProjectId) return
    if (
      (action === "disable" || action === "delete") &&
      !window.confirm(
        `${action === "delete" ? "Delete" : "Disable"} ${connection.name}?`,
      )
    )
      return
    try {
      const variables = { projectId: selectedProjectId, id: connection.id }
      if (action === "enable") await enable.mutateAsync(variables)
      if (action === "disable") await disable.mutateAsync(variables)
      if (action === "test") await test.mutateAsync(variables)
      if (action === "diagnose") await diagnose.mutateAsync(variables)
      if (action === "resetHealth") await resetHealth.mutateAsync(variables)
      if (action === "delete") await remove.mutateAsync(variables)
      setNotice({
        type: "success",
        text:
          action === "test"
            ? "Handshake test completed."
            : action === "diagnose"
              ? "Connection diagnostics completed."
              : action === "resetHealth"
                ? "Health state reset."
                : `MCP connection ${action}d.`,
      })
    } catch (error) {
      setNotice({
        type: "error",
        text:
          extractErrorMessage(error) ||
          `Could not ${action} the MCP connection.`,
      })
    }
  }

  const connectOAuth = async (connection: MCPConnectionResponse) => {
    if (!selectedProjectId) return
    try {
      const result = await startOAuth.mutateAsync({
        projectId: selectedProjectId,
        id: connection.id,
      })
      window.location.assign(result.authorizationUrl)
    } catch (error) {
      setNotice({
        type: "error",
        text:
          extractErrorMessage(error) || "Could not start OAuth authorization.",
      })
    }
  }

  const disconnectConnectionOAuth = async (
    connection: MCPConnectionResponse,
  ) => {
    if (
      !selectedProjectId ||
      !window.confirm(`Disconnect OAuth for ${connection.name}?`)
    )
      return
    try {
      await disconnectOAuth.mutateAsync({
        projectId: selectedProjectId,
        id: connection.id,
      })
      setNotice({ type: "success", text: "OAuth disconnected." })
    } catch (error) {
      setNotice({
        type: "error",
        text: extractErrorMessage(error) || "Could not disconnect OAuth.",
      })
    }
  }

  const toggleCatalog = (connectionId: string) => {
    setCatalogConnectionId((current) =>
      current === connectionId ? undefined : connectionId,
    )
  }

  const runRefresh = async (connection: MCPConnectionResponse) => {
    if (!selectedProjectId) return
    setCatalogConnectionId(connection.id)
    try {
      await refreshTools.mutateAsync({
        projectId: selectedProjectId,
        id: connection.id,
      })
      setNotice({ type: "success", text: "MCP tool catalog refreshed." })
    } catch (error) {
      setNotice({
        type: "error",
        text:
          extractErrorMessage(error) ||
          "Could not refresh the MCP tool catalog.",
      })
    }
  }

  const runToolEnablement = async (
    connection: MCPConnectionResponse,
    toolName: string,
    enabled: boolean,
  ) => {
    if (!selectedProjectId) return
    try {
      await setToolEnabled.mutateAsync({
        projectId: selectedProjectId,
        id: connection.id,
        toolName,
        enabled,
      })
      setNotice({
        type: "success",
        text: `Tool ${enabled ? "enabled" : "disabled"}.`,
      })
    } catch (error) {
      setNotice({
        type: "error",
        text: extractErrorMessage(error) || "Could not update tool enablement.",
      })
    }
  }

  const setField = <K extends keyof FormState>(field: K, value: FormState[K]) =>
    setForm((current) => ({ ...current, [field]: value }))
  const formPending = create.isPending || update.isPending

  return (
    <div className="space-y-6 p-8">
      <ContentTitle title="MCP Connections" />
      <p className="max-w-3xl text-sm text-slate-600">
        Register the external MCP servers owned by the selected project. State
        uses this registry as its controlled provider boundary; tool discovery
        and invocation are separate phases.
      </p>

      {notice && (
        <div
          className={`rounded-md px-4 py-3 text-sm ${notice.type === "success" ? "bg-green-50 text-green-700" : "bg-red-50 text-red-700"}`}
        >
          {notice.text}
        </div>
      )}

      <PanelCard
        title="Project scope"
        description="Connections are isolated by tenant and project."
      >
        <div className="flex flex-wrap items-end gap-4">
          <div className="min-w-[280px] flex-1">
            <label
              htmlFor="mcp-project"
              className="mb-1.5 block text-sm font-medium text-slate-700"
            >
              Project
            </label>
            <select
              id="mcp-project"
              className="w-full rounded-md border border-slate-300 bg-white px-3 py-2 text-sm"
              value={selectedProjectId}
              onChange={(event) =>
                router.push(
                  `/admin/mcp?projectId=${encodeURIComponent(event.target.value)}`,
                )
              }
            >
              <option value="">Select a project</option>
              {projects.data?.map((project) => (
                <option key={project.id} value={project.id}>
                  {project.name} ({project.slug})
                </option>
              ))}
            </select>
          </div>
          {selectedProject && (
            <Link
              className="text-sm text-blue-700 underline"
              href={`/admin/projects?projectId=${encodeURIComponent(selectedProject.id)}`}
            >
              Open project flow
            </Link>
          )}
        </div>
      </PanelCard>

      {!selectedProjectId ? (
        <PanelCard title="Choose a project">
          <p className="text-sm text-slate-600">
            Select a project to view and manage its external MCP connections.
          </p>
        </PanelCard>
      ) : (
        <>
          <PanelCard
            title={editing ? "Edit MCP connection" : "Register MCP connection"}
            description="Only the credential status is returned after saving; secret values are never read back."
            action={
              canManage ? (
                <Button intent="clean" onClick={openCreate}>
                  New connection
                </Button>
              ) : undefined
            }
          >
            {!canManage && (
              <p className="mb-4 rounded-md bg-slate-50 px-4 py-3 text-sm text-slate-600">
                You have read-only access to this project.
              </p>
            )}
            <div className="grid gap-4 md:grid-cols-2">
              <Input
                label="Name"
                required
                disabled={!canManage}
                value={form.name}
                error={formErrors.name}
                onChange={(event) => setField("name", event.target.value)}
                placeholder="Padel provider"
              />
              <Input
                label="Alias"
                required
                disabled={!canManage}
                value={form.alias}
                error={formErrors.alias}
                onChange={(event) =>
                  setField("alias", event.target.value.toLowerCase())
                }
                placeholder="padel-provider"
                hint="Unique within this project."
              />
              <Select
                label="Transport"
                required
                disabled={!canManage}
                options={transportOptions}
                getOptionLabel={(option) => option.label}
                getOptionValue={(option) => option.value}
                value={transportOptions.find(
                  (option) => option.value === form.transport,
                )}
                onChange={(option) =>
                  setField(
                    "transport",
                    (option as (typeof transportOptions)[number]).value,
                  )
                }
              />
              {form.transport === "stdio" ? (
                <>
                  <Input
                    label="Trusted STDIO profile"
                    required
                    disabled={!canManage}
                    value={form.stdioProfile}
                    error={formErrors.stdioProfile}
                    onChange={(event) =>
                      setField("stdioProfile", event.target.value)
                    }
                    placeholder="trusted-padel"
                    hint="The server must be approved by deployment configuration."
                  />
                  <Textarea
                    label="Additional arguments"
                    disabled={!canManage}
                    value={form.stdioArgs}
                    onChange={(event) =>
                      setField("stdioArgs", event.target.value)
                    }
                    placeholder="--mode, production"
                    hint="Comma-separated; no shell commands."
                  />
                </>
              ) : (
                <Input
                  label="MCP URL"
                  required
                  disabled={!canManage}
                  value={form.endpoint}
                  error={formErrors.endpoint}
                  onChange={(event) => setField("endpoint", event.target.value)}
                  placeholder="https://provider.example.com/mcp"
                />
              )}
              <Select
                label="Authentication"
                required
                disabled={!canManage}
                options={authOptions}
                getOptionLabel={(option) => option.label}
                getOptionValue={(option) => option.value}
                value={authOptions.find(
                  (option) => option.value === form.authType,
                )}
                onChange={(option) =>
                  setField(
                    "authType",
                    (option as (typeof authOptions)[number]).value,
                  )
                }
              />
              {form.authType === "bearer" && (
                <Input
                  label="Bearer credential (write-only)"
                  type="password"
                  disabled={!canManage}
                  value={form.credentialValue}
                  error={formErrors.credentialValue}
                  onChange={(event) =>
                    setField("credentialValue", event.target.value)
                  }
                  placeholder={
                    editing
                      ? "Leave empty to keep current credential"
                      : "Paste provider bearer token"
                  }
                  hint="The secret is stored server-side and is never returned to this page."
                />
              )}
            </div>
            {form.authType === "oauth" && (
              <div className="mt-4 grid gap-4 border-t border-slate-200 pt-4 md:grid-cols-2">
                <Input
                  label="Authorization endpoint"
                  required
                  disabled={!canManage}
                  value={form.oauthAuthorizationEndpoint}
                  onChange={(event) =>
                    setField("oauthAuthorizationEndpoint", event.target.value)
                  }
                  placeholder="https://provider.example.com/oauth/authorize"
                />
                <Input
                  label="Token endpoint"
                  required
                  disabled={!canManage}
                  value={form.oauthTokenEndpoint}
                  onChange={(event) =>
                    setField("oauthTokenEndpoint", event.target.value)
                  }
                  placeholder="https://provider.example.com/oauth/token"
                />
                <Input
                  label="Client ID"
                  required
                  disabled={!canManage}
                  value={form.oauthClientId}
                  onChange={(event) =>
                    setField("oauthClientId", event.target.value)
                  }
                  placeholder="provider-client-id"
                />
                <Input
                  label="Client secret (write-only)"
                  type="password"
                  disabled={!canManage}
                  value={form.oauthClientSecretValue}
                  onChange={(event) =>
                    setField("oauthClientSecretValue", event.target.value)
                  }
                  placeholder={
                    editing
                      ? "Leave empty to keep current secret"
                      : "Paste client secret"
                  }
                />
                <Input
                  label="Redirect URI"
                  required
                  disabled={!canManage}
                  value={form.oauthRedirectUri}
                  onChange={(event) =>
                    setField("oauthRedirectUri", event.target.value)
                  }
                  placeholder="https://app.example.com/oauth/callback"
                />
                <Input
                  label="Scopes"
                  disabled={!canManage}
                  value={form.oauthScopes}
                  onChange={(event) =>
                    setField("oauthScopes", event.target.value)
                  }
                  placeholder="calendar.read, booking.write"
                  hint="Comma-separated scopes."
                />
              </div>
            )}
            {canManage && (
              <div className="mt-5 flex gap-2">
                <Button
                  intent="primary"
                  loading={formPending}
                  onClick={() => void save()}
                >
                  {editing ? "Save changes" : "Create connection"}
                </Button>
                {editing && (
                  <Button intent="clean" onClick={openCreate}>
                    Cancel
                  </Button>
                )}
              </div>
            )}
          </PanelCard>

          <PanelCard
            title="Registered connections"
            description="Only connections belonging to this project are listed."
          >
            {connections.isLoading ? (
              <LoadingSpinner />
            ) : connections.isError ? (
              <div className="rounded-md bg-red-50 px-4 py-3 text-sm text-red-700">
                {extractErrorMessage(connections.error) ||
                  "Connections could not be loaded."}
              </div>
            ) : connections.data?.length ? (
              <div className="space-y-3">
                {connections.data.map((connection) => (
                  <article
                    key={connection.id}
                    className="rounded-lg border border-slate-200 p-4"
                  >
                    <div className="flex flex-wrap items-start justify-between gap-3">
                      <div>
                        <h3 className="font-semibold text-slate-950">
                          {connection.name}
                        </h3>
                        <p className="font-mono text-xs text-slate-500">
                          {connection.alias}
                        </p>
                        <div className="mt-2 flex flex-wrap gap-2 text-xs">
                          <span className="rounded-full bg-slate-100 px-2 py-1 text-slate-700">
                            {connection.transport}
                          </span>
                          <span className="rounded-full bg-blue-50 px-2 py-1 text-blue-700">
                            auth: {connection.authType}
                          </span>
                          <span
                            className={`rounded-full px-2 py-1 ${connection.status === "enabled" ? "bg-green-100 text-green-700" : "bg-slate-100 text-slate-500"}`}
                          >
                            {connection.status}
                          </span>
                          <span
                            className={`rounded-full px-2 py-1 ${connection.credentialStatus === "configured" ? "bg-green-50 text-green-700" : "bg-amber-50 text-amber-700"}`}
                          >
                            credential: {connection.credentialStatus}
                          </span>
                          {connection.authType === "oauth" && (
                            <span
                              className={`rounded-full px-2 py-1 ${connection.oauthStatus === "connected" ? "bg-green-50 text-green-700" : "bg-amber-50 text-amber-700"}`}
                            >
                              OAuth: {connection.oauthStatus}
                            </span>
                          )}
                          <span
                            className={`rounded-full px-2 py-1 ${connection.healthStatus === "healthy" ? "bg-green-50 text-green-700" : connection.healthStatus === "unknown" ? "bg-slate-100 text-slate-600" : "bg-amber-50 text-amber-700"}`}
                          >
                            health: {connection.healthStatus}
                          </span>
                          <span
                            className={`rounded-full px-2 py-1 ${connection.lastTestStatus === "ready" ? "bg-green-50 text-green-700" : "bg-slate-100 text-slate-600"}`}
                          >
                            test: {connection.lastTestStatus}
                          </span>
                        </div>
                        {connection.lastTestErrorCode && (
                          <p className="mt-2 text-xs text-red-600">
                            {connection.lastTestErrorCode}
                          </p>
                        )}
                      </div>
                      <div className="flex flex-wrap justify-end gap-2">
                        <Button
                          intent="clean"
                          onClick={() => openEdit(connection)}
                          disabled={!canManage}
                        >
                          Edit
                        </Button>
                        <Button
                          intent="clean"
                          onClick={() => void runAction("test", connection)}
                          disabled={!canManage || test.isPending}
                        >
                          Test
                        </Button>
                        <Button
                          intent="clean"
                          onClick={() => void runAction("diagnose", connection)}
                          disabled={!canManage || diagnose.isPending}
                        >
                          Diagnose
                        </Button>
                        <Button
                          intent="clean"
                          onClick={() =>
                            void runAction("resetHealth", connection)
                          }
                          disabled={!canManage || resetHealth.isPending}
                        >
                          Reset health
                        </Button>
                        {connection.authType === "oauth" &&
                          (connection.oauthStatus === "connected" ? (
                            <Button
                              intent="clean"
                              onClick={() =>
                                void disconnectConnectionOAuth(connection)
                              }
                              disabled={!canManage || disconnectOAuth.isPending}
                            >
                              Disconnect OAuth
                            </Button>
                          ) : (
                            <Button
                              intent="clean"
                              onClick={() => void connectOAuth(connection)}
                              disabled={!canManage || startOAuth.isPending}
                            >
                              Connect OAuth
                            </Button>
                          ))}
                        {connection.status === "enabled" ? (
                          <Button
                            intent="clean"
                            onClick={() =>
                              void runAction("disable", connection)
                            }
                            disabled={!canManage || disable.isPending}
                          >
                            Disable
                          </Button>
                        ) : (
                          <Button
                            intent="clean"
                            onClick={() => void runAction("enable", connection)}
                            disabled={!canManage || enable.isPending}
                          >
                            Enable
                          </Button>
                        )}
                        <Button
                          intent="clean"
                          onClick={() => toggleCatalog(connection.id)}
                          disabled={
                            catalog.isLoading &&
                            catalogConnectionId === connection.id
                          }
                        >
                          {catalogConnectionId === connection.id
                            ? "Hide tools"
                            : "View tools"}
                        </Button>
                        <Button
                          intent="danger"
                          onClick={() => void runAction("delete", connection)}
                          disabled={!canManage || remove.isPending}
                        >
                          Delete
                        </Button>
                      </div>
                    </div>
                    {catalogConnectionId === connection.id && (
                      <div className="mt-4 border-t border-slate-200 pt-4">
                        <div className="flex flex-wrap items-start justify-between gap-3">
                          <div>
                            <h4 className="font-medium text-slate-900">
                              Discovered tools
                            </h4>
                            <p className="text-xs text-slate-500">
                              This view uses stored metadata. Refresh explicitly
                              to contact the provider with initialize +
                              tools/list.
                            </p>
                          </div>
                          {canDiscover && (
                            <Button
                              intent="clean"
                              loading={refreshTools.isPending}
                              onClick={() => void runRefresh(connection)}
                              disabled={connection.status !== "enabled"}
                            >
                              Refresh catalog
                            </Button>
                          )}
                        </div>
                        {catalog.isLoading ? (
                          <div className="mt-4">
                            <LoadingSpinner />
                          </div>
                        ) : catalog.isError ? (
                          <div className="mt-4 rounded-md bg-red-50 px-4 py-3 text-sm text-red-700">
                            {extractErrorMessage(catalog.error) ||
                              "Tool catalog could not be loaded."}
                          </div>
                        ) : catalog.data ? (
                          <div className="mt-4 space-y-3">
                            {catalog.data.latestRun?.status === "failed" && (
                              <div className="rounded-md bg-amber-50 px-4 py-3 text-sm text-amber-800">
                                Latest refresh failed (
                                {catalog.data.latestRun.errorCode ||
                                  "classified provider error"}
                                ).{" "}
                                {catalog.data.lastSuccessfulRun && (
                                  <>
                                    Last successful discovery:{" "}
                                    {new Date(
                                      catalog.data.lastSuccessfulRun
                                        .completedAt,
                                    ).toLocaleString()}
                                    .
                                  </>
                                )}
                              </div>
                            )}
                            {!catalog.data.tools.length ? (
                              <div className="rounded-md border border-dashed border-slate-300 px-4 py-6 text-sm text-slate-500">
                                No tools have been discovered yet.
                              </div>
                            ) : (
                              <div className="overflow-x-auto rounded-md border border-slate-200">
                                <table className="min-w-full divide-y divide-slate-200 text-sm">
                                  <thead className="bg-slate-50 text-left text-xs uppercase tracking-wide text-slate-500">
                                    <tr>
                                      <th className="px-3 py-2">Tool</th>
                                      <th className="px-3 py-2">
                                        Input schema
                                      </th>
                                      <th className="px-3 py-2">Drift</th>
                                      <th className="px-3 py-2">Fingerprint</th>
                                      <th className="px-3 py-2">Status</th>
                                      <th className="px-3 py-2" />
                                    </tr>
                                  </thead>
                                  <tbody className="divide-y divide-slate-200">
                                    {catalog.data.tools.map((tool) => (
                                      <tr key={tool.id}>
                                        <td className="px-3 py-3 align-top">
                                          <div className="font-mono text-xs font-medium text-slate-900">
                                            {tool.name}
                                          </div>
                                          <div className="mt-1 max-w-xs text-xs text-slate-600">
                                            {tool.description ||
                                              "No description"}
                                          </div>
                                        </td>
                                        <td className="px-3 py-3 align-top text-xs text-slate-600">
                                          {schemaSummary(tool.inputSchema)}
                                        </td>
                                        <td className="px-3 py-3 align-top">
                                          <span
                                            className={`rounded-full px-2 py-1 text-xs ${tool.driftStatus === "changed" || tool.driftStatus === "removed" ? "bg-amber-100 text-amber-800" : "bg-slate-100 text-slate-700"}`}
                                          >
                                            {tool.driftStatus}
                                          </span>
                                        </td>
                                        <td
                                          className="px-3 py-3 align-top font-mono text-xs text-slate-500"
                                          title={tool.fingerprint}
                                        >
                                          {shortFingerprint(tool.fingerprint)}
                                        </td>
                                        <td className="px-3 py-3 align-top">
                                          <span
                                            className={`rounded-full px-2 py-1 text-xs ${tool.enabled && tool.availability === "present" ? "bg-green-100 text-green-700" : "bg-slate-100 text-slate-600"}`}
                                          >
                                            {tool.availability === "present" &&
                                            tool.enabled
                                              ? "enabled"
                                              : tool.availability}
                                          </span>
                                        </td>
                                        <td className="px-3 py-3 text-right align-top">
                                          {canManageTools &&
                                            tool.availability === "present" && (
                                              <Button
                                                intent="clean"
                                                loading={
                                                  setToolEnabled.isPending
                                                }
                                                onClick={() =>
                                                  void runToolEnablement(
                                                    connection,
                                                    tool.name,
                                                    !tool.enabled,
                                                  )
                                                }
                                              >
                                                {tool.enabled
                                                  ? "Disable"
                                                  : "Enable"}
                                              </Button>
                                            )}
                                        </td>
                                      </tr>
                                    ))}
                                  </tbody>
                                </table>
                              </div>
                            )}
                          </div>
                        ) : null}
                      </div>
                    )}
                  </article>
                ))}
              </div>
            ) : (
              <div className="rounded-md border border-dashed border-slate-300 px-4 py-10 text-center">
                <p className="font-medium text-slate-900">No MCP connections</p>
                <p className="mt-1 text-sm text-slate-500">
                  Register the first external MCP provider for this project.
                </p>
              </div>
            )}
          </PanelCard>
        </>
      )}
    </div>
  )
}
