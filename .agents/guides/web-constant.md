# Guide: Web Constant (`apps/web/constants/`)

## Folder Contract

✅ Allowed:
- `api-routers.ts` — all API endpoint URLs (flat object, `:id` path variables)
- `query-keys.ts` — all react-query cache keys (flat strings per operation)
- Other truly global constants (nav items, routes, etc.)

❌ Forbidden:
- Business logic
- Import from hooks, services, or components
- Values that should come from env (use `configs/`)
- Dynamic functions for URLs — use `pathVariable()` utility

---

## Conventions

### `api-routers.ts` — Flat Object with `:id` Path Variables

URL paths with parameters use `:key` placeholder. Use `pathVariable()` utility when calling URLs with parameters.

```typescript
// constants/api-routers.ts
export const apiRouters = {
  paymentMethods: {
    index: '/payment-methods',
    show: '/payment-methods/:id',
    insert: '/payment-methods',
    update: '/payment-methods/:id',
    delete: '/payment-methods/:id',
  },
  users: {
    index: '/users',
    show: '/users/:id',
    insert: '/users',
    update: '/users/:id',
    ban: '/users/:id/ban',
    resendVerification: '/users/resend-verification',
  },
  drivers: {
    index: '/vendors/:vendorId/drivers',
    show: '/vendors/:vendorId/drivers/:id',
    insert: '/vendors/:vendorId/drivers',
    update: '/vendors/:vendorId/drivers/:id',
    delete: '/vendors/:vendorId/drivers/:id',
  },
}
```

**Usage** with `pathVariable()`:

```typescript
import { pathVariable } from '$/utils/path-variable'
import { apiRouters } from '$/constants'

// Single param
const url = pathVariable(apiRouters.paymentMethods.show, { id: '123' })
// → '/payment-methods/123'

// Multiple params
const url = pathVariable(apiRouters.drivers.show, { vendorId: 'v1', id: 'd1' })
// → '/vendors/v1/drivers/d1'
```

---

### `query-keys.ts` — Flat String Values per Operation

Each operation has its own key as a plain string. Not an array, not a function.

```typescript
// constants/query-keys.ts
export const queryKeys = {
  paymentMethods: {
    index: 'paymentMethodsIndex',
    get: 'paymentMethodsGet',
    insert: 'paymentMethodsInsert',
    update: 'paymentMethodsUpdate',
    delete: 'paymentMethodsDelete',
  },
  users: {
    index: 'usersIndex',
    get: 'usersGet',
    insert: 'usersInsert',
    update: 'usersUpdate',
    ban: 'usersBan',
    resendVerification: 'usersResendVerification',
  },
  drivers: {
    index: 'driversIndex',
    get: 'driversGet',
    insert: 'driversInsert',
    update: 'driversUpdate',
    delete: 'driversDelete',
    ban: 'driversBan',
    unban: 'driversUnban',
  },
}
```

**Usage in hooks:**

```typescript
// As string key in queryKey array
queryKey: [queryKeys.paymentMethods.index, { page, limit, filter }]

// Or as mutationKey
mutationKey: [queryKeys.paymentMethods.insert]
```

---

## Additional Rules

- Grouping per domain/resource
- Key string follows format: `{domain}{OperationPascalCase}`
  - Example: `paymentMethodsIndex`, `usersInsert`, `driversLocationHistory`
- URL string without trailing slash
- Add new entries under existing groups, not elsewhere
- Files must end with newline
