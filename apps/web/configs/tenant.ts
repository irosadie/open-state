// Tenant identity for the capability admin UI (PRD §4, §96).
//
// The capability admin API is tenant-scoped and requires an `X-Tenant-ID` header.
// Until a full multi-tenant session exists, the admin UI uses a single configured
// tenant id (defaulting to the local dev value). Secret values are never sent;
// only the tenant id is transmitted as a request header.
export const tenantConfig = {
  tenantId:
    process.env.NEXT_PUBLIC_TENANT_ID ??
    process.env.NEXT_PUBLIC_API_TENANT_ID ??
    "00000000-0000-0000-0000-000000000001",
} as const
