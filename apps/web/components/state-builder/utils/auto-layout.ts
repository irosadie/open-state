import type { Edge, Node } from "@xyflow/react"
/**
 * Auto-layout menggunakan dagre (PRD section 14: automatic layout).
 * Mengatur posisi node menjadi hierarki terurut dari START.
 * Dimensi disesuaikan per bentuk node (oval/diamond/kotak).
 */
import dagre from "dagre"

const NODE_DIMENSION: Record<string, { width: number; height: number }> = {
  start: { width: 150, height: 44 },
  end: { width: 150, height: 44 },
  decision: { width: 240, height: 120 },
  state: { width: 220, height: 72 },
  event: { width: 220, height: 64 },
}

const FALLBACK = { width: 200, height: 64 }

/** Terapkan layout hierarki (top-down) pada nodes & edges */
export function getLayoutedNodes(
  nodes: Node[],
  edges: Edge[],
  direction: "TB" | "LR" = "TB",
): Node[] {
  const g = new dagre.graphlib.Graph()
  g.setDefaultEdgeLabel(() => ({}))
  g.setGraph({ rankdir: direction, nodesep: 60, ranksep: 80 })

  for (const node of nodes) {
    const dim = NODE_DIMENSION[node.type ?? "state"] ?? FALLBACK
    g.setNode(node.id, { width: dim.width, height: dim.height })
  }
  for (const edge of edges) {
    g.setEdge(edge.source, edge.target)
  }

  dagre.layout(g)

  return nodes.map((node) => {
    const pos = g.node(node.id)
    if (!pos) return node
    const dim = NODE_DIMENSION[node.type ?? "state"] ?? FALLBACK
    return {
      ...node,
      position: {
        x: pos.x - dim.width / 2,
        y: pos.y - dim.height / 2,
      },
    }
  })
}
