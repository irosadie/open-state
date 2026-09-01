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
import {
  useCompareWorkflowVersions,
  useWorkflowsSimulate,
  useWorkflowsVersions,
} from "$/hooks/transactions/use-workflow"
import type React from "react"
import { useCallback, useEffect, useMemo, useRef, useState } from "react"
import { edgeTypes } from "./edges/edges"
import { nodeTypes } from "./nodes/nodes"
import { Palette } from "./palette"
import { PropertiesPanel } from "./properties-panel"
import { SimulationPanel, type SimulationRunInput } from "./simulation-panel"
import { EXAMPLE_WORKFLOWS, useStateBuilderStore } from "./state-builder.store"
import { Toolbar } from "./toolbar"
import { getLayoutedNodes } from "./utils/auto-layout"
import { Toaster, toast } from "./utils/toast"
import { parseWorkflowExport } from "./utils/workflow-export"
import { downloadWorkflow } from "./utils/workflow-export"
import { VersionHistoryPanel } from "./version-history-panel"

const nodeColorByType: Record<string, string> = {
  start: "#25a45f",
  state: "#1f8fca",
  decision: "#f39e1a",
  event: "#1e8cc1",
  end: "#d83838",
}

interface StateBuilderProps {
  workflowId?: string
  projectId?: string
}

