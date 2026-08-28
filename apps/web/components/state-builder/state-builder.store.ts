"use client"

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
  loadApiId,
  loadApiVersion,
  loadDraftLocal,
  saveApiId,
  saveApiVersion,
  saveDraftLocal,
} from "./utils/draft-bridge"
import { toast } from "./utils/toast"
import {
  createWorkflowApi,
  publishWorkflowApi,
  updateWorkflowApi,
} from "./utils/workflow-api"
import { padelBookingWorkflow } from "./workflows"
import { ALL_WORKFLOWS } from "./workflows/intent-resolver"

/** Maksimal histori undo yang disimpan */
const MAX_HISTORY = 50

/** Daftar workflow contoh yang bisa di-load (single source: intent-resolver) */
export const EXAMPLE_WORKFLOWS: WorkflowDefinition[] = ALL_WORKFLOWS

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
  searchQuery: string

  /** Authoritative API workflow id (null until first successful API save) */
  apiWorkflowId: string | null
  /** Optimistic version tracked from the API response */
  apiVersion: number

  setSaving: (v: boolean) => void

  // actions
  hydrate: () => Promise<void>
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
  publish: () => Promise<void>

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
  })
  void schedulePersist()
}

let persistTimer: ReturnType<typeof setTimeout> | null = null

/** Auto-save (debounce 800ms): local bridge first, then API */
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
        toast.error(`Gagal menyimpan draft: ${(err as Error).message}`)
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
    searchQuery: "",
    apiWorkflowId: null,
    apiVersion: 0,

    hydrate: async () => {
      try {
        // Load draft body from local bridge (localStorage)
        const draft = loadDraftLocal()
        const apiId = loadApiId()
        const apiVer = loadApiVersion()
        if (draft) {
          const { nodes: n, edges: e } = materialize(draft, true)
          set({
            workflow: draft,
            nodes: n,
            edges: e,
            validation: validateWorkflow(draft),
            apiWorkflowId: apiId,
            apiVersion: apiVer ?? 0,
          })
        } else {
          set({ apiWorkflowId: apiId, apiVersion: apiVer ?? 0 })
        }
      } catch (err) {
        toast.error(`Gagal memuat draft: ${(err as Error).message}`)
      } finally {
        set({ isHydrated: true })
      }
    },

    setNodes: (nodes) => {
      const wf = buildSnapshot(get().workflow, nodes, get().edges)
      const history = [...get().history, get().workflow].slice(-MAX_HISTORY)
      set({ nodes, workflow: wf, history, future: [] })
      get().revalidate()
      void schedulePersist()
    },

    setEdges: (edges) => {
      const wf = buildSnapshot(get().workflow, get().nodes, edges)
      const history = [...get().history, get().workflow].slice(-MAX_HISTORY)
      set({
        edges: toFlowEdges(wf),
        workflow: wf,
        history,
        future: [],
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
      set({ workflow: { ...get().workflow, ...patch } })
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

    persist: async () => {
      const { workflow, nodes, edges, apiWorkflowId, apiVersion } = get()
      const snapshot = buildSnapshot(workflow, nodes, edges)

      // 1. Always save the draft body to the local bridge first (non-blocking
      //    for the API call; ensures the graph survives refresh even if offline).
      saveDraftLocal(snapshot)

      // 2. Sync the workflow root to the API.
      if (!apiWorkflowId) {
        // First save — create the workflow root via the Builder API.
        const created = await createWorkflowApi({
          slug: snapshot.slug,
          name: snapshot.name,
          description: snapshot.description,
        })
        set({ apiWorkflowId: created.id, apiVersion: created.version })
        saveApiId(created.id)
        saveApiVersion(created.version)
      } else {
        // Subsequent save — bump the optimistic version.
        const updated = await updateWorkflowApi({
          id: apiWorkflowId,
          version: apiVersion,
        })
        set({ apiVersion: updated.version })
        saveApiVersion(updated.version)
      }

      set({ isSaving: false, lastSavedAt: new Date().toISOString() })
    },

    publish: async () => {
      const { workflow, nodes, edges, apiWorkflowId, apiVersion } = get()
      const snapshot = buildSnapshot(workflow, nodes, edges)

      // Ensure there's an API workflow root before publishing.
      let id = apiWorkflowId
      let ver = apiVersion
      if (!id) {
        const created = await createWorkflowApi({
          slug: snapshot.slug,
          name: snapshot.name,
          description: snapshot.description,
        })
        id = created.id
        ver = created.version
        set({ apiWorkflowId: id, apiVersion: ver })
        saveApiId(id)
        saveApiVersion(ver)
      }

      await publishWorkflowApi({
        id,
        version: ver,
        definition: snapshot,
      })
      // After publish, update the local version tracking.
      set({ apiVersion: ver + 1, lastSavedAt: new Date().toISOString() })
      saveApiVersion(ver + 1)
    },

    resetValidation: () => set({ validation: null }),

    revalidate: () => {
      set({ validation: validateWorkflow(get().workflow) })
    },
  }
})
