# Guide: Web Type (`apps/web/types/`)

## Folder Contract

✅ Allowed:
- TypeScript `type` and `interface` that are FE-specific
- Types for local state, complex component props, route params
- `DataTableResponse<T>` and other general types
- Re-export `ErrorResponse<T>` from `@vibecoding-starter/types` if needed
- Re-export from `index.ts` per subfolder

❌ Forbidden:
- Use `any`
- Put API response types here — those belong in `packages/types/`
- Put form payload types here — those belong in `packages/schemas/` via `z.infer<>`
- Business logic or runtime code

---

## Conventions

### Structure

```
types/
└── generals/
    ├── data-table.ts   → DataTableResponse<T>, pagination types
    ├── next-auth.d.ts  → NextAuth module augmentation
    └── index.ts        → re-export all
```

### `DataTableResponse<T>`

```typescript
// types/generals/data-table.ts
export type PaginationMeta = {
  total: number
  currentPage: number
  perPage: number
  lastPage: number
}

export type CursorMeta = {
  next: string | null
  prev: string | null
} | null

export type DataTableMeta = {
  pagination: PaginationMeta
  cursor: CursorMeta
}

export type DataTableResponse<T> = {
  list: T[]
  meta: DataTableMeta
}
```

### Re-export

```typescript
// types/generals/index.ts
export type { ErrorResponse } from '@vibecoding-starter/types'
export * from './data-table'
```

---

## Additional Rules

- Use `type` for unions, intersections, and aliases
- Use `interface` for object shapes that might be extended
- No need for `I` prefix for interfaces
- Files must end with newline
