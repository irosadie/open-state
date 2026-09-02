export const apiRouters = {
  health: {
    check: "/health",
  },
  auth: {
    login: "/auth/login",
    register: "/auth/register",
    logout: "/auth/logout",
    me: "/auth/me",
  },
  capabilities: {
    index: "/capabilities",
    show: "/capabilities/:id",
    update: "/capabilities/:id",
    delete: "/capabilities/:id",
    bindings: "/capabilities/:id/bindings",
    test: "/capabilities/:id/test",
  },
  bindings: {
    delete: "/bindings/:id",
  },
  workflows: {
    index: "/workflows",
    create: "/workflows",
    show: "/workflows/:id",
    update: "/workflows/:id",
    publish: "/workflows/:id/publish",
    versions: "/workflows/:id/versions",
    version: "/workflows/:id/versions/:versionNo",
    compare: "/workflows/:id/versions/compare",
    simulate: "/workflows/simulate",
  },
  intents: {
    index: "/intents",
  },
  projects: {
    index: "/projects",
  },
  mcpConnections: {
    index: "/projects/:projectId/mcp-connections",
    show: "/projects/:projectId/mcp-connections/:id",
    create: "/projects/:projectId/mcp-connections",
    update: "/projects/:projectId/mcp-connections/:id",
    delete: "/projects/:projectId/mcp-connections/:id",
    enable: "/projects/:projectId/mcp-connections/:id/enable",
    disable: "/projects/:projectId/mcp-connections/:id/disable",
    test: "/projects/:projectId/mcp-connections/:id/test",
    diagnose: "/projects/:projectId/mcp-connections/:id/diagnose",
    resetHealth: "/projects/:projectId/mcp-connections/:id/reset-health",
    rotateCredential:
      "/projects/:projectId/mcp-connections/:id/credentials/rotate",
    revokeCredential:
      "/projects/:projectId/mcp-connections/:id/credentials/revoke",
    credentialStatus:
      "/projects/:projectId/mcp-connections/:id/credentials/status",
    oauthStart: "/projects/:projectId/mcp-connections/:id/oauth/start",
    oauthDisconnect:
      "/projects/:projectId/mcp-connections/:id/oauth/disconnect",
    oauthStatus: "/projects/:projectId/mcp-connections/:id/oauth/status",
    tools: "/projects/:projectId/mcp-connections/:id/tools",
    refreshTools: "/projects/:projectId/mcp-connections/:id/tools/refresh",
    updateTool: "/projects/:projectId/mcp-connections/:id/tools/:toolName",
  },
  projectMCPBindings: {
    options: "/projects/:projectId/mcp-tool-options",
    list: "/projects/:projectId/mcp-capability-bindings",
    upsert: "/projects/:projectId/capabilities/:capabilityId/mcp-binding",
    delete: "/projects/:projectId/capabilities/:capabilityId/mcp-binding",
  },
  audit: {
    index: "/audit",
  },
  runtimeInstances: {
    index: "/runtime/instances",
    show: "/runtime/instances/:id",
    debugTrace: "/runtime/instances/:id/debug-trace",
  },
  admin: {
    apiKeys: "/api-keys",
    revokeAPIKey: "/api-keys/:id/revoke",
    tenant: "/admin/tenant",
    members: "/admin/members",
    memberRole: "/admin/members/:userId/role",
    member: "/admin/members/:userId",
    instances: "/admin/instances",
    suspendInstance: "/admin/instances/:id/suspend",
    resumeInstance: "/admin/instances/:id/resume",
    retryInstance: "/admin/instances/:id/retry",
    events: "/admin/events",
    event: "/admin/events/:eventId",
  },
}
