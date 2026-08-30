# Frontend authorization policy

Frontend authorization is derived from the tenant-scoped `GET /api/auth/me`
response. The response is loaded by `AuthorizationProvider` with the active
`X-Tenant-ID`; the API remains the final authorization boundary.

## Registering a route

Add a `RoutePolicy` entry in `route-policy.ts` for every protected browser
route. Use the least-privilege read permission, keep dynamic segments in the
regular expression, and set `landing: true` only for a safe post-login
destination. Unknown routes are denied by default.

```ts
{
  id: "workflows",
  pattern: /^\/admin\/workflows(?:\/.*)?$/,
  requiredPermissions: ["workflow:read"],
  landing: true,
}
```

## Registering an action

Add an `ActionPolicy` entry for each sensitive action and use
`<PermissionGate action="...">` around the control. Keep the mutation handler
inside the gated control so a denied user cannot invoke it from the UI.

```ts
{ id: "workflow:publish", requiredPermissions: ["workflow:publish"] }
```

Use `enabled: authorization.hasPermission("resource:read")` for protected
queries. Do not fetch protected data and then hide it after the response.
Permission checks use exact permissions and the server-compatible
`resource:*` wildcard; role names must not be checked in page components.
