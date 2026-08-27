# Guide: Web Test (`apps/web/`)

## Test Runner

Vitest + React Testing Library. Config in `apps/web/vitest.config.ts`.

## Contract

✅ Allowed:
- Unit tests for hooks (mock react-query and axios)
- Unit tests for utils (pure functions)
- Mock axios with `vi.mock`

❌ Forbidden:
- Test implementation details (internal state, private methods)
- Use `any` in test code
- Skip tests for important logic

---

## Conventions

### Test File Structure

```
hooks/transactions/use-payment-methods/
├── use-data-table.ts
├── use-data-table.test.ts   → next to the file being tested
└── ...

utils/
├── format-currency.ts
└── format-currency.test.ts
```

### Test Hook dengan react-query

```typescript
// hooks/transactions/use-payment-methods/use-data-table.test.ts
import { renderHook, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import useDataTable from './use-data-table'
import * as axiosModule from '$/services/axios'

vi.mock('$/services/axios')

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  )
}

describe('useDataTable', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('fetches payment methods list successfully', async () => {
    vi.mocked(axiosModule.axios).mockResolvedValue({
      list: [{ id: '1', name: 'BCA' }],
      meta: { pagination: { total: 1, currentPage: 1, perPage: 10, lastPage: 1 }, cursor: null },
    })

    const { result } = renderHook(
      () => useDataTable({ page: 1, isAutoFetch: true }),
      { wrapper: createWrapper() }
    )

    await waitFor(() => expect(result.current.isLoading).toBe(false))
    expect(result.current.data).toHaveLength(1)
  })
})
```

### Test Pure Util

```typescript
// utils/format-currency.test.ts
import { describe, it, expect } from 'vitest'
import { formatCurrency } from './format-currency'

describe('formatCurrency', () => {
  it('formats IDR correctly', () => {
    expect(formatCurrency(50000)).toBe('Rp 50.000')
  })
})
```

---

## Additional Rules

- Test files next to the files being tested (not in a separate `__tests__` folder)
- Use `describe` for groups, `it` for individual cases
- Run: `bun run test` in `apps/web/`
