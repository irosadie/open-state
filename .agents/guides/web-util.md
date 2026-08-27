# Guide: Web Util (`apps/web/utils/`)

## Folder Contract

✅ Allowed:
- Pure functions without side effects
- Format, transform, parse helpers that are FE-specific
- Helpers tied to React/browser API belong here (not in `packages/utils/`)

❌ Forbidden:
- Axios or fetch calls
- Import React hooks here — move to `hooks/`
- State management
- Business logic that changes according to requirements

---

## Conventions

### `debounce` — Utility with `.cancel()`

```typescript
// utils/debounce.ts
type DebouncedFn<T extends (...args: unknown[]) => unknown> = {
  (...args: Parameters<T>): void
  cancel: () => void
}

export function debounce<T extends (...args: unknown[]) => unknown>(
  fn: T,
  delay: number,
): DebouncedFn<T> {
  let timer: ReturnType<typeof setTimeout> | null = null

  const debounced = (...args: Parameters<T>) => {
    if (timer) clearTimeout(timer)
    timer = setTimeout(() => fn(...args), delay)
  }

  debounced.cancel = () => {
    if (timer) {
      clearTimeout(timer)
      timer = null
    }
  }

  return debounced
}
```

**Usage with `useMemo`:**

```tsx
const debouncedSearch = useMemo(
  () => debounce((value: string) => {
    setQueryParams({ q: value || undefined, page: 1 })
  }, 250),
  [setQueryParams],
)

// Cleanup on unmount
useEffect(() => {
  return () => debouncedSearch.cancel()
}, [debouncedSearch])
```

### `pathVariable` — URL Path Builder

```typescript
// utils/path-variable.ts
export function pathVariable<T extends Record<string, string>>(
  url: string,
  params: T,
): string {
  return Object.entries(params).reduce(
    (acc, [key, value]) => acc.replace(`:${key}`, value),
    url,
  )
}
```

**Usage:**

```typescript
pathVariable('/payment-methods/:id', { id: '123' })
// → '/payment-methods/123'

pathVariable('/vendors/:vendorId/drivers/:id', { vendorId: 'v1', id: 'd1' })
// → '/vendors/v1/drivers/d1'
```

### Format Currency

```typescript
// utils/format-currency.ts
export function formatCurrency(amount: number, currency = 'IDR'): string {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency,
    minimumFractionDigits: 0,
  }).format(amount)
}
```

---

## Additional Rules

- One function per file, or group related functions in one file
- No need for classes — use plain functions
- Files must end with newline
