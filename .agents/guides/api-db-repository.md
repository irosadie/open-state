# Guide: API Prisma Repository (`apps/api/src/infrastructure/database/`)

## Folder Contract

✅ Allowed:
- Implement repository interface from `domain/repositories/`
- Access Prisma client
- Map Prisma model to domain Entity

❌ Forbidden:
- Business logic
- Return raw Prisma model — always map to Entity
- HTTP concerns

---

## Conventions

### Prisma Repository Pattern

```typescript
// infrastructure/database/PrismaUserRepository.ts
import { prisma } from '@/infrastructure/database/prisma'
import type { IUserRepository, UserListFilter, UserListResult } from '@/domain/repositories/IUserRepository'
import type { User } from '@/domain/entities/User'

export class PrismaUserRepository implements IUserRepository {
  async findById(id: string): Promise<User | null> {
    const row = await prisma.user.findUnique({ where: { id } })
    return row ? toEntity(row) : null
  }

  async findAll(filter: UserListFilter): Promise<UserListResult> {
    const where = {
      ...(filter.search && {
        OR: [
          { name: { contains: filter.search, mode: 'insensitive' as const } },
          { email: { contains: filter.search, mode: 'insensitive' as const } },
        ],
      }),
      ...(filter.isActive !== undefined && { isActive: filter.isActive }),
    }

    const [data, total] = await Promise.all([
      prisma.user.findMany({
        where,
        skip: (filter.page - 1) * filter.limit,
        take: filter.limit,
        orderBy: { createdAt: 'desc' },
      }),
      prisma.user.count({ where }),
    ])

    return { data: data.map(toEntity), total }
  }

  async create(input: Omit<User, 'id' | 'createdAt' | 'updatedAt'>): Promise<User> {
    const row = await prisma.user.create({ data: input })
    return toEntity(row)
  }

  async update(id: string, input: Partial<Pick<User, 'name' | 'role' | 'isActive'>>): Promise<User> {
    const row = await prisma.user.update({ where: { id }, data: input })
    return toEntity(row)
  }

  async delete(id: string): Promise<void> {
    await prisma.user.delete({ where: { id } })
  }
}

function toEntity(row: any): User {
  return {
    id: row.id,
    name: row.name,
    email: row.email,
    role: row.role,
    isActive: row.isActive,
    createdAt: row.createdAt,
    updatedAt: row.updatedAt,
  }
}
```

### Prisma Client Singleton

```typescript
// infrastructure/database/prisma.ts
import { PrismaClient } from '@prisma/client'

export const prisma = new PrismaClient()
```

### Naming

- File name: `Prisma{Domain}Repository.ts` — PascalCase with `Prisma` prefix
- Example: `PrismaUserRepository.ts`, `PrismaOrderRepository.ts`

---

## Additional Rules

- `toEntity` mapper is written as a local function below the class — no need to export
- Use `Promise.all` for concurrent count + data queries (pagination)
- Build `where` object conditionally — avoid `undefined` values in Prisma
- File must end with newline
