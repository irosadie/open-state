# Guide: Shared Schema (`packages/schemas/`)

## Folder Contract

✅ Allowed:
- Type constants array + labels array + `get{Type}Label()` helper
- Zod schema for form payload, request body, job data
- `z.infer<>` types from schema
- Shared enum/constant values
- Export from `index.ts`

❌ Forbidden:
- Import FE-specific libraries (React) or BE-specific libraries (Hono, Prisma)
- Business logic, side effects, API calls
- Response types — those belong in `packages/types/`
- Use `any`

---

## Conventions

### Schema File — Complete Pattern

```typescript
// packages/schemas/payment-method.ts
import { z } from 'zod'

// 1. Type constants (array of valid values)
export const paymentMethodTypes = [
  'BANK_TRANSFER',
  'E_WALLET',
  'CREDIT_CARD',
  'QRIS',
  'COD',
] as const

// 2. Labels array (for dropdown/select)
export const paymentMethodLabels = [
  { label: 'Bank Transfer', value: 'BANK_TRANSFER' },
  { label: 'E-Wallet', value: 'E_WALLET' },
  { label: 'Credit Card', value: 'CREDIT_CARD' },
  { label: 'QRIS', value: 'QRIS' },
  { label: 'Cash on Delivery', value: 'COD' },
]

// 3. Helper function for display label
export const getPaymentMethodLabel = (value: typeof paymentMethodTypes[number]) => {
  const method = paymentMethodLabels.find((m) => m.value === value)
  return method ? method.label : value
}

// 4. Zod schema
export const paymentMethodSchema = z.object({
  name: z.string().min(1, 'Name is required'),
  code: z.string().min(1, 'Code is required'),
  type: z.enum(paymentMethodTypes),
  merchantId: z.string().optional(),
  bankName: z.string().optional(),
  isActive: z.boolean().optional(),
})

// 5. Type alias
export type PaymentMethodSchemaProps = z.infer<typeof paymentMethodSchema>
```

### Simple Schema (without labels)

```typescript
// packages/schemas/user.ts
import { z } from 'zod'

export const createUserSchema = z.object({
  name: z.string().min(1, 'Name is required').max(100),
  email: z.string().email('Invalid email'),
  role: z.enum(['ADMIN', 'USER']),
})

export const updateUserSchema = z.object({
  name: z.string().min(1).max(100).optional(),
  isActive: z.boolean().optional(),
})

export type CreateUserPayload = z.infer<typeof createUserSchema>
export type UpdateUserPayload = z.infer<typeof updateUserSchema>
```

### Shared Enum Schema

```typescript
// packages/schemas/notification-channel.ts
import { z } from 'zod'

export const NOTIFICATION_CHANNELS = ['EMAIL', 'SMS', 'PUSH'] as const

export const notificationChannelSchema = z.enum(NOTIFICATION_CHANNELS)

export const notificationPreferenceSchema = z.object({
  channel: notificationChannelSchema,
  isEnabled: z.boolean(),
})

export type NotificationPreference = z.infer<
  typeof notificationPreferenceSchema
>
```

### Re-export dari Index

```typescript
// packages/schemas/index.ts
export * from './payment-method'
export * from './user'
export * from './notification-channel'
```

---

## Additional Rules

- **If a field has a fixed set of values, it MUST be declared here as an `as const` array + type alias + `z.enum()`.** This is the single source of truth for enums — Prisma, BE, and FE all import from `packages/schemas/`. Values are always `SCREAMING_SNAKE_CASE`.
- Each schema file must have a test file (`*.test.ts`)
- Export constant helpers only if they are needed alongside the related schema
- Type alias suffix: `SchemaProps` for form/payload (e.g., `PaymentMethodSchemaProps`)
- File must end with newline
