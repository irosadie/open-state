# Guide: API Validator (`apps/api/src/application/validators/`)

## Folder Contract

✅ Allowed:
- Zod schema for request validation (body, query, params)
- Export inferred types from schema
- Group by domain in a single file

❌ Forbidden:
- Business logic or database queries
- Import from `domain/` or `infrastructure/`
- Use `any`

---

## Conventions

### Validator Pattern

```typescript
// application/validators/user.schemas.ts
import { z } from 'zod'

export const createUserSchema = z.object({
  name: z.string().min(1, 'Name is required').max(100),
  email: z.string().email('Invalid email format'),
  role: z.enum(['ADMIN', 'USER', 'MANAGER']),
})

export const updateUserSchema = z.object({
  name: z.string().min(1).max(100).optional(),
  role: z.enum(['ADMIN', 'USER', 'MANAGER']).optional(),
  isActive: z.boolean().optional(),
})

export const listUserQuerySchema = z.object({
  page: z.coerce.number().int().positive().default(1),
  limit: z.coerce.number().int().positive().max(100).default(10),
  search: z.string().optional(),
  role: z.enum(['ADMIN', 'USER', 'MANAGER']).optional(),
})

export type CreateUserPayload = z.infer<typeof createUserSchema>
export type UpdateUserPayload = z.infer<typeof updateUserSchema>
export type ListUserQuery = z.infer<typeof listUserQuerySchema>
```

### Naming

- File name: `{domain}.schemas.ts` — kebab-case with `.schemas` suffix
- Example: `user.schemas.ts`, `order.schemas.ts`
- Schema name: `{action}{Domain}Schema` — camelCase

---

## Additional Rules

- Use `z.coerce.number()` for query params (always string from HTTP)
- Export inferred type alongside schema in the same file
- For enums, use `z.enum([...])` instead of `z.string()` if values are constrained
- File must end with newline
