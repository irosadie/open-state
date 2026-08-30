"use client"

import type { SimulationResultResponse } from "@openstate/types"
import type { WorkflowVersionResponse } from "@openstate/types"
import type { Edge, Node } from "@xyflow/react"
import { create } from "zustand"
import type {
  TransitionDefinition,
  WorkflowDefinition,
  WorkflowNode,
  WorkflowValidationResult,
} from "./types/workflow"
import {
  createTransition,
  fromFlowNodes,
  toFlowEdges,
  toFlowNodes,
  uid,
  validateWorkflow,
} from "./types/workflow.utils"
import { getLayoutedNodes } from "./utils/auto-layout"
import {
  clearDraftLocal,
  loadApiId,
  loadApiVersion,
  loadDraftLocal,
  saveApiId,
  saveApiVersion,
} from "./utils/draft-bridge"
import { toast } from "./utils/toast"
import {
  createWorkflowApi,
  getWorkflowApi,
  publishWorkflowApi,
  updateWorkflowApi,
} from "./utils/workflow-api"
import { padelBookingWorkflow } from "./workflows"
import { ALL_WORKFLOWS } from "./workflows/intent-resolver"

/** Maksimal histori undo yang disimpan */
const MAX_HISTORY = 50

/** Daftar workflow contoh yang bisa di-load (single source: intent-resolver) */
export const EXAMPLE_WORKFLOWS: WorkflowDefinition[] = ALL_WORKFLOWS

export interface SimulationDraftEvent {
  id: string
  type: string
  payloadText: string
}

export interface SimulationFocusTarget {
  nodeIds: string[]
  transitionId: string | null
}

interface StateBuilderState {
  workflow: WorkflowDefinition
  nodes: Node[]
  edges: Edge[]
  selectedNodeId: string | null
  selectedEdgeId: string | null
  validation: WorkflowValidationResult | null

  // undo/redo
  history: WorkflowDefinition[]
  future: WorkflowDefinition[]

  // persistence & ui status
  isHydrated: boolean
  isSaving: boolean
  lastSavedAt: string | null
  saveError: string | null
  saveConflict: boolean
  searchQuery: string

  /** Authoritative API workflow id (null until first successful API save) */
  apiWorkflowId: string | null
  /** Optimistic version tracked from the API response */
  apiVersion: number

  // transient simulation state (never persisted with the workflow draft)
  simulationOpen: boolean
  simulationInitialContextText: string
  simulationEvents: SimulationDraftEvent[]
  simulationResult: SimulationResultResponse | null
  simulationError: string | null
  simulationIsRunning: boolean
  simulationSelectedSequence: number | null
  simulationFocusTarget: SimulationFocusTarget | null
  simulationStale: boolean
  simulationFingerprint: string | null

  setSaving: (v: boolean) => void
  setSaveError: (message: string | null, conflict?: boolean) => void

  // actions
  hydrate: (workflowId?: string) => Promise<void>
  setNodes: (nodes: Node[]) => void
  setEdges: (edges: Edge[]) => void
  selectNode: (id: string | null) => void
  selectEdge: (id: string | null) => void
  addNode: (kind: WorkflowNode["kind"]) => void
  addNodeAt: (
    kind: WorkflowNode["kind"],
    position: { x: number; y: number },
  ) => void
  updateNode: (id: string, patch: Partial<WorkflowNode>) => void
  removeNode: (id: string) => void
  addTransition: (source: string, target: string) => void
  updateTransition: (id: string, patch: Partial<TransitionDefinition>) => void
  removeTransition: (id: string) => void
  setWorkflowMeta: (patch: Partial<WorkflowDefinition>) => void
  undo: () => void
  redo: () => void
  loadWorkflow: (wf: WorkflowDefinition) => void
  newWorkflow: () => void
  resetToWorkflow: (slug: string) => void
  clearAll: () => void
  setSearchQuery: (q: string) => void
  persist: () => Promise<void>
  publish: () => Promise<WorkflowVersionResponse>

