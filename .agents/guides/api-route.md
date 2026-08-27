# Guide: API Route (`apps/api/src/interfaces/http/routes/`)

## Folder Contract

✅ Allowed:
- Define HTTP method + path
- Validate request with Zod schema from `validators/`
- Attach middleware per route (auth, rate-limit)
- Delegate to controller handler

❌ Forbidden:
- Business logic
- Call use case or repository directly
- Format response directly

---

## Conventions

### Route Pattern

```typescript
// interfaces/http/routes/userRoutes.ts
import { Hono } from 'hono'
import { zValidator } from '@hono/zod-validator'
import { UserController } from '@/interfaces/http/controllers/UserController'
import { UserService } from '@/application/services/UserService'
import { PrismaUserRepository } from '@/infrastructure/database/PrismaUserRepository'
import { createUserSchema, updateUserSchema } from '@/application/validators/user.schemas'
import { authMiddleware } from '@/interfaces/http/middlewares/auth'

const userRepository = new PrismaUserRepository()
const userService = new UserService(userRepository)
const userController = new UserController(userService)

export const userRoutes = new Hono()
  .use(authMiddleware)
  .get('/', userController.list)
  .get('/:id', userController.getById)
  .post('/', zValidator('json', createUserSchema), userController.create)
  .put('/:id', zValidator('json', updateUserSchema), userController.update)
  .delete('/:id', userController.delete)
```

### Registration in `create-app.ts`

```typescript
// interfaces/http/create-app.ts
import { userRoutes } from './routes/userRoutes'

app.route('/users', userRoutes)
```

### Naming

- File name: `{domain}Routes.ts` — camelCase with `Routes` suffix
- Example: `userRoutes.ts`, `orderRoutes.ts`, `authRoutes.ts`

---

## Additional Rules

- Instantiate dependencies (repository, service, controller) inside the route file
- Use `zValidator` from `@hono/zod-validator` for automatic validation
- File must end with newline
