# Guide: API Service (`apps/api/src/application/services/`)

## Folder Contract

✅ Allowed:
- Orchestrate one or more use cases
- Transform Entity to DTO before returning to controller
- Inject and call repository interface or use case functions
- Coordinate side effects (send email via worker queue, etc.)

❌ Forbidden:
- Pure business logic — that goes in use cases
- Access Prisma directly — must go through repository interface
- HTTP concerns (status code, header, `c.json`)
- `try/catch` for domain errors — let them bubble

---

## Conventions

### Service Pattern

```typescript
// application/services/UserService.ts
import { createUser } from '@/domain/use-cases/create-user'
import { getUserById } from '@/domain/use-cases/get-user-by-id'
import { listUsers } from '@/domain/use-cases/list-users'
import { updateUser } from '@/domain/use-cases/update-user'
import { deleteUser } from '@/domain/use-cases/delete-user'
import type { IUserRepository } from '@/domain/repositories/IUserRepository'
import type { UserDto, UserListDto } from '@/application/dtos/UserDto'
import type { User } from '@/domain/entities/User'

export class UserService {
  constructor(private readonly repository: IUserRepository) {}

  async list(params: { page: number; limit: number; search?: string }): Promise<UserListDto> {
    const result = await listUsers(this.repository, params)
    return {
      data: result.data.map(toUserDto),
      meta: { total: result.total, page: params.page, limit: params.limit },
    }
  }

  async getById(id: string): Promise<UserDto> {
    const user = await getUserById(this.repository, id)
    return toUserDto(user)
  }

  async create(payload: CreateUserPayload): Promise<UserDto> {
    const user = await createUser(this.repository, payload)
    return toUserDto(user)
  }

  async update(id: string, payload: UpdateUserPayload): Promise<UserDto> {
    const user = await updateUser(this.repository, id, payload)
    return toUserDto(user)
  }

  async delete(id: string): Promise<void> {
    await deleteUser(this.repository, id)
  }
}

function toUserDto(user: User): UserDto {
  return {
    id: user.id,
    name: user.name,
    email: user.email,
    role: user.role,
    isActive: user.isActive,
    createdAt: user.createdAt.toISOString(),
  }
}
```

### Naming

- File name: `{Domain}Service.ts` — PascalCase with `Service` suffix
- Example: `UserService.ts`, `OrderService.ts`

---

## Additional Rules

- `toDto` mapper function is written below the class, no need to export
- Constructor only accepts repository interface — not concrete implementation
- Service has no knowledge of HTTP — don't import `hono`
- File must end with newline