  getSimulationSnapshot: () => WorkflowDefinition
  openSimulation: () => void
  closeSimulation: () => void
  setSimulationInitialContextText: (value: string) => void
  addSimulationEvent: () => void
  updateSimulationEvent: (
    id: string,
    patch: Partial<Pick<SimulationDraftEvent, "type" | "payloadText">>,
  ) => void
  removeSimulationEvent: (id: string) => void
  setSimulationIsRunning: (value: boolean) => void
  setSimulationError: (value: string | null) => void
  setSimulationResult: (
    result: SimulationResultResponse,
    fingerprint: string,
  ) => void
  selectSimulationStep: (sequence: number | null) => void
  resetSimulation: () => void
  markSimulationStale: () => void

  resetValidation: () => void
  revalidate: () => void
}

/** Bangun snapshot workflow dari state saat ini (tanpa side effect) */
function buildSnapshot(
  wf: WorkflowDefinition,
  nodes: Node[],
  edges: Edge[],
): WorkflowDefinition {
  return {
    ...wf,
    nodes: fromFlowNodes(nodes),
    transitions: edges.map((e) => {
      const existing = wf.transitions.find((t) => t.id === e.id)
      return {
        id: e.id,
        event:
          (e.data as { transition?: { event: string } } | undefined)?.transition
            ?.event ??
          existing?.event ??
          "event",
        sourceStateId: e.source,
        targetStateId: e.target,
        guards: existing?.guards ?? [],
        priority: existing?.priority ?? 10,
      }
    }),
  }
}

/** Terapkan workflow ke nodes/edges + layout */
function materialize(
  wf: WorkflowDefinition,
  layout: boolean,
): { nodes: Node[]; edges: Edge[] } {
  const edges = toFlowEdges(wf)
  const rawNodes = toFlowNodes(wf)
  return {
    edges,
    nodes: layout ? getLayoutedNodes(rawNodes, edges, "TB") : rawNodes,
  }
}

function parseServerDraft(
  value: Record<string, unknown>,
): WorkflowDefinition | null {
  if (!Array.isArray(value.nodes) || !Array.isArray(value.transitions)) {
    return null
  }
  return value as unknown as WorkflowDefinition
}

function getErrorMessage(error: unknown): string {
  if (error instanceof Error) return error.message
  if (typeof error === "object" && error !== null) {
    const value = error as {
      message?: unknown
      error?: unknown
      details?: unknown
    }
    if (typeof value.message === "string") return value.message
    if (typeof value.error === "string") {
      if (Array.isArray(value.details)) {
        const detailMessages = value.details
          .map((detail) =>
            typeof detail === "object" &&
            detail !== null &&
            "message" in detail &&
            typeof (detail as { message?: unknown }).message === "string"
              ? (detail as { message: string }).message
              : null,
          )
          .filter((message): message is string => Boolean(message))
        if (detailMessages.length > 0) {
          return `${value.error}: ${detailMessages.join("; ")}`
        }
      }
      return value.error
    }
  }
  return "Permintaan gagal"
}

function focusTargetFor(
  result: SimulationResultResponse | null,
  sequence: number | null,
): SimulationFocusTarget | null {
  if (!result || sequence === null) return null
  const step = result.steps.find((item) => item.sequence === sequence)
  if (!step) return null
  return {
    nodeIds: [step.stateBefore.id, step.stateAfter?.id].filter(
      (id): id is string => Boolean(id),
    ),
    transitionId: step.selectedTransitionId || null,
  }
}

function invalidateSimulation(
  state: StateBuilderState,
): Partial<StateBuilderState> {
  return state.simulationResult
    ? {
        simulationStale: true,
        simulationSelectedSequence: null,
        simulationFocusTarget: null,
      }
    : {}
}

function invalidateSimulationFor(
  state: StateBuilderState,
  nextWorkflow: WorkflowDefinition,
): Partial<StateBuilderState> {
  return JSON.stringify(state.workflow) === JSON.stringify(nextWorkflow)
    ? {}
    : invalidateSimulation(state)
}

/** Wrapper action yang otomatis: push history + revalidate */
function tracked(
  get: () => StateBuilderState,
  set: (partial: Partial<StateBuilderState>) => void,
  mutate: () => WorkflowDefinition,
) {
  const prev = get()
  const nextWf = mutate()
  const { nodes, edges } = materialize(nextWf, false)
  const history = [...prev.history, prev.workflow].slice(-MAX_HISTORY)
  set({
    workflow: nextWf,
    nodes,
    edges,
    history,
    future: [],
    selectedNodeId: prev.selectedNodeId,
    selectedEdgeId: prev.selectedEdgeId,
    validation: validateWorkflow(nextWf),
    ...invalidateSimulationFor(prev, nextWf),
  })
  void schedulePersist()
}

