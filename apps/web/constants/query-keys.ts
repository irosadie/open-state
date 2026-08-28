export const queryKeys = {
  health: {
    check: "healthCheck",
  },
  auth: {
    login: "authLogin",
    register: "authRegister",
    logout: "authLogout",
    me: "authMe",
  },
  capabilities: {
    list: "capabilitiesList",
    get: "capabilitiesGet",
    create: "capabilitiesCreate",
    update: "capabilitiesUpdate",
    delete: "capabilitiesDelete",
    bindings: "capabilitiesBindings",
    test: "capabilitiesTest",
  },
  bindings: {
    delete: "bindingsDelete",
  },
  workflows: {
    list: "workflowsList",
    get: "workflowsGet",
    create: "workflowsCreate",
    update: "workflowsUpdate",
    publish: "workflowsPublish",
    versions: "workflowsVersions",
  },
}
