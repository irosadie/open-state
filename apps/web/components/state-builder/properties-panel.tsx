"use client"

import { useCapabilitiesList } from "$/hooks/transactions/use-capability"
import {
  useDeleteProjectMCPBinding,
  useListProjectMCPBindings,
  useListProjectMCPToolOptions,
  useUpsertProjectMCPBinding,
} from "$/hooks/transactions/use-project-mcp-binding"
import type {
  CapabilityResponse,
  MCPToolOptionResponse,
  ProjectCapabilityMCPBindingHealth,
  ProjectCapabilityMCPBindingResponse,
} from "@openstate/types"
import { Boxes, GitBranch, MousePointerClick, Plus, Trash2 } from "lucide-react"
import Link from "next/link"
import type React from "react"
import { type ReactNode, useMemo, useState } from "react"
import { nodeTypeList } from "./nodes/nodes"
import { useStateBuilderStore } from "./state-builder.store"
import type {
  GuardCondition,
  GuardGroup,
  GuardOperator,
} from "./types/workflow"
import { createGuardCondition, createGuardGroup } from "./types/workflow.utils"

/* ------------------------------------------------------------------ */
/* Komponen form primitif kecil (reusable)                             */
/* ------------------------------------------------------------------ */

function Field({ label, children }: { label: string; children: ReactNode }) {
  return (
    <div className="flex flex-col gap-1">
      <span className="text-xs font-medium text-slate-500">{label}</span>
      {children}
    </div>
  )
}

function TextInput(props: React.InputHTMLAttributes<HTMLInputElement>) {
  return (
    <input
      {...props}
      className="rounded-md border border-slate-200 bg-white px-2.5 py-1.5 text-sm text-slate-800 outline-none focus:border-primary-400 focus:ring-1 focus:ring-primary-200"
    />
  )
}

function TextArea(props: React.TextareaHTMLAttributes<HTMLTextAreaElement>) {
  return (
    <textarea
      {...props}
      className="min-h-[4.5rem] resize-y rounded-md border border-slate-200 bg-white px-2.5 py-1.5 text-sm text-slate-800 outline-none focus:border-primary-400 focus:ring-1 focus:ring-primary-200"
    />
  )
}

function Section({
  title,
  children,
  action,
}: {
  title: string
  children: ReactNode
  action?: ReactNode
}) {
  return (
    <section className="border-t border-slate-200 px-3 py-3 first:border-t-0">
      <div className="mb-2 flex items-center justify-between">
        <h3 className="text-xs font-semibold uppercase tracking-wide text-slate-400">
          {title}
        </h3>
        {action}
      </div>
      <div className="flex flex-col gap-2">{children}</div>
    </section>
  )
}

function AddButton({ onClick, label }: { onClick: () => void; label: string }) {
  return (
    <button
      type="button"
      onClick={onClick}
      className="flex items-center gap-1 text-xs font-medium text-primary-600 hover:text-primary-700"
    >
      <Plus className="h-3.5 w-3.5" /> {label}
    </button>
  )
}

function TagInput({
  values,
  placeholder,
  onChange,
}: {
  values: string[]
  placeholder?: string
  onChange: (values: string[]) => void
}) {
  const [draft, setDraft] = useState("")

  const commit = () => {
    const v = draft.trim()
    if (v && !values.includes(v)) {
      onChange([...values, v])
    }
    setDraft("")
  }

  return (
    <div className="flex flex-wrap items-center gap-1.5 rounded-md border border-slate-200 bg-white px-2 py-1.5">
      {values.map((v) => (
        <span
          key={v}
          className="flex items-center gap-1 rounded bg-primary-50 px-1.5 py-0.5 text-xs font-medium text-primary-700"
        >
          {v}
          <button
            type="button"
            onClick={() => onChange(values.filter((x) => x !== v))}
            className="text-primary-400 hover:text-primary-600"
          >
            <Trash2 className="h-3 w-3" />
          </button>
        </span>
      ))}
      <input
        value={draft}
        onChange={(e) => setDraft(e.target.value)}
        onKeyDown={(e) => {
          if (e.key === "Enter" || e.key === ",") {
            e.preventDefault()
            commit()
          }
        }}
        onBlur={commit}
        placeholder={values.length === 0 ? placeholder : ""}
        className="min-w-[80px] flex-1 bg-transparent text-sm text-slate-800 outline-none placeholder:text-slate-400"
      />
    </div>
  )
}

