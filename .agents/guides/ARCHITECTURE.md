# Architecture: Layer Map + Folder Contracts

Read this file at the start of each implementation session. Contains complete layer map and contracts per folder.

---

## Monorepo Overview

```
open-state/
├── apps/
│   ├── web/      → Next.js 16 App Router (frontend)
│   ├── api/      → Hono (backend, Clean Architecture)
│   └── worker/   → BullMQ (background job processor)
└── packages/
    ├── schemas/  → Zod validation schemas (shared FE + BE + Worker)
    ├── types/    → API response TypeScript types (shared FE + BE)
    └── utils/    → Pure utility functions (shared all)
```

---

## apps/web — Next.js App Router

### Layer Map

```
apps/web/
├── app/                    → Router: pages, layouts, route groups, route handlers
│   └── (group)/
│       └── feature/
│           ├── page.tsx                 → Server Component (thin Suspense wrapper)
│           ├── feature-page-content.tsx → Client Component (route orchestration)
│           └── _components/             → Private components for this route
├── auth.ts                 → NextAuth credentials config (server-only)
├── proxy.ts                → Route protection / redirect logic in edge layer
├── components/             → Reusable UI components (used >1 page)
├── hooks/
│   ├── transactions/       → Data-fetching hooks per domain
│   │   └── use-{domain}/   → One folder per domain
│   │       ├── use-data-table.ts
│   │       ├── use-get-one.ts
│   │       ├── use-insert-one.ts
│   │       ├── use-update-one.ts
│   │       ├── use-delete-one.ts
│   │       └── index.ts
│   └── utility/            → Non-data hooks (useQueryParam, etc.)
├── services/
│   └── axios/              → Axios instance with interceptors
├── constants/
│   ├── api-routers.ts      → All API URL constants (use :id path variable)
│   └── query-keys.ts       → All react-query cache key constants (flat strings)
├── types/generals/         → FE-specific types not from API
├── configs/                → App-wide config (env, auth, etc.)
├── providers/              → React context providers
└── utils/                  → FE-specific helper functions (debounce, pathVariable, etc.)
```

### Data Flow

```
page.tsx (Server Component — thin Suspense wrapper)
  └→ feature-page-content.tsx (Client Component — route state, hooks, layout)
      ├→ _components/ (private form, table, dialog, drawer)
      └→ Custom Hook (hooks/transactions/use-{domain}/)
          └→ react-query useQuery / useMutation
              └→ axios instance (services/axios/)
                  └→ app/api/proxy/[...path]/route.ts
                      └→ API Server
```

### Contracts per Folder

#### `app/` — Pages & Layouts

✅ Allowed:
- `page.tsx` contains only Suspense wrapper + import content component
- Import `LoadingSpinner` or skeleton component for Suspense fallback
- Export `generateMetadata`, `generateStaticParams`
- Thin route handler for auth or proxy allowed under `app/api/`

❌ Forbidden:
- Call `axios` or `fetch` directly
- Import from `services/` directly
- Business logic or state management
- Place reusable cross-route components in `_components/` — use `apps/web/components/` instead

---

#### `app/api/(auth)/auth/[...nextauth]/route.ts` — NextAuth Route Handler

✅ Allowed:
- Wrap `NextAuth(authOptions)` and export handler `GET`/`POST`
- Stay thin and only be App Router entrypoint for auth

❌ Forbidden:
- Put login business logic directly in route handler
- Call backend auth directly here if logic already exists in `auth.ts`

---

#### `app/api/proxy/[...path]/route.ts` — Internal BFF Proxy

✅ Allowed:
- Forward browser request to backend API
- Add bearer token from NextAuth session
- Refresh access token and update session cookie when needed
- Pass public auth endpoints like login/refresh without bearer token

❌ Forbidden:
- Put feature business logic
- Add domain-specific response transformation at this route level
- Make this route a stateful cache or business orchestration point

---

#### `proxy.ts` — Edge Route Protection

✅ Allowed:
- Redirect guest to login page for protected routes
- Redirect logged-in user from `/login` to default page
- Read cookie/session token for lightweight guard
- Read config from `auth.ts` and `configs/auth-server.ts`

❌ Forbidden:
- Application business logic
- Fetch business data or call internal API to render page
- Put main auth config here — keep it in `auth.ts`

---

#### `*-page-content.tsx` — Main Client Component

