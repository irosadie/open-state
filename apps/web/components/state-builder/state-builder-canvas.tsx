"use client"

import {
  Background,
  BackgroundVariant,
  type Connection,
  Controls,
  MarkerType,
  MiniMap,
  type OnNodesChange,
  ReactFlow,
  ReactFlowProvider,
  applyNodeChanges,
  useReactFlow,
} from "@xyflow/react"
import "@xyflow/react/dist/style.css"
import type React from "react"
import { useCallback, useEffect, useMemo, useRef } from "react"
import { edgeTypes } from "./edges/edges"
import { nodeTypes } from "./nodes/nodes"
import { Palette } from "./palette"
import { PropertiesPanel } from "./properties-panel"
import { EXAMPLE_WORKFLOWS, useStateBuilderStore } from "./state-builder.store"
import { Toolbar } from "./toolbar"
import { getLayoutedNodes } from "./utils/auto-layout"
import { Toaster, toast } from "./utils/toast"
import { parseWorkflowExport } from "./utils/workflow-export"
import { downloadWorkflow } from "./utils/workflow-export"

const nodeColorByType: Record<string, string> = {
  start: "#25a45f",
  state: "#1f8fca",
  decision: "#f39e1a",
  event: "#1e8cc1",
  end: "#d83838",
}

