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
  audit: {
    index: "/audit",
  },
  runtimeInstances: {
    index: "/runtime/instances",
    show: "/runtime/instances/:id",
    debugTrace: "/runtime/instances/:id/debug-trace",
  },
  admin: {
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
