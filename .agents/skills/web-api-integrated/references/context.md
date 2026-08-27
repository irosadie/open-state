# Context: Web API Integrated

## Target Folder

```
packages/
├── schemas/              → Zod schema + constants + labels + helper + type
└── types/                → flat API response TypeScript types

apps/web/
├── hooks/transactions/
│   └── use-{domain}/     → one folder per domain
│       ├── use-data-table.ts
│       ├── use-get-one.ts
│       ├── use-insert-one.ts
│       ├── use-update-one.ts
│       ├── use-delete-one.ts
│       └── index.ts
└── constants/
    ├── api-routers.ts    → `:id` path variables
    └── query-keys.ts     → flat string per operation
```

## Real Code Examples

See: `.agents/examples/web-api-integrated/hooks/use-examples/`

## `useDataTable` Pattern (react-query hook — not TanStack Table)

```typescript
const useDataTable = (args?: UseDataTableProps) => {
  const { key = queryKeys.examples.index, page = 1, filter, limit = 10, isAutoFetch } = args || {}

  const dataTable = useQuery<DataTableResponse<DataTypeProps>, ErrorResponse<AxiosError>, ...>({
    queryKey: [key, { page, limit, filter }],
    enabled: isAutoFetch,
    queryFn: async (args) => {
      const [, { page, limit, filter }] = args.queryKey
      return axios({ method: 'GET', url: apiRouters.examples.index, params: { ...filter, page, limit } })
    },
  })

  return {
    limit,
    data: dataTable.data?.list,
    pagination: dataTable.data?.meta.pagination,
    isLoading: dataTable.isLoading,
    refetch: dataTable.refetch,
  }
}
```

## Hook Export Pattern

```typescript
// Each hook file: default export + named export alias
export default useDataTable
export { useDataTable as useExamplesDataTable }

// index.ts: re-export everything with domain-prefixed names
export { default as useExamplesDataTable } from './use-data-table'
export { default as useExamplesGetOne } from './use-get-one'
```

## Constants Pattern

```typescript
// constants/api-routers.ts — use :id path variable
export const apiRouters = {
  examples: {
    index: '/examples',
    show: '/examples/:id',
    insert: '/examples',
    update: '/examples/:id',
    delete: '/examples/:id',
  },
}

// constants/query-keys.ts — flat string per operation
export const queryKeys = {
  examples: {
    index: 'examplesIndex',
    get: 'examplesGet',
    insert: 'examplesInsert',
    update: 'examplesUpdate',
    delete: 'examplesDelete',
  },
}
```

## Schema Pattern

```typescript
// packages/schemas/example.ts
import { z } from 'zod'

export const exampleTypes = ['EXAMPLE_TYPE_A', 'EXAMPLE_TYPE_B'] as const

export const exampleLabels = [
  { label: 'Type A', value: 'EXAMPLE_TYPE_A' },
  { label: 'Type B', value: 'EXAMPLE_TYPE_B' },
]

export const getExampleLabel = (value: typeof exampleTypes[number]) => {
  return exampleLabels.find((l) => l.value === value)?.label ?? value
}

export const exampleSchema = z.object({
  name: z.string().min(1, 'Name is required'),
  code: z.string().min(1, 'Code is required'),
  type: z.enum(exampleTypes),
  isActive: z.boolean().optional(),
})

export type ExampleSchemaProps = z.infer<typeof exampleSchema>
```

## Response Type Pattern

```typescript
// packages/types/example-response.ts
export type ExampleResponseProps = {
  id: string
  name: string
  code: string
  type: 'EXAMPLE_TYPE_A' | 'EXAMPLE_TYPE_B'
  isActive: boolean
  createdAt: string
  updatedAt: string
}
```
