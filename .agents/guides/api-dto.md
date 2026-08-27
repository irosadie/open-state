# Guide: API DTO (`apps/api/src/application/dtos/`)

## Folder Contract

✅ Allowed:
- TypeScript `type` or `interface` for output sent to client
- Subset or transform from Entity (serializable: string, number, boolean — not Date object)

❌ Forbidden:
- Zod schema (that goes in `validators/`)
- Business logic or methods
- Prisma types
- Use `any`

---

## Conventions

### DTO Pattern

```typescript
// application/dtos/UserDto.ts
export type UserDto = {
  id: string
  name: string
  email: string
  role: string
  isActive: boolean
  createdAt: string   // ISO string, not Date
}

export type UserListDto = {
  data: UserDto[]
  meta: {
    total: number
    page: number
    limit: number
  }
}
```

### Naming

- File name: `{Domain}Dto.ts` — PascalCase with `Dto` suffix
- Example: `UserDto.ts`, `OrderDto.ts`
- Type name: `{Domain}Dto`, `{Domain}ListDto`, `{Domain}DetailDto`

---

## Additional Rules

- `Date` → `string` (ISO format) in DTO for JSON serialization
- Nullable fields use `type | null`, not `type | undefined` for responses
- File must end with newline