let persistTimer: ReturnType<typeof setTimeout> | null = null
let persistQueue: Promise<void> = Promise.resolve()
let legacyImportPending = false

/** Auto-save (debounce 800ms): persist the complete graph to the API. */
function schedulePersist() {
  if (persistTimer) clearTimeout(persistTimer)
  persistTimer = setTimeout(() => {
    const state = useStateBuilderStore.getState()
    state.setSaving(true)
    void state
      .persist()
      .then(() => {
        useStateBuilderStore.getState().setSaving(false)
      })
      .catch((err) => {
        toast.error(`Gagal menyimpan draft: ${getErrorMessage(err)}`)
        useStateBuilderStore.getState().setSaving(false)
      })
  }, 800)
}

export const useStateBuilderStore = create<StateBuilderState>()((set, get) => {
  const initialWf = structuredClone(padelBookingWorkflow)
  const { nodes, edges } = materialize(initialWf, true)

  return {
    workflow: initialWf,
    nodes,
    edges,
    selectedNodeId: null,
    selectedEdgeId: null,
    validation: validateWorkflow(initialWf),

    history: [],
    future: [],

    isHydrated: false,
    isSaving: false,
    lastSavedAt: null,
    saveError: null,
    saveConflict: false,
    searchQuery: "",
    apiWorkflowId: null,
    apiVersion: 0,

    simulationOpen: false,
    simulationInitialContextText: "{}",
    simulationEvents: [],
    simulationResult: null,
    simulationError: null,
    simulationIsRunning: false,
    simulationSelectedSequence: null,
    simulationFocusTarget: null,
    simulationStale: false,
    simulationFingerprint: null,

    hydrate: async (workflowId) => {
      try {
        if (workflowId) {
          const serverWorkflow = await getWorkflowApi({ id: workflowId })
          const draft = parseServerDraft(serverWorkflow.definition)
          if (!draft) {
            throw new Error("Draft server tidak memiliki graph yang valid")
          }
          const { nodes: n, edges: e } = materialize(draft, true)
          set({
            workflow: draft,
            nodes: n,
            edges: e,
            validation: validateWorkflow(draft),
            apiWorkflowId: serverWorkflow.id,
            apiVersion: serverWorkflow.version,
          })
          saveApiId(serverWorkflow.id)
          saveApiVersion(serverWorkflow.version)
        } else {
          // A legacy local draft is only considered when no server workflow id
          // was requested; it can never overwrite a server draft.
          const apiId = loadApiId()
          const apiVer = loadApiVersion()
          if (apiId) {
            const serverWorkflow = await getWorkflowApi({ id: apiId })
            const draft = parseServerDraft(serverWorkflow.definition)
            if (!draft)
              throw new Error("Draft server tidak memiliki graph yang valid")
            const { nodes: n, edges: e } = materialize(draft, true)
            set({
              workflow: draft,
              nodes: n,
              edges: e,
              validation: validateWorkflow(draft),
              apiWorkflowId: serverWorkflow.id,
              apiVersion: serverWorkflow.version,
            })
          } else {
            const localDraft = loadDraftLocal()
            const draft =
              localDraft &&
              typeof window !== "undefined" &&
              window.confirm(
                "Ditemukan draft lama di browser. Import ke server sekarang?",
              )
                ? localDraft
                : null
            if (!draft) {
              set({ apiWorkflowId: null, apiVersion: 0 })
            } else {
              legacyImportPending = true
              const { nodes: n, edges: e } = materialize(draft, true)
              set({
                workflow: draft,
                nodes: n,
                edges: e,
                validation: validateWorkflow(draft),
                apiWorkflowId: null,
                apiVersion: apiVer ?? 0,
              })
            }
          }
        }
      } catch (err) {
        toast.error(`Gagal memuat draft: ${getErrorMessage(err)}`)
      } finally {
        set((state) => ({ isHydrated: true, ...invalidateSimulation(state) }))
      }
    },

    setNodes: (nodes) => {
      const state = get()
      const wf = buildSnapshot(state.workflow, nodes, state.edges)
      const history = [...state.history, state.workflow].slice(-MAX_HISTORY)
      set({
        nodes,
        workflow: wf,
        history,
        future: [],
        ...invalidateSimulationFor(state, wf),
      })
      get().revalidate()
      void schedulePersist()
    },

    setEdges: (edges) => {
      const state = get()
      const wf = buildSnapshot(state.workflow, state.nodes, edges)
      const history = [...state.history, state.workflow].slice(-MAX_HISTORY)
      set({
        edges: toFlowEdges(wf),
        workflow: wf,
        history,
        future: [],
        ...invalidateSimulationFor(state, wf),
      })
      get().revalidate()
      void schedulePersist()
    },

    selectNode: (id) => set({ selectedNodeId: id, selectedEdgeId: null }),
    selectEdge: (id) => set({ selectedEdgeId: id, selectedNodeId: null }),

    addNode: (kind) => {
      tracked(get, set, () => {
        const wf = get().workflow
        const node: WorkflowNode = {
          id: uid(),
          kind,
          name: kind,
          description: "",
          requiredContext: [],
          capabilities: [],
          instructions: "",
          policy: {},
          isTerminal: kind === "END",
          position: {
            x: 120 + wf.nodes.length * 40,
            y: 120 + wf.nodes.length * 40,
          },
        }
        if (kind === "START" && !wf.entryNodeId) node.position = { x: 0, y: 0 }
        const newWf: WorkflowDefinition = {
          ...wf,
          entryNodeId:
            kind === "START" && !wf.entryNodeId ? node.id : wf.entryNodeId,
          nodes: [...wf.nodes, node],
        }
        set({ selectedNodeId: node.id })
        return newWf
      })
    },

    addNodeAt: (kind, position) => {
      tracked(get, set, () => {
        const wf = get().workflow
        const node: WorkflowNode = {
          id: uid(),
          kind,
          name: kind,
          description: "",
          requiredContext: [],
          capabilities: [],
          instructions: "",
          policy: {},
          isTerminal: kind === "END",
          position,
        }
        const newWf: WorkflowDefinition = {
          ...wf,
          entryNodeId:
            kind === "START" && !wf.entryNodeId ? node.id : wf.entryNodeId,
          nodes: [...wf.nodes, node],
        }
        set({ selectedNodeId: node.id })
        return newWf
      })
    },

    updateNode: (id, patch) => {
      tracked(get, set, () => {
        const wf = get().workflow
        const nodes = wf.nodes.map((n) =>
          n.id === id ? { ...n, ...patch } : n,
        )
        return { ...wf, nodes }
      })
    },

    removeNode: (id) => {
      tracked(get, set, () => {
        const wf = get().workflow
        const nodes = wf.nodes.filter((n) => n.id !== id)
        const transitions = wf.transitions.filter(
          (t) => t.sourceStateId !== id && t.targetStateId !== id,
        )
        set({ selectedNodeId: null, selectedEdgeId: null })
        return {
          ...wf,
          entryNodeId: wf.entryNodeId === id ? undefined : wf.entryNodeId,
          nodes,
          transitions,
        }
      })
    },

    addTransition: (source, target) => {
      if (!source || !target || source === target) return
      const wf = get().workflow
      const existing = wf.transitions.some(
        (t) => t.sourceStateId === source && t.targetStateId === target,
      )
      if (existing) return
      tracked(get, set, () => {
        const tr = createTransition(source, "event", target)
        const next: WorkflowDefinition = {
          ...wf,
          transitions: [...wf.transitions, tr],
        }
        set({ selectedEdgeId: tr.id })
        return next
      })
    },

    updateTransition: (id, patch) => {
      tracked(get, set, () => {
        const wf = get().workflow
        const transitions = wf.transitions.map((t) =>
          t.id === id ? { ...t, ...patch } : t,
        )
        return { ...wf, transitions }
      })
    },

    removeTransition: (id) => {
      tracked(get, set, () => {
        const wf = get().workflow
        const transitions = wf.transitions.filter((t) => t.id !== id)
        set({ selectedEdgeId: null })
        return { ...wf, transitions }
      })
    },

    setWorkflowMeta: (patch) => {
      const state = get()
      set({
        workflow: { ...state.workflow, ...patch },
        ...invalidateSimulationFor(state, {
          ...state.workflow,
          ...patch,
        }),
      })
      void schedulePersist()
    },

    undo: () => {
      const { history, future, workflow } = get()
      const prev = history.at(-1)
      if (!prev) return
      const { nodes, edges } = materialize(prev, false)
      set({
        workflow: prev,
        nodes,
        edges,
        history: history.slice(0, -1),
        future: [workflow, ...future].slice(0, MAX_HISTORY),
        selectedNodeId: null,
        selectedEdgeId: null,
        validation: validateWorkflow(prev),
        ...invalidateSimulation(get()),
      })
      void schedulePersist()
    },

    redo: () => {
      const { future, history, workflow } = get()
      const next = future[0]
      if (!next) return
      const { nodes, edges } = materialize(next, false)
      set({
        workflow: next,
        nodes,
        edges,
        future: future.slice(1),
        history: [...history, workflow].slice(-MAX_HISTORY),
        selectedNodeId: null,
        selectedEdgeId: null,
        validation: validateWorkflow(next),
        ...invalidateSimulation(get()),
      })
      void schedulePersist()
    },

    loadWorkflow: (wf) => {
      const { nodes, edges } = materialize(wf, true)
      set({
        workflow: wf,
        nodes,
        edges,
        history: [],
        future: [],
        selectedNodeId: null,
        selectedEdgeId: null,
        validation: validateWorkflow(wf),
        // Clear API id when loading a new workflow — needs a fresh create
        apiWorkflowId: null,
        apiVersion: 0,
        ...invalidateSimulation(get()),
      })
      saveApiId(null)
      void schedulePersist()
    },

    newWorkflow: () => {
      const fresh: WorkflowDefinition = {
        slug: `draft-${Date.now().toString(36)}`,
        name: "Untitled Workflow",
        description: "",
        schemaVersion: 1,
        status: "DRAFT",
        nodes: [],
        transitions: [],
        policy: {
          interruptible: "USER_REQUESTED",
          priority: 10,
        },
        triggers: [],
      }
      const { nodes, edges } = materialize(fresh, false)
      set({
        workflow: fresh,
        nodes,
        edges,
        history: [],
        future: [],
        selectedNodeId: null,
        selectedEdgeId: null,
        validation: validateWorkflow(fresh),
        apiWorkflowId: null,
        apiVersion: 0,
        ...invalidateSimulation(get()),
      })
      saveApiId(null)
      void schedulePersist()
    },

    resetToWorkflow: (slug) => {
      const example = EXAMPLE_WORKFLOWS.find((w) => w.slug === slug)
      get().loadWorkflow(structuredClone(example ?? padelBookingWorkflow))
    },

    clearAll: () => {
      get().newWorkflow()
    },

    setSearchQuery: (q) => set({ searchQuery: q }),

    setSaving: (v) => set({ isSaving: v }),
    setSaveError: (message, conflict = false) =>
      set({ saveError: message, saveConflict: conflict }),

    persist: () => {
      // Serialize all saves so autosave, manual save, and publish cannot race
      // with the same optimistic-lock version.
      const operation = async () => {
        const { workflow, nodes, edges, apiWorkflowId, apiVersion } = get()
        const snapshot = buildSnapshot(workflow, nodes, edges)

        // Sync the workflow root and its complete draft graph to the API.
        set({ isSaving: true, saveError: null, saveConflict: false })
        try {
          if (!apiWorkflowId) {
            // First save — create the workflow root via the Builder API.
            const created = await createWorkflowApi({
              slug: snapshot.slug,
              name: snapshot.name,
              description: snapshot.description,
              definition: snapshot,
            })
            set({ apiWorkflowId: created.id, apiVersion: created.version })
            saveApiId(created.id)
            saveApiVersion(created.version)
            if (legacyImportPending) {
              clearDraftLocal()
              legacyImportPending = false
            }
          } else {
            // Subsequent save — bump the optimistic version.
            const updated = await updateWorkflowApi({
              id: apiWorkflowId,
              version: apiVersion,
              name: snapshot.name,
              description: snapshot.description,
              definition: snapshot,
            })
            set({ apiVersion: updated.version })
            saveApiVersion(updated.version)
          }

          set({ isSaving: false, lastSavedAt: new Date().toISOString() })
        } catch (err) {
          const apiError = err as {
            status?: number
            response?: { status?: number }
          }
          const conflict =
            apiError.response?.status === 409 || apiError.status === 409
          const message = getErrorMessage(err)
          set({ isSaving: false, saveError: message, saveConflict: conflict })
          throw err
        }
      }

      const next = persistQueue.then(operation, operation)
      persistQueue = next.catch(() => undefined)
      return next
    },

    publish: async () => {
      // Flush a pending debounce and ensure the server draft is current before
      // creating an immutable version.
      if (persistTimer) {
        clearTimeout(persistTimer)
        persistTimer = null
      }
      await get().persist()
      const { workflow, nodes, edges, apiWorkflowId, apiVersion } = get()
      const snapshot = buildSnapshot(workflow, nodes, edges)
      const validation = validateWorkflow(snapshot)
      set({ validation })
      if (!validation.valid) {
        throw new Error("Workflow tidak valid. Perbaiki error sebelum publish.")
      }

      // Ensure there's an API workflow root before publishing.
      let id = apiWorkflowId
      let ver = apiVersion
      if (!id) {
        const created = await createWorkflowApi({
          slug: snapshot.slug,
          name: snapshot.name,
          description: snapshot.description,
          definition: snapshot,
        })
        id = created.id
        ver = created.version
        set({ apiWorkflowId: id, apiVersion: ver })
        saveApiId(id)
        saveApiVersion(ver)
      }

      const published = await publishWorkflowApi({
        id,
        version: ver,
      })
      // After publish, update the local version tracking.
      set({ apiVersion: ver + 1, lastSavedAt: new Date().toISOString() })
      saveApiVersion(ver + 1)
      return published
    },

    getSimulationSnapshot: () => {
      const state = get()
      return buildSnapshot(state.workflow, state.nodes, state.edges)
    },

    openSimulation: () =>
      set({
        simulationOpen: true,
        simulationError: null,
        simulationSelectedSequence: null,
        simulationFocusTarget: null,
      }),

    closeSimulation: () => set({ simulationOpen: false }),

    setSimulationInitialContextText: (value) =>
      set({ simulationInitialContextText: value, simulationError: null }),

    addSimulationEvent: () =>
      set((state) => ({
        simulationEvents: [
          ...state.simulationEvents,
          { id: uid("sim-event"), type: "", payloadText: "{}" },
        ],
        simulationError: null,
      })),

    updateSimulationEvent: (id, patch) =>
      set((state) => ({
        simulationEvents: state.simulationEvents.map((event) =>
          event.id === id ? { ...event, ...patch } : event,
        ),
        simulationError: null,
      })),

    removeSimulationEvent: (id) =>
      set((state) => ({
        simulationEvents: state.simulationEvents.filter(
          (event) => event.id !== id,
        ),
        simulationError: null,
      })),

    setSimulationIsRunning: (value) => set({ simulationIsRunning: value }),

    setSimulationError: (value) =>
      set({ simulationError: value, simulationIsRunning: false }),

    setSimulationResult: (result, fingerprint) =>
      set({
        simulationResult: result,
        simulationFingerprint: fingerprint,
        simulationError: null,
        simulationIsRunning: false,
        simulationStale: false,
        simulationSelectedSequence: result.steps[0]?.sequence ?? null,
        simulationFocusTarget: focusTargetFor(
          result,
          result.steps[0]?.sequence ?? null,
        ),
      }),

    selectSimulationStep: (sequence) =>
      set((state) => ({
        simulationSelectedSequence: sequence,
        simulationFocusTarget: focusTargetFor(state.simulationResult, sequence),
      })),

    resetSimulation: () =>
      set({
        simulationInitialContextText: "{}",
        simulationEvents: [],
        simulationResult: null,
        simulationError: null,
        simulationIsRunning: false,
        simulationSelectedSequence: null,
        simulationFocusTarget: null,
        simulationStale: false,
        simulationFingerprint: null,
      }),

    markSimulationStale: () =>
      set((state) =>
        state.simulationResult
          ? {
              simulationStale: true,
              simulationSelectedSequence: null,
              simulationFocusTarget: null,
            }
          : {},
      ),

    resetValidation: () => set({ validation: null }),

    revalidate: () => {
      set({ validation: validateWorkflow(get().workflow) })
    },
  }
})
