# Guide: API Controller (`apps/api/src/interfaces/http/controllers/`)

## Folder Contract

✅ Allowed:
- Parse `c.req` (body, params, query)
- Call service method
- Format and return HTTP response via `c.json(...)`

❌ Forbidden:
- Business logic — that goes in service and use case
- Call use case or repository directly — must go through service
- `try/catch` for domain errors — let them bubble to errorHandler

---

## Conventions

### Controller Pattern

```typescript
// interfaces/http/controllers/UserController.ts
import type { Context } from 'hono'
import type { UserService } from '@/application/services/UserService'

export class UserController {
  constructor(private readonly userService: UserService) {}

  list = async (c: Context) => {
    const query = c.req.query()
    const result = await this.userService.list({
      page: Number(query.page ?? 1),
      limit: Number(query.limit ?? 10),
      search: query.search,
    })
    return c.json(result)
  }

  getById = async (c: Context) => {
    const { id } = c.req.param()
    const result = await this.userService.getById(id)
    return c.json(result)
  }

  create = async (c: Context) => {
    const body = await c.req.json()
    const result = await this.userService.create(body)
    return c.json(result, 201)
  }

  update = async (c: Context) => {
    const { id } = c.req.param()
    const body = await c.req.json()
    const result = await this.userService.update(id, body)
    return c.json(result)
  }

  delete = async (c: Context) => {
    const { id } = c.req.param()
    await this.userService.delete(id)
    return c.json({ message: 'Deleted successfully' })
  }
}
```

### Naming

- File name: `{Domain}Controller.ts` — PascalCase with `Controller` suffix
- Example: `UserController.ts`, `OrderController.ts`
- Method = action name: `list`, `getById`, `create`, `update`, `delete`

---

## Additional Rules

- Arrow function as method for safe `this` binding
- All data transformation is in service, all business logic is in use case
- File must end with newline
