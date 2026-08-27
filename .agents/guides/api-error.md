# Guide: API Error (`apps/api/src/domain/errors/`)

## Folder Contract

✅ Allowed:
- Define `DomainError` class and error code enum
- Export from a single `index.ts`

❌ Forbidden:
- Import from `hono` or HTTP library
- HTTP error handling here — that goes in errorHandler middleware

---

## Conventions

### DomainError

```typescript
// domain/errors/DomainError.ts
export type DomainErrorCode =
  | 'NOT_FOUND'
  | 'UNAUTHORIZED'
  | 'FORBIDDEN'
  | 'CONFLICT'
  | 'VALIDATION'
  | 'INTERNAL'

export class DomainError extends Error {
  constructor(
    public readonly code: DomainErrorCode,
    message: string,
  ) {
    super(message)
    this.name = 'DomainError'
  }
}
```

### Error Handler Middleware

```typescript
// interfaces/http/middlewares/error-handler.ts
import type { Context } from 'hono'
import { DomainError } from '@/domain/errors/DomainError'

const statusMap: Record<string, number> = {
  NOT_FOUND: 404,
  UNAUTHORIZED: 401,
  FORBIDDEN: 403,
  CONFLICT: 409,
  VALIDATION: 422,
  INTERNAL: 500,
}

export async function errorHandler(err: Error, c: Context) {
  if (err instanceof DomainError) {
    return c.json(
      { success: false, message: err.message, code: err.code },
      (statusMap[err.code] ?? 500) as any,
    )
  }
  console.error(err)
  return c.json({ success: false, message: 'Internal server error' }, 500)
}
```

### Usage in Use Case

```typescript
import { DomainError } from '@/domain/errors'

throw new DomainError('NOT_FOUND', 'User not found')
throw new DomainError('CONFLICT', 'Email already exists')
throw new DomainError('FORBIDDEN', 'Insufficient permissions')
```

---

## Additional Rules

- Always import `DomainError` from `@/domain/errors` (index re-export)
- Don't throw `HTTPException` from use case — use `DomainError`
- File must end with newline