function BuilderInner() {
  const nodes = useStateBuilderStore((s) => s.nodes)
  const edges = useStateBuilderStore((s) => s.edges)
  const setNodes = useStateBuilderStore((s) => s.setNodes)
  const addNode = useStateBuilderStore((s) => s.addNode)
  const addNodeAt = useStateBuilderStore((s) => s.addNodeAt)
  const addTransition = useStateBuilderStore((s) => s.addTransition)
  const selectNode = useStateBuilderStore((s) => s.selectNode)
  const selectEdge = useStateBuilderStore((s) => s.selectEdge)
  const selectedNodeId = useStateBuilderStore((s) => s.selectedNodeId)
  const selectedEdgeId = useStateBuilderStore((s) => s.selectedEdgeId)
  const validation = useStateBuilderStore((s) => s.validation)
  const revalidate = useStateBuilderStore((s) => s.revalidate)
  const workflow = useStateBuilderStore((s) => s.workflow)
  const history = useStateBuilderStore((s) => s.history)
  const future = useStateBuilderStore((s) => s.future)
  const isHydrated = useStateBuilderStore((s) => s.isHydrated)
  const isSaving = useStateBuilderStore((s) => s.isSaving)
  const lastSavedAt = useStateBuilderStore((s) => s.lastSavedAt)
  const searchQuery = useStateBuilderStore((s) => s.searchQuery)
  const setSearchQuery = useStateBuilderStore((s) => s.setSearchQuery)
  const hydrate = useStateBuilderStore((s) => s.hydrate)
  const undo = useStateBuilderStore((s) => s.undo)
  const redo = useStateBuilderStore((s) => s.redo)
  const loadWorkflow = useStateBuilderStore((s) => s.loadWorkflow)
  const newWorkflow = useStateBuilderStore((s) => s.newWorkflow)
  const resetToWorkflow = useStateBuilderStore((s) => s.resetToWorkflow)
  const persist = useStateBuilderStore((s) => s.persist)
  const removeNode = useStateBuilderStore((s) => s.removeNode)
  const removeTransition = useStateBuilderStore((s) => s.removeTransition)

  const { screenToFlowPosition } = useReactFlow()
  const fileInputRef = useRef<HTMLInputElement>(null)

  // Load draft dari PGlite saat pertama mount
  useEffect(() => {
    void hydrate()
  }, [hydrate])

  // Keyboard shortcuts
  useEffect(() => {
    const onKeyDown = (e: KeyboardEvent) => {
      const target = e.target as HTMLElement
      const isTyping =
        target.tagName === "INPUT" ||
        target.tagName === "TEXTAREA" ||
        target.isContentEditable

      // Ctrl+Z undo / Ctrl+Shift+Z atau Ctrl+Y redo
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === "z") {
        if (isTyping) return
        e.preventDefault()
        if (e.shiftKey) redo()
        else undo()
        return
      }
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === "y") {
        if (isTyping) return
        e.preventDefault()
        redo()
        return
      }
      // Ctrl+S save
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === "s") {
        if (isTyping) return
        e.preventDefault()
        void persist().then(() => toast.success("Draft tersimpan"))
        return
      }
      // Delete hapus node/edge terpilih
      if (e.key === "Delete" || e.key === "Backspace") {
        if (isTyping) return
        const s = useStateBuilderStore.getState()
        if (s.selectedNodeId) {
          e.preventDefault()
          removeNode(s.selectedNodeId)
          toast.info("Node dihapus")
        } else if (s.selectedEdgeId) {
          e.preventDefault()
          removeTransition(s.selectedEdgeId)
          toast.info("Transisi dihapus")
        }
      }
    }
    window.addEventListener("keydown", onKeyDown)
    return () => window.removeEventListener("keydown", onKeyDown)
  }, [undo, redo, persist, removeNode, removeTransition])

  const _selectedNode = useMemo(
    () => nodes.find((n) => n.id === selectedNodeId) ?? null,
    [nodes, selectedNodeId],
  )
  const _selectedEdge = useMemo(
    () => edges.find((e) => e.id === selectedEdgeId) ?? null,
    [edges, selectedEdgeId],
  )

  const onNodesChange: OnNodesChange = useCallback(
    (changes) => {
      const next = applyNodeChanges(changes, nodes)
      setNodes(next)
    },
    [nodes, setNodes],
  )

  const onConnect = useCallback(
    (connection: Connection) => {
      if (!connection.source || !connection.target) return
      addTransition(connection.source, connection.target)
    },
    [addTransition],
  )

  const onDragOver = useCallback((event: React.DragEvent) => {
    event.preventDefault()
    event.dataTransfer.dropEffect = "move"
  }, [])

  const onDrop = useCallback(
    (event: React.DragEvent) => {
      event.preventDefault()
      const kind = event.dataTransfer.getData("application/x-workflow-node") as
        | "START"
        | "STATE"
        | "DECISION"
        | "EVENT"
        | "END"
      if (!kind) return
      const position = screenToFlowPosition({
        x: event.clientX,
        y: event.clientY,
      })
      addNodeAt(kind, position)
    },
    [screenToFlowPosition, addNodeAt],
  )

  const handleAutoLayout = useCallback(() => {
    setNodes(getLayoutedNodes(nodes, edges, "TB"))
  }, [nodes, edges, setNodes])

  const handleValidate = useCallback(() => {
    revalidate()
    toast.info("Validasi dijalankan")
  }, [revalidate])

  const handleSave = useCallback(() => {
    void persist().then(() => toast.success("Draft tersimpan"))
  }, [persist])

  const handleExport = useCallback(() => {
    downloadWorkflow(useStateBuilderStore.getState().workflow)
    toast.success("Export JSON diunduh")
  }, [])

  const handleImportFile = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const file = e.target.files?.[0]
      if (!file) return
      const reader = new FileReader()
      reader.onload = () => {
        try {
          const { workflow: wf } = parseWorkflowExport(String(reader.result))
          loadWorkflow(wf)
          toast.success(`Workflow "${wf.name}" dimuat`)
        } catch (err) {
          toast.error(`Import gagal: ${(err as Error).message}`)
        }
      }
      reader.readAsText(file)
      e.target.value = ""
    },
    [loadWorkflow],
  )

  const handleNew = useCallback(() => {
    if (
      window.confirm(
        "Buat workflow baru? Perubahan saat ini akan hilang dari canvas (draft lama masih tersimpan di database).",
      )
    ) {
      newWorkflow()
      toast.info("Workflow baru dibuat")
    }
  }, [newWorkflow])

  const handleReset = useCallback(
    (slug: string) => {
      const example = EXAMPLE_WORKFLOWS.find((w) => w.slug === slug)
      if (
        window.confirm(
          `Muat contoh workflow "${example?.name ?? slug}"? Perubahan pada canvas akan ditimpa (draft lama tetap tersimpan).`,
        )
      ) {
        resetToWorkflow(slug)
        toast.info(`Contoh "${example?.name ?? slug}" dimuat`)
      }
    },
    [resetToWorkflow],
  )

  const hasNode = useCallback(
    (kind: string) => workflow.nodes.some((n) => n.kind === kind),
    [workflow.nodes],
  )

  // Filter node yang tampil sesuai search
  const visibleNodes = useMemo(() => {
    if (!searchQuery.trim()) return nodes
    const q = searchQuery.toLowerCase()
    return nodes.filter((n) => {
      const data = n.data as { name?: string; kind?: string }
      return (
        data.name?.toLowerCase().includes(q) ||
        data.kind?.toLowerCase().includes(q)
      )
    })
  }, [nodes, searchQuery])

  const stats = useMemo(
    () => ({
      states: workflow.nodes.filter((n) => n.kind === "STATE").length,
      decisions: workflow.nodes.filter((n) => n.kind === "DECISION").length,
      events: workflow.nodes.filter((n) => n.kind === "EVENT").length,
      transitions: workflow.transitions.length,
      total: workflow.nodes.length,
    }),
    [workflow],
  )

  return (
    <div className="flex h-full flex-col">
      <Toaster />
      <Toolbar
        validation={validation}
        onValidate={handleValidate}
        onAutoLayout={handleAutoLayout}
        onSave={handleSave}
        onExport={handleExport}
        onNewStart={() => addNode("START")}
        onImport={() => fileInputRef.current?.click()}
        onNew={handleNew}
        onReset={handleReset}
        examples={EXAMPLE_WORKFLOWS.map((w) => ({
          slug: w.slug,
          name: w.name,
        }))}
        onUndo={undo}
        onRedo={redo}
        canUndo={history.length > 0}
        canRedo={future.length > 0}
        isSaving={isSaving}
        lastSavedAt={lastSavedAt}
        stats={stats}
        hasNode={hasNode}
      />

      <input
        ref={fileInputRef}
        type="file"
        accept="application/json,.json"
        className="hidden"
        onChange={handleImportFile}
      />

      <div className="flex flex-1 overflow-hidden">
        {/* Palette */}
        <aside className="sb-scroll w-44 shrink-0 overflow-y-auto border-r border-slate-200 bg-slate-50">
          <Palette />
        </aside>

        {/* Canvas */}
        <main className="relative flex-1">
          {!isHydrated ? (
            <div className="absolute inset-0 z-20 flex items-center justify-center bg-white/70">
              <span className="text-sm text-slate-500">Memuat draft…</span>
            </div>
          ) : null}

          <ReactFlow
            nodes={visibleNodes}
            edges={edges}
            onNodesChange={onNodesChange}
            onConnect={onConnect}
            onDrop={onDrop}
            onDragOver={onDragOver}
            nodeTypes={nodeTypes}
            edgeTypes={edgeTypes}
            onNodeClick={(_e, n) => selectNode(n.id)}
            onPaneClick={() => {
              selectNode(null)
              selectEdge(null)
            }}
            onEdgeClick={(_e, e) => selectEdge(e.id)}
            fitView
            fitViewOptions={{ padding: 0.2 }}
            minZoom={0.2}
            defaultEdgeOptions={{
              type: "transition",
              markerEnd: {
                type: MarkerType.ArrowClosed,
                width: 18,
                height: 18,
                color: "#64748b",
              },
            }}
          >
            <Background variant={BackgroundVariant.Dots} gap={20} size={1} />
            <Controls />
            <MiniMap
              nodeColor={(n) => nodeColorByType[n.type ?? "state"] ?? "#1f8fca"}
              maskColor="rgba(240, 240, 240, 0.6)"
            />
          </ReactFlow>

          {/* Search box */}
          <div className="absolute left-1/2 top-3 z-10 w-72 -translate-x-1/2">
            <input
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Cari node… (nama / type)"
              className="w-full rounded-lg border border-slate-300 bg-white px-3 py-1.5 text-sm shadow-md outline-none focus:border-primary-400 focus:ring-1 focus:ring-primary-200"
            />
          </div>

          {/* Validation issues overlay */}
          {validation && validation.issues.length > 0 ? (
            <div className="sb-scroll absolute bottom-3 left-3 z-10 max-h-40 w-80 overflow-y-auto rounded-lg border border-danger-200 bg-white p-2 shadow-lg">
              <p className="mb-1 text-xs font-semibold text-slate-600">
                Validation ({validation.issues.length})
              </p>
              {validation.issues.map((issue, i) => (
                <div
                  key={`${issue.code}-${issue.nodeId ?? ""}-${issue.edgeId ?? ""}-${i}`}
                  className={`mb-1 rounded px-2 py-1 text-xs ${
                    issue.severity === "error"
                      ? "bg-danger-50 text-danger-700"
                      : "bg-warning-50 text-warning-700"
                  }`}
                >
                  {issue.message}
                </div>
              ))}
            </div>
          ) : null}
        </main>

        {/* Properties Panel */}
        <aside className="sb-scroll w-80 shrink-0 overflow-y-auto border-l border-slate-200 bg-white">
          <PropertiesPanel
            selectedNodeId={selectedNodeId}
            selectedEdgeId={selectedEdgeId}
          />
        </aside>
      </div>
    </div>
  )
}

export function StateBuilder() {
  return (
    <ReactFlowProvider>
      <BuilderInner />
    </ReactFlowProvider>
  )
}
