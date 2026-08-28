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
  },
  audit: {
    index: "/audit",
  },
}
