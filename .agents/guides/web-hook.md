# Guide: Web Hook (`apps/web/hooks/`)

## Folder Contract

✅ Allowed:
- Wrap `useQuery`, `useMutation` from react-query
- Call `axios` instance **directly** in hooks (no need for separate service functions)
- Use `queryKeys` and `apiRouters` from `constants/`
- Return data, loading state, error state, and mutation handlers
- Utility hooks in `hooks/utility/` (useQueryParam, etc.)

❌ Forbidden:
- Contains JSX or render logic
- One hook for all operations — separate per file
- Hardcode URLs — use `apiRouters` from constants
- Use `any` as type

---

## Conventions

### Folder Structure

```
hooks/
├── transactions/
│   └── use-{domain}/          → one folder per domain/resource
│       ├── use-data-table.ts  → fetch paginated list
│       ├── use-get-one.ts     → fetch single item by id
│       ├── use-insert-one.ts  → create mutation
│       ├── use-update-one.ts  → update mutation
│       ├── use-delete-one.ts  → delete mutation
│       └── index.ts           → re-export all hooks
└── utility/
    └── use-query-param/
        └── use-query-param.ts → wrap searchParams + router
```

### Naming (Required to Follow)

| Operation | Hook | File |
|---------|------|------|
| Fetch list + pagination | `useDataTable` | `use-data-table.ts` |
| Fetch single by ID | `useGetOne` | `use-get-one.ts` |
| Create | `useInsertOne` | `use-insert-one.ts` |
| Update | `useUpdateOne` | `use-update-one.ts` |
| Delete | `useDeleteOne` | `use-delete-one.ts` |

Export alias with domain name: `usePaymentMethodsDataTable`, `usePaymentMethodsInsertOne`, etc.

---

### `useDataTable` — React-Query Hook for Paginated List

`useDataTable` is a **react-query `useQuery` hook** that fetches paginated data from the API.
Not a TanStack Table wrapper.

```typescript
// hooks/transactions/use-payment-methods/use-data-table.ts
import { QueryFunctionContext, useQuery } from '@tanstack/react-query'
import type { PaymentMethodResponseProps as DataTypeProps } from '@vibecoding-starter/types/payment-method-response'
import { axios } from '$/services/axios'
import { apiRouters, queryKeys } from '$/constants'
import { DataTableResponse, ErrorResponse } from '$/types/generals'
import { AxiosError } from 'axios'

const DEFAULT_LIMIT = 10

type DataTableFilterProps = {
  search?: string
  isActive?: boolean
}

type UseDataTableProps = {
  isAutoFetch?: boolean
  key?: string
  page?: number
  filter?: DataTableFilterProps
  limit?: number
}

type DataTableQueryKey = [
  string,
  { page: number; limit: number; filter?: DataTableFilterProps },
]

const fetchDataTable = async (args: QueryFunctionContext<DataTableQueryKey>) => {
  const [, { page, limit, filter }] = args.queryKey

  const result = await axios<DataTableResponse<DataTypeProps>>({
    method: 'GET',
    url: apiRouters.paymentMethods.index,
    params: {
      ...filter,
      search: filter?.search?.trim() || undefined,
      page,
      limit,
    },
  })

  return result
}

const useDataTable = (args?: UseDataTableProps) => {
  const {
    key = queryKeys.paymentMethods.index,
    page = 1,
    filter,
    limit = DEFAULT_LIMIT,
    isAutoFetch,
  } = args || {}

  const dataTable = useQuery<
    DataTableResponse<DataTypeProps>,
    ErrorResponse<AxiosError>,
    DataTableResponse<DataTypeProps>,
    DataTableQueryKey
  >({
    queryKey: [key, { page, limit, filter }],
    enabled: isAutoFetch,
    queryFn: fetchDataTable,
  })

  return {
    limit,
    refetch: dataTable.refetch,
    data: dataTable.data?.list,
    pagination: dataTable.data?.meta.pagination,
    error: dataTable.error,
    isLoading: dataTable.isLoading,
  }
}

export default useDataTable
export { useDataTable as usePaymentMethodsDataTable }
```

**Usage in content:**

```tsx
const { data: methods, isLoading, refetch, pagination, limit } = usePaymentMethodsDataTable({
  filter: { search: currentSearch || undefined },
  page: currentPage,
  isAutoFetch: true,
})
```

---

### `useGetOne` — Fetch Single Item

```typescript
// hooks/transactions/use-payment-methods/use-get-one.ts
import { useQuery } from '@tanstack/react-query'
import { axios } from '$/services/axios'
import { apiRouters, queryKeys } from '$/constants'
import { ErrorResponse } from '$/types/generals'
import { AxiosError } from 'axios'
import type { PaymentMethodResponseProps as ResponseProps } from '@vibecoding-starter/types/payment-method-response'
import { pathVariable } from '$/utils/path-variable'

type UseGetOneProps = { id: string }

const getOne = async (id: string) => {
  return axios<ResponseProps>({
    method: 'GET',
    url: pathVariable(apiRouters.paymentMethods.show, { id }),
  })
}

const useGetOne = ({ id }: UseGetOneProps) => {
  return useQuery<ResponseProps, ErrorResponse<AxiosError>, ResponseProps, [string, string]>({
    queryKey: [queryKeys.paymentMethods.get, id],
    queryFn: () => getOne(id),
    enabled: !!id,
  })
}

export default useGetOne
export { useGetOne as usePaymentMethodsGetOne }
```

---

### `useInsertOne` — Create Mutation