function BuilderInner({ workflowId, projectId }: StateBuilderProps) {
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
  const saveError = useStateBuilderStore((s) => s.saveError)
  const saveConflict = useStateBuilderStore((s) => s.saveConflict)
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
  const publish = useStateBuilderStore((s) => s.publish)
  const apiWorkflowId = useStateBuilderStore((s) => s.apiWorkflowId)
  const activeProjectId = useStateBuilderStore((s) => s.activeProjectId)
  const removeNode = useStateBuilderStore((s) => s.removeNode)
  const removeTransition = useStateBuilderStore((s) => s.removeTransition)
  const simulationOpen = useStateBuilderStore((s) => s.simulationOpen)
  const simulationInitialContextText = useStateBuilderStore(
    (s) => s.simulationInitialContextText,
  )
  const simulationEvents = useStateBuilderStore((s) => s.simulationEvents)
  const simulationResult = useStateBuilderStore((s) => s.simulationResult)
  const simulationError = useStateBuilderStore((s) => s.simulationError)
  const simulationIsRunning = useStateBuilderStore((s) => s.simulationIsRunning)
  const simulationSelectedSequence = useStateBuilderStore(
    (s) => s.simulationSelectedSequence,
  )
  const simulationFocusTarget = useStateBuilderStore(
    (s) => s.simulationFocusTarget,
  )
  const simulationStale = useStateBuilderStore((s) => s.simulationStale)
  const simulationFingerprint = useStateBuilderStore(
    (s) => s.simulationFingerprint,
  )
  const openSimulation = useStateBuilderStore((s) => s.openSimulation)
  const closeSimulation = useStateBuilderStore((s) => s.closeSimulation)
  const setSimulationInitialContextText = useStateBuilderStore(
    (s) => s.setSimulationInitialContextText,
  )
  const addSimulationEvent = useStateBuilderStore((s) => s.addSimulationEvent)
  const updateSimulationEvent = useStateBuilderStore(
    (s) => s.updateSimulationEvent,
  )
  const removeSimulationEvent = useStateBuilderStore(
    (s) => s.removeSimulationEvent,
  )
  const setSimulationIsRunning = useStateBuilderStore(
    (s) => s.setSimulationIsRunning,
  )
  const setSimulationError = useStateBuilderStore((s) => s.setSimulationError)
  const setSimulationResult = useStateBuilderStore((s) => s.setSimulationResult)
  const selectSimulationStep = useStateBuilderStore(
    (s) => s.selectSimulationStep,
  )
  const resetSimulation = useStateBuilderStore((s) => s.resetSimulation)
  const markSimulationStale = useStateBuilderStore((s) => s.markSimulationStale)
  const getSimulationSnapshot = useStateBuilderStore(
    (s) => s.getSimulationSnapshot,
  )
  const simulationMutation = useWorkflowsSimulate()
  const [isPublishing, setIsPublishing] = useState(false)
  const [showVersions, setShowVersions] = useState(false)
  const [baseVersion, setBaseVersion] = useState<number | null>(null)
  const [targetVersion, setTargetVersion] = useState<number | null>(null)
  const versionsQuery = useWorkflowsVersions({
    id: apiWorkflowId ?? "",
    projectId: activeProjectId,
    enabled: showVersions && Boolean(apiWorkflowId),
  })
  const diffQuery = useCompareWorkflowVersions(
    apiWorkflowId ?? "",
    baseVersion,
    targetVersion,
    activeProjectId,
  )

  const { screenToFlowPosition } = useReactFlow()
  const fileInputRef = useRef<HTMLInputElement>(null)

  // Hydrate the draft from the Builder API on first mount.
  useEffect(() => {
    void hydrate(workflowId, projectId)
  }, [hydrate, projectId, workflowId])

  useEffect(() => {
    if (
      workflowId ||
      !apiWorkflowId ||
      window.location.pathname !== "/state-builder"
    ) {
      return
    }
    window.history.replaceState(null, "", `/state-builder/${apiWorkflowId}`)
  }, [apiWorkflowId, workflowId])

  useEffect(() => {
    const versions = versionsQuery.data ?? []
    if (!showVersions || versions.length === 0) return
    setTargetVersion((current) => current ?? versions[0]?.versionNo ?? null)
    setBaseVersion((current) => current ?? versions[1]?.versionNo ?? null)
  }, [showVersions, versionsQuery.data])

  const simulationSnapshotFingerprint = JSON.stringify(getSimulationSnapshot())

  useEffect(() => {
    if (
      simulationResult &&
      simulationFingerprint &&
      simulationFingerprint !== simulationSnapshotFingerprint
    ) {
      markSimulationStale()
    }
  }, [
    markSimulationStale,
    simulationFingerprint,
    simulationResult,
    simulationSnapshotFingerprint,
  ])

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
        useStateBuilderStore.getState().setSaving(true)
        void persist()
          .then(() => toast.success("Draft tersimpan"))
          .catch((err) =>
            toast.error(`Gagal menyimpan draft: ${getErrorMessage(err)}`),
          )
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
    useStateBuilderStore.getState().setSaving(true)
    void persist()
      .then(() => toast.success("Draft tersimpan"))
      .catch((err) =>
        toast.error(`Gagal menyimpan draft: ${getErrorMessage(err)}`),
      )
  }, [persist])

  const handlePublish = useCallback(async () => {
    if (!validation?.valid) {
      revalidate()
      toast.error("Workflow belum valid. Perbaiki error sebelum publish.")
      return
    }
    setIsPublishing(true)
    try {
      const published = await publish()
      toast.success(`Workflow berhasil dipublish v${published.versionNo}`)
      setShowVersions(true)
    } catch (err) {
      const message = getErrorMessage(err)
      toast.error(`Publish gagal: ${message}`)
    } finally {
      setIsPublishing(false)
    }
  }, [publish, revalidate, validation?.valid])

  const handleSimulationRun = useCallback(
    (input: SimulationRunInput) => {
      const snapshot = getSimulationSnapshot()
      const fingerprint = JSON.stringify(snapshot)
      setSimulationError(null)
      setSimulationIsRunning(true)
      simulationMutation.mutate(
        {
          definition: snapshot as unknown as Record<string, unknown>,
          initialContext: input.initialContext,
          events: input.events,
        },
        {
          onSuccess: (result) => setSimulationResult(result, fingerprint),
          onError: (error) => {
            const message =
              error instanceof Error
                ? error.message
                : typeof error === "object" &&
                    error !== null &&
                    "message" in error
                  ? String((error as { message: unknown }).message)
                  : "Simulation gagal dijalankan"
            setSimulationError(message)
          },
        },
      )
    },
    [
      getSimulationSnapshot,
      setSimulationError,
      setSimulationIsRunning,
      setSimulationResult,
      simulationMutation,
    ],
  )

  const handleOpenSimulation = useCallback(() => {
    setShowVersions(false)
    openSimulation()
  }, [openSimulation])

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

  const focusedNodeIds = useMemo(
    () => new Set(simulationFocusTarget?.nodeIds ?? []),
    [simulationFocusTarget],
  )
  const simulationNodes = useMemo(
    () =>
      visibleNodes.map((node) => ({
        ...node,
        selected: Boolean(node.selected || focusedNodeIds.has(node.id)),
      })),
    [focusedNodeIds, visibleNodes],
  )
  const simulationEdges = useMemo(
    () =>
      edges.map((edge) => ({
        ...edge,
        selected: Boolean(
          edge.selected || edge.id === simulationFocusTarget?.transitionId,
        ),
      })),
    [edges, simulationFocusTarget?.transitionId],
  )

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
    <div className="flex h-full flex-col" data-testid="builder-root">
      <Toaster />
      <Toolbar
        validation={validation}
        onValidate={handleValidate}
        onAutoLayout={handleAutoLayout}
        onSave={handleSave}
        onPublish={() => void handlePublish()}
        onSimulate={handleOpenSimulation}
        onVersions={() => setShowVersions(true)}
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
        saveError={saveError}
        saveConflict={saveConflict}
        isPublishing={isPublishing}
        hasWorkflowId={Boolean(apiWorkflowId)}
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
            nodes={simulationNodes}
            edges={simulationEdges}
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
            <div
              className="sb-scroll absolute bottom-3 left-3 z-10 max-h-40 w-80 overflow-y-auto rounded-lg border border-danger-200 bg-white p-2 shadow-lg"
              data-testid="builder-validation"
            >
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

          {showVersions ? (
            <VersionHistoryPanel
              versions={versionsQuery.data ?? []}
              isLoading={versionsQuery.isLoading}
              error={
                versionsQuery.error ? "Gagal memuat version history" : null
              }
              baseVersion={baseVersion}
              targetVersion={targetVersion}
              onBaseVersionChange={setBaseVersion}
              onTargetVersionChange={setTargetVersion}
              diff={diffQuery.data}
              isDiffLoading={diffQuery.isLoading}
              onClose={() => setShowVersions(false)}
            />
          ) : null}

          {simulationOpen ? (
            <SimulationPanel
              initialContextText={simulationInitialContextText}
              events={simulationEvents}
              result={simulationResult}
              error={simulationError}
              isRunning={simulationIsRunning}
              stale={simulationStale}
              selectedSequence={simulationSelectedSequence}
              onInitialContextChange={setSimulationInitialContextText}
              onAddEvent={addSimulationEvent}
              onUpdateEvent={updateSimulationEvent}
              onRemoveEvent={removeSimulationEvent}
              onRun={handleSimulationRun}
              onReset={resetSimulation}
              onSelectStep={selectSimulationStep}
              onClose={closeSimulation}
            />
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

export function StateBuilder({
  workflowId,
  projectId,
}: StateBuilderProps = {}) {
  return (
    <ReactFlowProvider>
      <BuilderInner workflowId={workflowId} projectId={projectId} />
    </ReactFlowProvider>
  )
}
