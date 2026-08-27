# Guide: API Entity (`apps/api/src/domain/entities/`)

## Folder Contract

✅ Allowed:
- Plain TypeScript `type` or `class` that represents a domain model
- Domain-specific fields — not database columns
- Pure domain methods (calculations, internal validation)

❌ Forbidden:
- Import Prisma types or Prisma client
- HTTP or database dependencies
- Business logic that changes per use case — put that in use cases

---

## Conventions

### Entity as Type

```typescript
// domain/entities/User.ts
export type User = {
  id: string
  name: string
  email: string
  role: string
  isActive: boolean
  createdAt: Date
  updatedAt: Date
}
```

### Entity as Class (if domain methods exist)

```typescript
// domain/entities/Order.ts
export class Order {
  constructor(
    public readonly id: string,
    public readonly total: number,
    public readonly status: string,
    public readonly createdAt: Date,
  ) {}

  isCompleted(): boolean {
    return this.status === 'COMPLETED'
  }

  canBeCancelled(): boolean {
    return ['PENDING', 'PROCESSING'].includes(this.status)
  }
}
```

### Naming

- File name: `{Domain}.ts` — PascalCase without suffix
- Type/class name: same as file name
- Example: `User.ts`, `Order.ts`, `ProductCategory.ts`

---

## Additional Rules

- Entity is the source of truth for the domain — not a mirror of Prisma schema
- If Prisma model has field `is_active` (`snake_case`), Entity uses `isActive` (`camelCase`)
- Use primitive types — not Prisma scalar types
- File must end with newline
