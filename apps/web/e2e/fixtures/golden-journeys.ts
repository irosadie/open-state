export type GoldenRole = "EDITOR" | "OPERATOR" | "VIEWER"

export type GoldenIdentity = {
  id: string
  email: string
  role: GoldenRole
  tenantId: string
}

export type GoldenJourney = {
  id: string
  actor: GoldenIdentity
  activeTenantId: string
  startingResources: Record<string, string>
  actions: string[]
  expectedCheckpoints: string[]
}

export const GOLDEN_IDS = {
  tenantA: "00000000-0000-0000-0000-0000000000a1",
  tenantB: "00000000-0000-0000-0000-0000000000b1",
  editorUser: "10000000-0000-0000-0000-000000000001",
  operatorUser: "10000000-0000-0000-0000-000000000002",
  viewerUser: "10000000-0000-0000-0000-000000000003",
  sentinelUser: "10000000-0000-0000-0000-000000000004",
  builderWorkflow: "30000000-0000-0000-0000-000000000001",
  invalidWorkflow: "30000000-0000-0000-0000-000000000002",
  staleWorkflow: "30000000-0000-0000-0000-000000000003",
  runningInstance: "50000000-0000-0000-0000-000000000001",
  suspendedInstance: "50000000-0000-0000-0000-000000000002",
  failedInstance: "50000000-0000-0000-0000-000000000003",
  sentinelInstance: "50000000-0000-0000-0000-000000000004",
} as const

export const GOLDEN_IDENTITIES = {
  editor: {
    id: GOLDEN_IDS.editorUser,
    email: "editor.golden@tenant-a.invalid",
    role: "EDITOR",
    tenantId: GOLDEN_IDS.tenantA,
  },
  operator: {
    id: GOLDEN_IDS.operatorUser,
    email: "operator.golden@tenant-a.invalid",
    role: "OPERATOR",
    tenantId: GOLDEN_IDS.tenantA,
  },
  viewer: {
    id: GOLDEN_IDS.viewerUser,
    email: "viewer.golden@tenant-a.invalid",
    role: "VIEWER",
    tenantId: GOLDEN_IDS.tenantA,
  },
  sentinel: {
    id: GOLDEN_IDS.sentinelUser,
    email: "sentinel.golden@tenant-b.invalid",
    role: "VIEWER",
    tenantId: GOLDEN_IDS.tenantB,
  },
} satisfies Record<string, GoldenIdentity>

export const GOLDEN_JOURNEYS: Record<string, GoldenJourney> = {
  builder: {
    id: "builder-editor-lifecycle",
    actor: GOLDEN_IDENTITIES.editor,
    activeTenantId: GOLDEN_IDS.tenantA,
    startingResources: {
      workflowId: GOLDEN_IDS.builderWorkflow,
      invalidWorkflowId: GOLDEN_IDS.invalidWorkflow,
      staleWorkflowId: GOLDEN_IDS.staleWorkflow,
    },
    actions: [
      "sign in as Editor",
      "open the seeded Builder draft",
      "rename the INTAKE node to INTAKE_EDITED",
      "save and reload the draft",
      "publish the valid draft",
      "open newest-first version history",
      "compare the two newest versions",
      "attempt invalid publish",
      "attempt stale save",
    ],
    expectedCheckpoints: [
      "saved draft retains INTAKE_EDITED",
      "published version is newer than v2",
      "version history is newest first",
      "graph diff identifies the changed INTAKE node",
      "invalid publish sends no mutation",
      "stale save shows reload guidance and keeps the local graph",
    ],
  },
  operator: {
    id: "operator-runtime-inspection",
    actor: GOLDEN_IDENTITIES.operator,
    activeTenantId: GOLDEN_IDS.tenantA,
    startingResources: {
      runningInstanceId: GOLDEN_IDS.runningInstance,
      suspendedInstanceId: GOLDEN_IDS.suspendedInstance,
      failedInstanceId: GOLDEN_IDS.failedInstance,
      sentinelInstanceId: GOLDEN_IDS.sentinelInstance,
    },
    actions: [
      "sign in as Operator",
      "open the tenant-scoped Runtime Inspector",
      "inspect workflow/version, state, context, timeline, and sanitized debug metadata",
      "confirm suspend, resume, and retry",
      "open denied management and sentinel routes",
    ],
    expectedCheckpoints: [
      "only tenant-A instances are discoverable",
      "timeline entries are chronological",
      "debug view contains only sanctioned sanitized metadata",
      "each command persists its lifecycle result and audit actor/action",
      "denied routes do not fetch protected data",
    ],
  },
}

export function fixturePassword(): string {
  const password = process.env.E2E_FIXTURE_PASSWORD
  if (!password) {
    throw new Error(
      "E2E_FIXTURE_PASSWORD must be set; fixture manifests never contain passwords",
    )
  }
  return password
}
