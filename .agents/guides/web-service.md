# Guide: Web Service (`apps/web/services/`)

## Folder Contract

The `services/` folder only contains **axios instance** with configuration and interceptors.
No service functions per endpoint — hooks call axios **directly**.

✅ Allowed:
- Setup axios instance with internal proxy base URL from `configs/env.ts`
- Response interceptor (unwrap data, handle 401)

❌ Forbidden:
- Service functions per endpoint here — put them in hook files
- Contains JSX or React hooks
- Hardcode base URL — use env variable
- Handle domain errors here
- Attach bearer token browser-side for requests that go through internal proxy

---

## Conventions

### Structure

```
services/
└── axios/
    └── index.ts    → Axios instance + interceptors
```

### Axios Instance

```typescript
// services/axios/index.ts
import { authConfig } from "$/configs/auth"
import { env } from "$/configs/env"
import _axios from 'axios'

export const axios = _axios.create({
  baseURL: env.apiBaseUrl,
  headers: { 'Content-Type': 'application/json' },
})

// Response: unwrap data
axios.interceptors.response.use(
  (response) => {
    const { data: responseData, meta } = response.data

    // List response (array) → DataTableResponse shape
    if (Array.isArray(responseData)) {
      return {
        list: responseData,
        meta: {
          pagination: meta,
          cursor: null,
        },
      }
    }

    // Single response → return data directly
    return responseData
  },
  (error) => {
    // 401 → redirect to login
    if (error.response?.status === 401) {
      window.location.href = authConfig.loginPath
    }
    return Promise.reject(error?.response?.data || error)
  },
)
```

---

## Usage in Hooks

Hooks import axios instance from `$/services/axios` and directly call endpoints:

```typescript
// hooks/transactions/use-payment-methods/use-data-table.ts
import { axios } from '$/services/axios'
import { apiRouters } from '$/constants'

const fetchDataTable = async (args) => {
  const [, { page, limit, filter }] = args.queryKey

  // Call axios directly — no need for service function
  const result = await axios<DataTableResponse<DataTypeProps>>({
    method: 'GET',
    url: apiRouters.paymentMethods.index,
    params: { ...filter, page, limit },
  })

  return result
}
```

---

## Additional Rules

- Only one axios instance (singleton)
- Import in hooks as `import { axios } from '$/services/axios'`
- Browser frontend hits `/api/proxy/*`; token injection and refresh happen server-side in route `app/api/proxy/[...path]/route.ts`
- Files must end with newline