✅ Allowed:
- All `useState`, `useEffect`, hooks
- Import and call hooks from `hooks/`
- Define `columns` array for table
- Form handling with react-hook-form
- Dialog state and logic
- Query params via `useQueryParam`
- Delete confirmation via SweetAlert2

❌ Forbidden:
- Call `axios` or `fetch` directly
- Import from `services/` directly

---

#### `_components/` — Private Route Components

✅ Allowed:
- Presentational or focused components used only by the current route
- Route-specific toolbar, table, form dialog, drawer, and loading components

❌ Forbidden:
- Reuse from another route — move cross-route components to `apps/web/components/`
- Direct `axios` or `fetch` calls

---

#### `components/` — Reusable UI Components

✅ Allowed:
- Accept props, render JSX
- Import UI library components (Button, Input, Dialog, Table, etc.)
- `useState`, `useEffect` for local UI state

❌ Forbidden:
- Call `axios` or `fetch` directly
- Import data-fetching hooks from `hooks/`
- Hardcode API URL or query key

---

#### `hooks/transactions/use-{domain}/` — Custom React Hooks

✅ Allowed:
- Wrap `useQuery`, `useMutation` from react-query
- Call `axios` instance directly (no need for separate service function)
- Use `queryKeys` and `apiRouters` from `constants/`
- `useDataTable` = react-query `useQuery` to fetch paginated list data

❌ Forbidden:
- Contains JSX
- One hook for all operations — separate per file
- Hardcode URL — use `apiRouters` from constants

---

#### `hooks/utility/` — Utility Hooks

✅ Allowed:
- `useQueryParam` — wrap `useSearchParams` + `useRouter` + `usePathname`

❌ Forbidden:
- Contains data-fetching or business logic

---

#### `services/axios/` — Axios Instance

✅ Allowed:
- Setup axios instance with internal proxy base URL from `configs/env.ts`
- Response interceptor to unwrap data and lightweight `401` redirect to login
- Response interceptor unwrap `{ meta, data }` → `DataTableResponse` for list

❌ Forbidden:
- Service function per endpoint — that goes directly in hook
- Business logic
- Inject bearer token browser-side if request goes through internal proxy

---

#### `constants/` — Application Constants

✅ Allowed:
- `api-routers.ts`: flat object with `:id` path variables (e.g. `/users/:id`)
- `query-keys.ts`: flat string values per operation (e.g. `index: 'usersIndex'`)

❌ Forbidden:
- Business logic
- Function for path variable — use `pathVariable()` utility
- Values from env (use `configs/`)

---

#### `packages/schemas/` — Zod Schemas (Shared)

✅ Allowed:
- Type constants array + labels array + `get{Type}Label()` helper
- Zod schema for form payload
- Export `type Props = z.infer<typeof schema>`

❌ Forbidden:
- Import FE or BE-specific library
- Business logic or API call

---

#### `packages/types/` — API Response Types (Shared)

✅ Allowed:
- TypeScript `type` for API response
- Re-export from `index.ts`

❌ Forbidden:
- Use `any`
- Request/payload types (use `packages/schemas/`)

---

## apps/api — Hono Backend (Clean Architecture)

### Layer Map

```
apps/api/src/
├── interfaces/http/
│   ├── routes/             → HTTP routing + Zod validation
│   └── controllers/        → Parse request, call service, format response
├── application/
│   ├── services/           → Orchestrate use cases, transform Entity → DTO
│   ├── use-cases/          → Business logic (one file per operation)
│   ├── dtos/               → Data Transfer Objects (output shapes)
│   └── validators/         → Zod schemas for request validation
├── domain/
│   ├── entities/           → Domain models (plain objects/classes)
│   └── repositories/       → Repository interfaces (contracts)
└── infrastructure/
    ├── config/             → Runtime config (env, database config, etc.)
    └── database/           → Prisma repository implementations
```

### Request Lifecycle

```
HTTP Request
  → routes/         (Zod validation, delegate to controller)
  → controllers/    (parse req, call service, format response)
  → services/       (orchestrate use cases, Entity → DTO)
  → use-cases/      (business logic, throw DomainError)
  → domain/repos/   (interface contract)
  → database/       (Prisma implementation, return Entity)
  ↑
  Error bubble up → errorHandler middleware → HTTP Response
```

### Error Handling