/* ------------------------------------------------------------------ */
/* Guard editor (dipakai untuk guard node & edge)                      */
/* ------------------------------------------------------------------ */

function GuardsEditor({
  guards,
  onChange,
}: {
  guards: GuardGroup[]
  onChange: (guards: GuardGroup[]) => void
}) {
  const addGroup = () => onChange([...guards, createGuardGroup()])
  const addCondition = (gi: number) => {
    const next = guards.map((g, i) =>
      i === gi
        ? { ...g, conditions: [...g.conditions, createGuardCondition()] }
        : g,
    )
    onChange(next)
  }
  const updateCondition = (
    gi: number,
    ci: number,
    patch: Partial<GuardCondition>,
  ) => {
    const next = guards.map((g, i) =>
      i === gi
        ? {
            ...g,
            conditions: g.conditions.map((c, j) =>
              j === ci ? { ...c, ...patch } : c,
            ),
          }
        : g,
    )
    onChange(next)
  }
  const removeCondition = (gi: number, ci: number) => {
    if (ci < 0) {
      // hapus seluruh group
      onChange(guards.filter((_, i) => i !== gi))
      return
    }
    const next = guards.map((g, i) =>
      i === gi
        ? { ...g, conditions: g.conditions.filter((_, j) => j !== ci) }
        : g,
    )
    onChange(next)
  }

  if (guards.length === 0) {
    return (
      <div className="flex flex-col gap-1">
        <p className="text-xs text-slate-400">Belum ada guard.</p>
        <AddButton label="Group" onClick={addGroup} />
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-3">
      {guards.map((group, gi) => (
        <div
          key={group.id ?? gi}
          className="rounded-md border border-slate-200 bg-white p-2"
        >
          <div className="mb-1.5 flex items-center justify-between">
            <div className="flex items-center gap-1.5">
              <span className="text-[10px] font-semibold uppercase text-slate-400">
                Group {gi + 1}
              </span>
              <select
                value={group.logic}
                onChange={(e) =>
                  onChange(
                    guards.map((g, i) =>
                      i === gi
                        ? { ...g, logic: e.target.value as "AND" | "OR" }
                        : g,
                    ),
                  )
                }
                className="rounded border border-slate-200 px-1 py-0.5 text-xs text-slate-600"
              >
                <option value="AND">AND</option>
                <option value="OR">OR</option>
              </select>
            </div>
            <button
              type="button"
              onClick={() => removeCondition(gi, -1)}
              className="text-[10px] text-danger-400 hover:text-danger-600"
            >
              hapus group
            </button>
          </div>
          {group.conditions.map((cond, ci) => (
            <div key={cond.id ?? ci} className="mb-1 flex items-center gap-1">
              <TextInput
                value={cond.field}
                placeholder="field"
                onChange={(e) =>
                  updateCondition(gi, ci, { field: e.target.value })
                }
                className="!w-[42%] !py-1"
              />
              <select
                value={cond.operator}
                onChange={(e) =>
                  updateCondition(gi, ci, {
                    operator: e.target.value as GuardOperator,
                  })
                }
                className="rounded-md border border-slate-200 px-1 py-1 text-xs text-slate-600"
              >
                {[
                  "==",
                  "!=",
                  ">",
                  ">=",
                  "<",
                  "<=",
                  "IN",
                  "NOT_IN",
                  "EXISTS",
                  "NOT_EXISTS",
                ].map((op) => (
                  <option key={op} value={op}>
                    {op}
                  </option>
                ))}
              </select>
              {["EXISTS", "NOT_EXISTS"].includes(cond.operator) ? null : (
                <TextInput
                  value={cond.value ?? ""}
                  placeholder="value"
                  onChange={(e) =>
                    updateCondition(gi, ci, { value: e.target.value })
                  }
                  className="!w-[30%] !py-1"
                />
              )}
              <button
                type="button"
                onClick={() => removeCondition(gi, ci)}
                className="text-slate-300 hover:text-danger-500"
              >
                <Trash2 className="h-3.5 w-3.5" />
              </button>
            </div>
          ))}
          <AddButton label="Condition" onClick={() => addCondition(gi)} />
        </div>
      ))}
    </div>
  )
}

/* ------------------------------------------------------------------ */
/* Panel untuk NODE (state)                                            */
/* ------------------------------------------------------------------ */

type NodePanelProps = {
  nodeId: string
  projectId?: string
  capabilities: CapabilityResponse[]
  toolOptions: MCPToolOptionResponse[]
  bindings: ProjectCapabilityMCPBindingResponse[]
  capabilitiesLoading: boolean
  catalogLoading: boolean
  catalogError: boolean
  bindingPending: boolean
  onBind: (capabilityId: string, connectionId: string, toolId: string) => void
  onUnbind: (capabilityId: string) => void
}

function healthLabel(health: ProjectCapabilityMCPBindingHealth) {
  return health.replaceAll("_", " ")
}

function healthClass(health: ProjectCapabilityMCPBindingHealth) {
  if (health === "ACTIVE") return "text-emerald-600"
  if (health === "MISSING_MAPPING") return "text-amber-600"
  return "text-danger-600"
}

function MCPBindingsEditor({
  projectId,
  nodeCapabilities,
  capabilities,
  toolOptions,
  bindings,
  capabilitiesLoading,
  catalogLoading,
  catalogError,
  bindingPending,
  onChange,
  onBind,
  onUnbind,
}: {
  projectId?: string
  nodeCapabilities: string[]
  capabilities: CapabilityResponse[]
  toolOptions: MCPToolOptionResponse[]
  bindings: ProjectCapabilityMCPBindingResponse[]
  capabilitiesLoading: boolean
  catalogLoading: boolean
  catalogError: boolean
  bindingPending: boolean
  onChange: (values: string[]) => void
  onBind: (capabilityId: string, connectionId: string, toolId: string) => void
  onUnbind: (capabilityId: string) => void
}) {
  const [selectedCapability, setSelectedCapability] = useState("")
  const [connectionSelection, setConnectionSelection] = useState<
    Record<string, string>
  >({})
  const connectionChoices = useMemo(() => {
    const choices = new Map<string, MCPToolOptionResponse>()
    for (const option of toolOptions) {
      if (!choices.has(option.connectionId))
        choices.set(option.connectionId, option)
    }
    for (const binding of bindings) {
      if (
        binding.connectionId &&
        binding.connectionAlias &&
        !choices.has(binding.connectionId)
      ) {
        choices.set(binding.connectionId, {
          connectionId: binding.connectionId,
          connectionName: binding.connectionName ?? binding.connectionAlias,
          connectionAlias: binding.connectionAlias,
          connectionStatus: binding.connectionStatus ?? "disabled",
          toolId: binding.toolId ?? "",
          toolName: binding.toolName ?? "",
          toolTitle: binding.toolTitle ?? null,
          toolDescription: binding.toolDescription ?? "",
          inputSchema: {},
          toolFingerprint: binding.currentToolFingerprint ?? "",
        })
      }
    }
    return [...choices.values()]
  }, [bindings, toolOptions])

  const addCapability = () => {
    if (!selectedCapability || nodeCapabilities.includes(selectedCapability))
      return
    onChange([...nodeCapabilities, selectedCapability])
    setSelectedCapability("")
  }

  return (
    <div className="flex flex-col gap-2">
      {!projectId ? (
        <p className="rounded-md bg-amber-50 px-2.5 py-2 text-xs text-amber-700">
          Save this workflow in a project before binding MCP tools.
        </p>
      ) : null}
      <div className="flex gap-1.5">
        <select
          aria-label="MCP capability"
          value={selectedCapability}
          onChange={(event) => setSelectedCapability(event.target.value)}
          disabled={capabilitiesLoading}
          className="min-w-0 flex-1 rounded-md border border-slate-200 bg-white px-2 py-1.5 text-xs text-slate-700"
        >
          <option value="">
            {capabilitiesLoading
              ? "Loading capabilities…"
              : "Select capability"}
          </option>
          {capabilities
            .filter((capability) => !nodeCapabilities.includes(capability.name))
            .map((capability) => (
              <option key={capability.id} value={capability.name}>
                {capability.name}
              </option>
            ))}
        </select>
        <button
          type="button"
          onClick={addCapability}
          disabled={!selectedCapability}
          className="rounded-md bg-primary-600 px-2.5 py-1.5 text-xs font-medium text-white disabled:cursor-not-allowed disabled:opacity-40"
        >
          Add
        </button>
      </div>

      {!catalogLoading &&
      !catalogError &&
      projectId &&
      toolOptions.length === 0 ? (
        <p className="rounded-md bg-slate-50 px-2.5 py-2 text-xs leading-relaxed text-slate-500">
          No enabled MCP tools are available for this project. Configure and
          test a connection in{" "}
          <Link className="font-medium text-primary-600" href="/admin/mcp">
            MCP Connections
          </Link>
          .
        </p>
      ) : null}
      {catalogError ? (
        <p className="rounded-md bg-danger-50 px-2.5 py-2 text-xs text-danger-600">
          MCP tool catalog could not be loaded. Check MCP Connections.
        </p>
      ) : null}

      {nodeCapabilities.length === 0 ? (
        <p className="text-xs text-slate-400">No MCP capabilities selected.</p>
      ) : (
        nodeCapabilities.map((capabilityName) => {
          const capability = capabilities.find(
            (item) => item.name === capabilityName,
          )
          const binding = capability
            ? bindings.find((item) => item.capabilityId === capability.id)
            : undefined
          const selectedConnectionId =
            connectionSelection[capabilityName] ?? binding?.connectionId ?? ""
          const availableTools = toolOptions.filter(
            (option) => option.connectionId === selectedConnectionId,
          )
          const currentTool = binding?.toolId
            ? {
                connectionId: binding.connectionId ?? selectedConnectionId,
                connectionName:
                  binding.connectionName ?? "Unavailable connection",
                connectionAlias: binding.connectionAlias ?? "unavailable",
                connectionStatus: binding.connectionStatus ?? "disabled",
                toolId: binding.toolId,
                toolName: binding.toolName ?? "Unavailable tool",
                toolTitle: binding.toolTitle ?? null,
                toolDescription: binding.toolDescription ?? "",
                inputSchema: {},
                toolFingerprint: binding.currentToolFingerprint ?? "",
              }
            : null
          const toolChoices =
            currentTool &&
            !availableTools.some((tool) => tool.toolId === currentTool.toolId)
              ? [currentTool, ...availableTools]
              : availableTools

          return (
            <div
              key={capabilityName}
              className="rounded-md border border-slate-200 bg-slate-50 p-2"
            >
              <div className="mb-2 flex items-center justify-between gap-2">
                <span className="truncate text-xs font-medium text-slate-700">
                  {capabilityName}
                </span>
                <button
                  type="button"
                  aria-label={`Remove ${capabilityName}`}
                  onClick={() => {
                    onChange(
                      nodeCapabilities.filter(
                        (item) => item !== capabilityName,
                      ),
                    )
                    if (capability) onUnbind(capability.id)
                  }}
                  className="text-slate-300 hover:text-danger-500"
                >
                  <Trash2 className="h-3.5 w-3.5" />
                </button>
              </div>
              {!capability ? (
                <p className="text-xs text-amber-600">
                  Capability is not registered. Remove it or select a registered
                  capability.
                </p>
              ) : capability.providerType !== "MCP" ? (
                <p className="text-xs text-slate-500">
                  {capability.providerType} capability — managed by OpenState;
                  no external MCP binding required.
                </p>
              ) : (
                <>
                  <select
                    aria-label={`${capabilityName} MCP connection`}
                    value={selectedConnectionId}
                    onChange={(event) =>
                      setConnectionSelection((current) => ({
                        ...current,
                        [capabilityName]: event.target.value,
                      }))
                    }
                    disabled={!projectId || bindingPending}
                    className="mb-1.5 w-full rounded-md border border-slate-200 bg-white px-2 py-1.5 text-xs text-slate-700"
                  >
                    <option value="">Select MCP connection</option>
                    {connectionChoices.map((option) => (
                      <option
                        key={option.connectionId}
                        value={option.connectionId}
                      >
                        {option.connectionAlias} · {option.connectionStatus}
                      </option>
                    ))}
                  </select>
                  <select
                    aria-label={`${capabilityName} MCP tool`}
                    value={binding?.toolId ?? ""}
                    onChange={(event) => {
                      const option = toolOptions.find(
                        (item) => item.toolId === event.target.value,
                      )
                      if (option)
                        onBind(
                          capability.id,
                          option.connectionId,
                          option.toolId,
                        )
                    }}
                    disabled={
                      !projectId || !selectedConnectionId || bindingPending
                    }
                    className="w-full rounded-md border border-slate-200 bg-white px-2 py-1.5 text-xs text-slate-700"
                  >
                    <option value="">Select discovered tool</option>
                    {toolChoices.map((option) => (
                      <option key={option.toolId} value={option.toolId}>
                        {option.toolName}
                        {option === currentTool ? " · current" : ""}
                      </option>
                    ))}
                  </select>
                  {binding ? (
                    <p
                      className={`mt-1.5 text-[11px] font-medium ${healthClass(binding.health)}`}
                    >
                      {healthLabel(binding.health)}
                      {binding.healthReason ? ` — ${binding.healthReason}` : ""}
                    </p>
                  ) : (
                    <p className="mt-1.5 text-[11px] font-medium text-amber-600">
                      MISSING MAPPING — select a discovered tool
                    </p>
                  )}
                </>
              )}
            </div>
          )
        })
      )}
    </div>
  )
}

function NodePanel({
  nodeId,
  projectId,
  capabilities,
  toolOptions,
  bindings,
  capabilitiesLoading,
  catalogLoading,
  catalogError,
  bindingPending,
  onBind,
  onUnbind,
}: NodePanelProps) {
  const node = useStateBuilderStore((s) =>
    s.workflow.nodes.find((n) => n.id === nodeId),
  )
  const updateNode = useStateBuilderStore((s) => s.updateNode)
  const removeNode = useStateBuilderStore((s) => s.removeNode)
  const workflow = useStateBuilderStore((s) => s.workflow)
  const addTransition = useStateBuilderStore((s) => s.addTransition)
  const updateTransition = useStateBuilderStore((s) => s.updateTransition)
  const removeTransition = useStateBuilderStore((s) => s.removeTransition)

  if (!node) return null

  const outTransitions = workflow.transitions.filter(
    (t) => t.sourceStateId === node.id,
  )
  const isTerminal = node.isTerminal || node.kind === "END"
  const editable = node.kind === "STATE" || node.kind === "DECISION"
  const nodeMeta = nodeTypeList.find((n) => n.kind === node.kind)
  const NodeIcon = nodeMeta?.icon ?? Boxes
  const nodeColor = nodeMeta?.color ?? "#1f8fca"

  const handleAddTransition = () => {
    const end = workflow.nodes.find((n) => n.kind === "END")
    const target = end ?? workflow.nodes.find((n) => n.id !== node.id)
    if (target) addTransition(node.id, target.id)
  }

  return (
    <div className="flex flex-col overflow-y-auto">
      <div className="flex items-center justify-between gap-2 px-3 py-2">
        <div className="flex min-w-0 items-center gap-2.5">
          <span
            className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg text-white shadow-sm"
            style={{ backgroundColor: nodeColor }}
          >
            <NodeIcon className="h-5 w-5" />
          </span>
          <div className="min-w-0">
            <h2 className="truncate text-sm font-semibold text-slate-800">
              {node.name}
            </h2>
            <p className="text-xs text-slate-400">{node.kind} / Node</p>
          </div>
        </div>
        <button
          type="button"
          onClick={() => {
            if (
              window.confirm(
                `Hapus node "${node.name}"? Semua transisi yang terhubung juga akan dihapus.`,
              )
            ) {
              removeNode(node.id)
            }
          }}
          className="rounded-md border border-danger-200 p-1.5 text-danger-500 hover:bg-danger-50"
          title="Hapus node"
        >
          <Trash2 className="h-4 w-4" />
        </button>
      </div>

      <Section title="General">
        <Field label="Name">
          <TextInput
            data-testid="workflow-node-name"
            value={node.name}
            onChange={(e) => updateNode(node.id, { name: e.target.value })}
          />
        </Field>
        <Field label="Description">
          <TextArea
            rows={2}
            value={node.description ?? ""}
            onChange={(e) =>
              updateNode(node.id, { description: e.target.value })
            }
          />
        </Field>
        {node.kind !== "START" && node.kind !== "END" ? (
          <label className="flex items-center gap-2 text-sm text-slate-700">
            <input
              type="checkbox"
              checked={!!node.isTerminal}
              onChange={(e) =>
                updateNode(node.id, { isTerminal: e.target.checked })
              }
              className="h-4 w-4 accent-primary-500"
            />
            Terminal state
          </label>
        ) : null}
      </Section>

      {editable ? (
        <>
          <Section title="Context (required)">
            <TagInput
              values={node.requiredContext}
              placeholder="cth: payment.status"
              onChange={(v) => updateNode(node.id, { requiredContext: v })}
            />
          </Section>

          <Section title="Capabilities (allowed)">
            <MCPBindingsEditor
              projectId={projectId}
              nodeCapabilities={node.capabilities}
              capabilities={capabilities}
              toolOptions={toolOptions}
              bindings={bindings}
              capabilitiesLoading={capabilitiesLoading}
              catalogLoading={catalogLoading}
              catalogError={catalogError}
              bindingPending={bindingPending}
              onChange={(values) =>
                updateNode(node.id, { capabilities: values })
              }
              onBind={onBind}
              onUnbind={onUnbind}
            />
          </Section>

          <Section title="Instructions (LLM)">
            <TextArea
              rows={5}
              value={node.instructions ?? ""}
              onChange={(e) =>
                updateNode(node.id, { instructions: e.target.value })
              }
            />
          </Section>

          <Section title="Timeout">
            <Field label="Timeout (detik)">
              <TextInput
                type="number"
                value={node.policy.timeoutSeconds ?? ""}
                placeholder="cth: 3600"
                onChange={(e) =>
                  updateNode(node.id, {
                    policy: {
                      ...node.policy,
                      timeoutSeconds: e.target.value
                        ? Number(e.target.value)
                        : undefined,
                    },
                  })
                }
              />
            </Field>
            <label className="flex items-center gap-2 text-sm text-slate-700">
              <input
                type="checkbox"
                checked={!!node.policy.onTimeout}
                onChange={(e) =>
                  updateNode(node.id, {
                    policy: {
                      ...node.policy,
                      onTimeout: e.target.checked ? "state.timeout" : undefined,
                    },
                  })
                }
                className="h-4 w-4 accent-primary-500"
              />
              Tambahkan transisi saat timeout
            </label>
            {node.policy.onTimeout ? (
              <p className="text-xs text-slate-500">
                on timeout → event <b>{node.policy.onTimeout}</b>
              </p>
            ) : null}
          </Section>

          <Section title="Retry">
            <Field label="Max attempts">
              <TextInput
                type="number"
                value={node.policy.retry?.maxAttempts ?? ""}
                placeholder="cth: 3"
                onChange={(e) =>
                  updateNode(node.id, {
                    policy: {
                      ...node.policy,
                      retry: {
                        maxAttempts: e.target.value
                          ? Number(e.target.value)
                          : 0,
                        backoffMs: node.policy.retry?.backoffMs ?? 1000,
                        retryableEvents:
                          node.policy.retry?.retryableEvents ?? [],
                      },
                    },
                  })
                }
              />
            </Field>
          </Section>
        </>
      ) : null}

      {!isTerminal ? (
        <Section title="Transitions (keluar)">
          <AddButton label="Transisi" onClick={handleAddTransition} />
          {outTransitions.length === 0 ? (
            <p className="text-xs text-slate-400">
              Belum ada transisi keluar. Hubungkan ke node lain di canvas, atau
              tambah di sini.
            </p>
          ) : (
            outTransitions.map((t) => {
              const target = workflow.nodes.find(
                (n) => n.id === t.targetStateId,
              )
              return (
                <div
                  key={t.id}
                  className="rounded-md border border-slate-200 bg-slate-50 px-2 py-1.5"
                >
                  <div className="flex items-center justify-between">
                    <span className="text-sm font-medium text-slate-700">
                      {t.event || "event"}{" "}
                      <span className="text-slate-400">→</span>{" "}
                      {target?.name ?? "?"}
                    </span>
                    <button
                      type="button"
                      onClick={() => {
                        if (
                          window.confirm(
                            `Hapus transisi "${t.event || "event"}"?`,
                          )
                        ) {
                          removeTransition(t.id)
                        }
                      }}
                      className="text-slate-300 hover:text-danger-500"
                    >
                      <Trash2 className="h-3.5 w-3.5" />
                    </button>
                  </div>
                  <div className="mt-1.5 flex items-center gap-1.5">
                    <TextInput
                      value={t.event}
                      placeholder="event"
                      onChange={(e) =>
                        updateTransition(t.id, { event: e.target.value })
                      }
                      className="!py-1 !text-xs"
                    />
                    <TextInput
                      type="number"
                      value={t.priority}
                      title="priority: lebih kecil dievaluasi lebih dulu"
                      onChange={(e) =>
                        updateTransition(t.id, {
                          priority: Number(e.target.value),
                        })
                      }
                      className="!w-16 !py-1 !text-xs"
                    />
                  </div>
                </div>
              )
            })
          )}
        </Section>
      ) : null}

      {!isTerminal ? (
        <Section title="Guards">
          <GuardsEditor
            guards={node.guardGroups ?? []}
            onChange={(guards) => updateNode(node.id, { guardGroups: guards })}
          />
        </Section>
      ) : null}
    </div>
  )
}

/* ------------------------------------------------------------------ */
/* Panel untuk EDGE (transition)                                       */
/* ------------------------------------------------------------------ */

function EdgePanel({ edgeId }: { edgeId: string }) {
  const workflow = useStateBuilderStore((s) => s.workflow)
  const updateTransition = useStateBuilderStore((s) => s.updateTransition)
  const removeTransition = useStateBuilderStore((s) => s.removeTransition)

  const transition = workflow.transitions.find((t) => t.id === edgeId)
  const source = workflow.nodes.find((n) => n.id === transition?.sourceStateId)
  const target = workflow.nodes.find((n) => n.id === transition?.targetStateId)

  if (!transition) return null

  return (
    <div className="flex flex-col overflow-y-auto">
      <div className="flex items-center justify-between gap-2 px-3 py-2">
        <div className="flex min-w-0 items-center gap-2.5">
          <span className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-slate-700 text-white shadow-sm">
            <GitBranch className="h-5 w-5" />
          </span>
          <div className="min-w-0">
            <h2 className="text-sm font-semibold text-slate-800">Transition</h2>
            <p className="truncate text-xs text-slate-400">
              {source?.name ?? "?"} <span className="text-slate-300">→</span>{" "}
              {target?.name ?? "?"}
            </p>
          </div>
        </div>
        <button
          type="button"
          onClick={() => {
            if (
              window.confirm(`Hapus transisi "${transition.event || "event"}"?`)
            ) {
              removeTransition(transition.id)
            }
          }}
          className="rounded-md border border-danger-200 p-1.5 text-danger-500 hover:bg-danger-50"
        >
          <Trash2 className="h-4 w-4" />
        </button>
      </div>

      <Section title="Event">
        <Field label="Event (pemicu)">
          <TextInput
            value={transition.event}
            placeholder="cth: payment.success"
            onChange={(e) =>
              updateTransition(transition.id, { event: e.target.value })
            }
          />
        </Field>
        <Field label="Priority (kecil = dulu)">
          <TextInput
            type="number"
            value={transition.priority}
            onChange={(e) =>
              updateTransition(transition.id, {
                priority: Number(e.target.value),
              })
            }
          />
        </Field>
      </Section>

      <Section title="Guards">
        <GuardsEditor
          guards={transition.guards}
          onChange={(guards) => updateTransition(transition.id, { guards })}
        />
      </Section>
    </div>
  )
}

/* ------------------------------------------------------------------ */
/* Panel utama                                                        */
/* ------------------------------------------------------------------ */

export function PropertiesPanel({
  selectedNodeId,
  selectedEdgeId,
  projectId,
}: {
  selectedNodeId: string | null
  selectedEdgeId: string | null
  projectId?: string
}) {
  const capabilitiesQuery = useCapabilitiesList({
    status: "ACTIVE",
  })
  const optionsQuery = useListProjectMCPToolOptions(projectId)
  const bindingsQuery = useListProjectMCPBindings(projectId)
  const upsertBinding = useUpsertProjectMCPBinding()
  const deleteBinding = useDeleteProjectMCPBinding()

  if (selectedNodeId) {
    return (
      <NodePanel
        nodeId={selectedNodeId}
        projectId={projectId}
        capabilities={capabilitiesQuery.data ?? []}
        toolOptions={optionsQuery.data ?? []}
        bindings={bindingsQuery.data ?? []}
        capabilitiesLoading={capabilitiesQuery.isLoading}
        catalogLoading={optionsQuery.isLoading}
        catalogError={optionsQuery.isError}
        bindingPending={upsertBinding.isPending || deleteBinding.isPending}
        onBind={(capabilityId, connectionId, toolId) => {
          if (!projectId) return
          void upsertBinding.mutateAsync({
            projectId,
            capabilityId,
            connectionId,
            toolId,
          })
        }}
        onUnbind={(capabilityId) => {
          if (!projectId) return
          void deleteBinding.mutateAsync({ projectId, capabilityId })
        }}
      />
    )
  }
  if (selectedEdgeId) {
    return <EdgePanel edgeId={selectedEdgeId} />
  }
  return (
    <div className="flex h-full flex-col items-center justify-center gap-3 p-8 text-center">
      <span className="flex h-14 w-14 items-center justify-center rounded-2xl border border-slate-200 bg-slate-50 text-slate-300 shadow-sm">
        <MousePointerClick className="h-7 w-7" />
      </span>
      <div>
        <p className="text-sm font-semibold text-slate-600">
          Belum ada yang dipilih
        </p>
        <p className="mt-1 text-xs leading-relaxed text-slate-400">
          Pilih sebuah <b className="text-slate-500">state</b> atau{" "}
          <b className="text-slate-500">transisi</b> di canvas untuk mengatur
          propertinya.
        </p>
      </div>
      <div className="mt-1 flex items-center gap-1.5 text-[10px] uppercase tracking-wide text-slate-300">
        <Boxes className="h-3.5 w-3.5" />
        Properties Panel
      </div>
    </div>
  )
}
