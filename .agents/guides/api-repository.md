# Guide: API Repository Interface (`apps/api/src/domain/repositories/`)

## Folder Contract

✅ Allowed:
- Define interface or abstract class for repository
- Method signature uses Entity types
- Optional filter/pagination types local to each interface

❌ Forbidden:
- Concrete implementation — that goes in `infrastructure/database/`
- Import Prisma
- Business logic

---

## Conventions

### Interface Pattern

```typescript
// domain/repositories/IUserRepository.ts
import type { User } from '@/domain/entities/User'

export type UserListFilter = {
  search?: string
  role?: string
  isActive?: boolean
  page: number
  limit: number
}

export type UserListResult = {
  data: User[]
  total: number
}

export interface IUserRepository {
  findById(id: string): Promise<User | null>
  findByEmail(email: string): Promise<User | null>
  findAll(filter: UserListFilter): Promise<UserListResult>
  create(input: Omit<User, 'id' | 'createdAt' | 'updatedAt'>): Promise<User>
  update(id: string, input: Partial<Pick<User, 'name' | 'role' | 'isActive'>>): Promise<User>
  delete(id: string): Promise<void>
}
```

### Naming

- File name: `I{Domain}Repository.ts` — `I` prefix for interface
- Example: `IUserRepository.ts`, `IOrderRepository.ts`
- Interface name = file name

---

## Additional Rules

- Return type is always Entity, not Prisma model
- Method `findById` returns `T | null` (nullable)
- Method `delete` returns `void`
- File must end with newline