```
DomainError → errorHandler middleware
  ├── NOT_FOUND    → 404
  ├── UNAUTHORIZED → 401
  ├── FORBIDDEN    → 403
  ├── CONFLICT     → 409
  ├── VALIDATION   → 422
  └── INTERNAL     → 500
```

### Contracts per Layer

#### `interfaces/http/routes/` — HTTP Routes

✅ Allowed:
- Define HTTP method + path
- Validate request body/query with Zod schema from `validators/`
- Delegate to controller handler
- Set middleware per route (auth, rate-limit)

❌ Forbidden:
- Business logic
- Call use case or repository directly
- Format response manually

---

#### `interfaces/http/controllers/` — Controllers

✅ Allowed:
- Parse `c.req` (body, params, query)
- Call service method
- Format and return HTTP response (`c.json(...)`)

❌ Forbidden:
- Business logic
- Call use case or repository directly — must go through service
- Throw `DomainError` — that's use case's job
- `try/catch` for domain error — let it bubble to errorHandler

---

#### `application/services/` — Application Services

✅ Allowed:
- Orchestrate one or more use cases
- Transform Entity to DTO before returning to controller
- Inject and call repository or use case

❌ Forbidden:
- Business logic — that's in use case
- Access Prisma directly — must go through repository interface
- HTTP concern (status code, header)
- `try/catch` for domain error

---

#### `application/use-cases/` — Use Cases

✅ Allowed:
- Contains one business logic operation
- Throw `DomainError` for expected errors
- Call repository interface
- One file per operation: `create-user.ts`, `get-user.ts`, etc.

❌ Forbidden:
- Access Prisma or database directly
- HTTP concern (import from `hono` for response/exception)
- Throw `HTTPException` — use `DomainError`

---

#### `domain/entities/` — Domain Entities

✅ Allowed:
- Plain TypeScript type or class
- Fields matching domain model
- Pure domain methods (without external dependency)

❌ Forbidden:
- Import Prisma types directly
- HTTP or database dependency

---

#### `domain/repositories/` — Repository Interfaces

✅ Allowed:
- Define interface/abstract class
- Method signature: `findById(id: string): Promise<Entity | null>`
- Use Entity types from `domain/entities/`

❌ Forbidden:
- Concrete implementation — that's in `infrastructure/database/`
- Import Prisma
- Business logic

---

#### `infrastructure/database/` — Prisma Repositories

✅ Allowed:
- Implement repository interface from `domain/repositories/`
- Access Prisma client
- Map Prisma model to domain Entity

❌ Forbidden:
- Business logic
- Return raw Prisma model — must map to Entity
- HTTP concern

---

#### `infrastructure/config/` — Runtime Config

✅ Allowed:
- Parse env with Zod
- Setup application config needed at runtime bootstrap

❌ Forbidden:
- Business logic
- Generic cross-app utility — move to `packages/*`
- Create `shared/` folder at app level for similar config

---

## apps/worker — BullMQ Worker

### Layer Map

```
apps/worker/src/
├── infrastructure/config/  → Runtime config (env, Redis config, etc.)
├── infrastructure/queue/   → Worker setup, consume BullMQ job
├── application/use-cases/  → Process job logic
└── domain/entities/        → Job entity types
```

### Job Flow

```
BullMQ Queue (job data defined in packages/schemas/)
  → infrastructure/queue/   (Worker setup, parse + validate job data)
  → application/use-cases/  (process job: send email, sync data, etc.)
  → domain/entities/        (job entity types)
```

### Worker Notes

- Don't create `shared/` folder at worker app level
- Shared across apps stays in `packages/*`
- Worker env goes into `apps/worker/src/infrastructure/config/`

---

## packages/ — Shared Packages

### Contracts

#### `packages/schemas/` — Zod Schemas

✅ Allowed:
- Zod schema for form payload, request body, job data
- `z.infer<>` types
- Shared enum/constant values

❌ Forbidden:
- FE or BE specific imports
- Business logic, side effects

---

#### `packages/types/` — API Response Types

✅ Allowed:
- TypeScript `type`/`interface` for API responses
- Re-export from `index.ts`

❌ Forbidden:
- `any`
- Zod dependency
- Request/payload types (use `packages/schemas/`)

---

#### `packages/utils/` — Pure Utilities

✅ Allowed:
- Pure functions without side effects
- Format, transform, parse helpers

❌ Forbidden:
- Import FE-specific (React) or BE-specific (Hono, Prisma) library
- State management
- API calls