```typescript
// hooks/transactions/use-payment-methods/use-insert-one.ts
import { useMutation } from '@tanstack/react-query'
import { axios } from '$/services/axios'
import { apiRouters, queryKeys } from '$/constants'
import { ErrorResponse } from '$/types/generals'
import { AxiosError } from 'axios'
import type { PaymentMethodSchemaProps as PayloadProps } from '$/schemas/payment-method'
import type { PaymentMethodResponseProps as ResponseProps } from '@vibecoding-starter/types/payment-method-response'

const insertOne = async (payload: PayloadProps) => {
  return axios<ResponseProps>({
    method: 'POST',
    url: apiRouters.paymentMethods.insert,
    data: payload,
  })
}

const useInsertOne = () => {
  return useMutation<ResponseProps, ErrorResponse<AxiosError>, PayloadProps, unknown>({
    mutationKey: [queryKeys.paymentMethods.insert],
    mutationFn: insertOne,
  })
}

export default useInsertOne
export { useInsertOne as usePaymentMethodsInsertOne }
```

---

### `useUpdateOne` — Update Mutation

```typescript
// hooks/transactions/use-payment-methods/use-update-one.ts
import { useMutation } from '@tanstack/react-query'
import { axios } from '$/services/axios'
import { apiRouters, queryKeys } from '$/constants'
import { ErrorResponse } from '$/types/generals'
import { AxiosError } from 'axios'
import type { PaymentMethodSchemaProps as PayloadProps } from '$/schemas/payment-method'
import type { PaymentMethodResponseProps as ResponseProps } from '@vibecoding-starter/types/payment-method-response'
import { pathVariable } from '$/utils/path-variable'

type UpdateParamsProps = { id: string; payload: PayloadProps }

const updateOne = async ({ id, payload }: UpdateParamsProps) => {
  return axios<ResponseProps>({
    method: 'PUT',
    url: pathVariable(apiRouters.paymentMethods.update, { id }),
    data: payload,
  })
}

const useUpdateOne = () => {
  return useMutation<ResponseProps, ErrorResponse<AxiosError>, UpdateParamsProps, UpdateParamsProps>({
    mutationKey: [queryKeys.paymentMethods.update],
    mutationFn: updateOne,
  })
}

export default useUpdateOne
export { useUpdateOne as usePaymentMethodsUpdateOne }
```

---

### `useDeleteOne` — Delete Mutation

```typescript
// hooks/transactions/use-payment-methods/use-delete-one.ts
import { useMutation } from '@tanstack/react-query'
import { axios } from '$/services/axios'
import { apiRouters, queryKeys } from '$/constants'
import { ErrorResponse } from '$/types/generals'
import { AxiosError } from 'axios'
import { pathVariable } from '$/utils/path-variable'

const deleteOne = async (id: string) => {
  return axios<unknown>({
    method: 'DELETE',
    url: pathVariable(apiRouters.paymentMethods.delete, { id }),
  })
}

const useDeleteOne = () => {
  return useMutation<unknown, ErrorResponse<AxiosError>, string, unknown>({
    mutationKey: [queryKeys.paymentMethods.delete],
    mutationFn: deleteOne,
  })
}

export default useDeleteOne
export { useDeleteOne as usePaymentMethodsDeleteOne }
```

---

### `index.ts` — Re-export

```typescript
// hooks/transactions/use-payment-methods/index.ts
export { default as usePaymentMethodsDataTable } from './use-data-table'
export { default as usePaymentMethodsGetOne } from './use-get-one'
export { default as usePaymentMethodsInsertOne } from './use-insert-one'
export { default as usePaymentMethodsUpdateOne } from './use-update-one'
export { default as usePaymentMethodsDeleteOne } from './use-delete-one'
```

---

### `useQueryParam` — Utility Hook

```typescript
// hooks/utility/use-query-param/use-query-param.ts
'use client'

import { useCallback } from 'react'
import { usePathname, useRouter, useSearchParams } from 'next/navigation'

const useQueryParam = () => {
  const searchParams = useSearchParams()
  const pathname = usePathname()
  const router = useRouter()
  const searchParamsString = searchParams.toString()

  const setQueryParams = useCallback(
    (valuesToUpdate: Record<string, unknown>, options?: { method?: 'push' | 'replace' }) => {
      const params = new URLSearchParams(searchParamsString)

      Object.entries(valuesToUpdate).forEach(([key, value]) => {
        const resolvedValue = typeof value === 'string' ? value.trim() : value
        if (resolvedValue !== undefined && String(resolvedValue).length > 0) {
          params.set(key, String(resolvedValue))
        } else {
          params.delete(key)
        }
      })

      const queryString = params.toString()
      const updatedPath = queryString ? `${pathname}?${queryString}` : pathname

      if (options?.method === 'replace') {
        router.replace(updatedPath)
        return
      }
      router.push(updatedPath)
    },
    [pathname, router, searchParamsString],
  )

  return { setQueryParams }
}

export default useQueryParam
export { useQueryParam }
```

---

## Usage in Content File

```tsx
// Insert + update use mutateAsync (can await, can use callbacks)
const { mutateAsync: insertMutateAsync, isPending: isInsertPending } = useFeatureInsertOne()
const { mutateAsync: updateMutateAsync, isPending: isUpdatePending } = useFeatureUpdateOne()

// Delete uses mutate (because it's used in SweetAlert preConfirm Promise)
const { mutate: deleteMutate } = useFeatureDeleteOne()

// How to submit
insertMutateAsync(data, {
  onSuccess: () => toast.success('Created successfully'),
  onError: ({ message }) => toast.error(message || 'Failed'),
  onSettled: () => { handleCloseDialog(); refetch() },
})
```

---

## Additional Rules

- One file = one hook
- File name = kebab-case of hook name
- Default export + named export alias for each hook file
- Files must end with newline
