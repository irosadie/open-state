# Guide: API Use Case (`apps/api/src/domain/use-cases/`)

## Folder Contract

✅ Allowed:
- One business logic operation per file
- Throw `DomainError` for expected errors (not found, conflict, unauthorized)
- Call repository interface (not direct implementation)
- Validate business rules: check duplicates, domain authorization, etc.

❌ Forbidden:
- Access Prisma or database directly
- Import from `hono` or HTTP library
- Throw `HTTPException` — use `DomainError`
- Orchestrate multiple domain operations — that's the service's job

---

## Conventions

### Structure

```
domain/use-cases/
├── create-user.ts
├── get-user-by-id.ts
├── update-user.ts
└── delete-user.ts
```

### Use Case Pattern

```typescript
// domain/use-cases/create-user.ts
import { DomainError } from '@/domain/errors'
import type { IUserRepository } from '@/domain/repositories/IUserRepository'
import type { User } from '@/domain/entities/User'

type CreateUserInput = {
  name: string
  email: string
  role: string
}

export async function createUser(
  repository: IUserRepository,
  input: CreateUserInput,
): Promise<User> {
  const existing = await repository.findByEmail(input.email)
  if (existing) {
    throw new DomainError('CONFLICT', 'Email already registered')
  }

  return repository.create(input)
}
```

### Use Case with Auth Check

```typescript
// domain/use-cases/delete-user.ts
export async function deleteUser(
  repository: IUserRepository,
  id: string,
  requesterId: string,
): Promise<void> {
  const user = await repository.findById(id)
  if (!user) {
    throw new DomainError('NOT_FOUND', 'User not found')
  }

  if (user.id !== requesterId) {
    throw new DomainError('FORBIDDEN', 'Cannot delete other user')
  }

  await repository.delete(id)
}
```

### Naming

- File name: `{verb}-{domain}.ts` — kebab-case, starts with verb
- Example: `create-user.ts`, `get-order-by-id.ts`, `cancel-shipment.ts`
- Export: named function (not class)

---

## Additional Rules

- First parameter is always repository — dependency injection via parameter
- Input type is defined locally in the use case file
- File must end with newline
