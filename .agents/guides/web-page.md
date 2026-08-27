# Guide: Web Page (`apps/web/app/`)

## Folder Contract

✅ Allowed:
- `page.tsx` = thin Suspense wrapper (Server Component)
- `_components/` = private components to that route (not used by other routes)
- `*-page-content.tsx` = main Client Component (state, hooks, layout)
- Components in `_components/` such as: `*-toolbar.tsx`, `*-table-card.tsx`, `*-form-dialog.tsx`, `*-drawer.tsx`
- Export `generateMetadata`, `generateStaticParams`
- Use `Suspense`, `loading.tsx`, `error.tsx`
- Route group `(group)/` for layout without URL segment

❌ Forbidden:
- Call `axios` or `fetch` directly
- Import from `services/` directly
- Business logic or state management in `page.tsx`
- Share `_components/` between routes — if reusable, move to `$/components/`

---

## File Structure per Route

```
app/
└── (dashboard)/
    └── users/
        ├── page.tsx                        → Server Component (thin Suspense wrapper)
        ├── users-page-content.tsx          → Main Client Component (route orchestration)
        ├── loading.tsx                     → Skeleton placeholder (optional)
        ├── error.tsx                       → Error boundary (optional)
        └── _components/
            ├── users-toolbar.tsx           → Search bar + filter + add button
            ├── users-table-card.tsx        → Table in PanelCard
            ├── users-form-dialog.tsx       → Create/edit dialog
            └── users-page-loading.tsx      → Skeleton loading (optional)
```

> Components that are **only** used in this route go in `_components/`. If used in more than one route, move to `$/components/`.

### `page.tsx` — Thin Wrapper with Suspense

```tsx
import { Suspense } from 'react'
import { UsersPageContent } from './users-page-content'
import { UsersPageLoading } from './_components/users-page-loading'

export default function UsersPage() {
  return (
    <Suspense fallback={<UsersPageLoading />}>
      <UsersPageContent />
    </Suspense>
  )
}
```

### `loading.tsx` — Route-level Loading

```tsx
import { UsersPageLoading } from './_components/users-page-loading'

export default function Loading() {
  return <UsersPageLoading />
}
```

---

## Prioritize Reusable Components

**Always** use components from `$/components/` — don't write primitive HTML if the component already exists.

| Need          | Component                                          |
|-------------------|----------------------------------------------------|
| Button             | `Button`                                          |
| Text input         | `Input`                                           |
| Textarea           | `Textarea`                                        |
| Dropdown select    | `Select`                                          |
| Radio buttons      | `RadioGroup`                                      |
| Checkbox           | `Checkbox`                                        |
| Table + pagination | `Table`                                           |
| Card container     | `PanelCard`                                       |
| Modal              | `Dialog`, `DialogContent`, `DialogHeader`, etc.   |
| Side panel/drawer  | `Sheet`, `SheetContent`, `SheetHeader`, etc.      |
| Row action menu  | `ActionsDropdown`                                 |
| Status label       | `StatusBadge`                                     |
| Badge              | `Badge`                                           |
| Loading            | `LoadingSpinner`                                  |

---

## Query Params — `useSearchParams` vs `useQueryParam`

**`useSearchParams`** (from `next/navigation`) → only for **reading** values from URL.
**`useQueryParam`** (custom hook `$/hooks/utility/use-query-param`) → for **changing** URL params.

```tsx
'use client'

import { useSearchParams } from 'next/navigation'
import { useQueryParam } from '$/hooks/utility/use-query-param'

export function UsersPageContent() {
  const searchParams = useSearchParams()

  // Read from URL
  const currentPage = Number(searchParams.get('page') || '1')
  const currentSearch = searchParams.get('q') || ''

  // Write to URL
  const { setQueryParams } = useQueryParam()

  function handlePageChange(newPage: number) {
    setQueryParams({ page: newPage })
  }

  // undefined → remove param from URL
  function handleClearSearch() {
    setQueryParams({ q: undefined, page: 1 })
  }
}
```

### Debounced Search Pattern

```tsx
import { useState, useEffect, useMemo } from 'react'
import { debounce } from '$/utils/debounce'  // not useDebounce hook

const [tempQuery, setTempQuery] = useState(currentSearch)

// Sync input when URL changes (e.g.: back button)
useEffect(() => {
  setTempQuery(currentSearch)
}, [currentSearch])

const debouncedSearch = useMemo(
  () =>
    debounce((value: string) => {
      setQueryParams({ q: value.trim() || undefined, page: 1 })
    }, 250),
  [setQueryParams],
)

// Cleanup debounce on unmount
useEffect(() => {
  return () => debouncedSearch.cancel()
}, [debouncedSearch])

const handleOnSearch = (e: React.ChangeEvent<HTMLInputElement>) => {
  const value = e.target.value
  setTempQuery(value)       // update input immediately (immediate feedback)
  debouncedSearch(value)    // update URL after delay
}
```

**Rules:**
- `tempQuery` = local state for input (immediate UI feedback)
- `currentSearch` = value from URL (used for API call)
- `debounce` from `$/utils/debounce` + `useMemo` — **not** `useDebounce` hook

---

## Table + `useDataTable`

`useDataTable` = react-query hook that fetches paginated data. **Not** a TanStack Table wrapper — no `useReactTable` here.

### 1. Define Columns

```tsx
import { Table, type ColumnDef } from '$/components/table'
import { ActionsDropdown } from '$/components/actions-dropdown'
import { StatusBadge } from '$/components/status-badge'
import type { UserResponseProps } from '@vibecoding-starter/types/user-response'

const columns: ColumnDef<UserResponseProps>[] = [
  {
    accessorKey: 'name',
    header: 'Name',
    cell: (info) => (
      <span className="text-sm font-medium text-gray-900">
        {info.getValue() as string}
      </span>
    ),
  },
  {
    accessorKey: 'isActive',
    header: 'Status',
    cell: (info) => <StatusBadge status={info.getValue() as boolean} />,
  },
  {
    id: 'actions',
    header: '',
    cell: (info) => (
      <ActionsDropdown
        actions={[
          {
            label: 'Edit',
            onClick: () => handleOpenDialog(true, info.row.original),
          },
          {
            label: 'Delete',
            onClick: () => handleOnDelete(info.row.original.id),
            destructive: true,
          },
        ]}
      />
    ),
  },
]
```

### 2. Call Hook + Render Table

```tsx
const { data, isLoading, refetch, pagination, limit } = useUsersDataTable({
  filter: { search: currentSearch || undefined },
  page: currentPage,
  isAutoFetch: true,
})

return (
  <PanelCard>
    <Table
      data={data || []}
      columns={columns}
      intent="clean"
      isLoading={isLoading}
      isShowPagination={true}
      wrapperClassName="overflow-x-auto"
      thClassName="whitespace-nowrap"
      pagination={{
        page: currentPage,
        limit: limit,
        totalData: pagination?.total || 0,
        onPageClick(newPage) {
          setQueryParams({ page: newPage })
        },
      }}
    />
  </PanelCard>
)
```

**Rules:**
- Always wrap `Table` in `PanelCard`
- `isLoading` from hook directly passed to `Table` prop
- Use `currentSearch` (from URL) as API filter — not `tempQuery`

---

## Dialog (Modal)

Dialog can be extracted to `_components/*-form-dialog.tsx` or kept inline in page-content — adjust based on complexity.

### State and Handlers (in page-content)

```tsx
const [isDialogOpen, setIsDialogOpen] = useState(false)
const [editingId, setEditingId] = useState<string | null>(null)

function handleOpenDialog(editing = false, data?: EntityResponseProps) {
  if (editing && data) {
    setEditingId(data.id)
    reset({ name: data.name, isActive: data.isActive })
  } else {
    setEditingId(null)
    reset(initialFormData)
  }
  setIsDialogOpen(true)
}

function handleCloseDialog() {
  setIsDialogOpen(false)
  setEditingId(null)
  reset(initialFormData)
}
```

### Dialog with Scrollable Form (standard pattern)

```tsx
import {
  Dialog, DialogContent, DialogDescription,
  DialogFooter, DialogHeader, DialogTitle,
} from '$/components/dialog'

<Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
  <DialogContent className="w-[calc(100vw-2rem)] sm:min-w-[40%] sm:max-w-3xl max-h-[90vh] overflow-hidden flex flex-col gap-0 p-0">
    {/* Header fixed — doesn't scroll */}
    <DialogHeader className="shrink-0 bg-white border-b border-gray-200 px-6 py-4 pr-12">
      <DialogTitle>{editingId ? 'Edit User' : 'Add User'}</DialogTitle>
      <DialogDescription>
        {editingId ? 'Update details below.' : 'Fill in the details below.'}
      </DialogDescription>
    </DialogHeader>

    <form onSubmit={handleSubmit(onSubmit)} className="flex flex-1 flex-col min-h-0">
      {/* Scrollable area */}
      <div className="flex-1 min-h-0 overflow-y-auto px-6 pt-0 pb-6 space-y-4">
        <Input
          label="Name"
          placeholder="Full name"
          required
          error={errors.name?.message}
          {...register('name')}
          rounded="large"
          intent="clean"
        />
        <Select
          label="Role"
          required
          options={roleOptions}
          value={roleOptions.find(o => o.value === formValues.role)}
          onChange={(option) => setValue('role', option?.value)}
          getOptionLabel={(o) => o.label}
          getOptionValue={(o) => o.value}
          error={errors.role?.message}
          rounded="large"
          intent="clean"
        />
        <RadioGroup
          label="Status"
          data={statusOptions}
          checkedValue={String(formValues.isActive ?? true)}
          onChange={(e) => setValue('isActive', e.target.value === 'true')}
          getDataLabel={(o) => o.label}
          getDataValue={(o) => o.value}
          position="horizontal"
          name="isActive"
        />
      </div>

      {/* Footer fixed at bottom */}
      <DialogFooter className="shrink-0 bg-white border-t border-gray-200 px-6 py-4">
        <Button type="button" intent="warning" rounded="large" textOnly onClick={handleCloseDialog} disabled={isPending}>
          Cancel
        </Button>
        <Button type="submit" intent="primary" rounded="large" loading={isPending}>
          {editingId ? 'Update' : 'Save'}
        </Button>
      </DialogFooter>
    </form>
  </DialogContent>
</Dialog>
```

**Rules:**
- `shrink-0` on header and footer so they don't scroll
- `flex-1 min-h-0 overflow-y-auto` on form content area
- `pr-12` on header to make room for built-in dialog close button
- `onOpenChange={setIsDialogOpen}` so it closes when clicking backdrop

### Edit via URL Param

```tsx
const editId = searchParams.get('edit')

const { data: editData, isLoading: isEditLoading } = useUsersGetOne({ id: editId || '' })

useEffect(() => {
  if (editData) handleOpenDialog(true, editData)
}, [editData])

if (isEditLoading) return null
```

---

## Drawer (Side Panel)

Use `Sheet` for side panels — complex filters, detail views, or forms that don't need a full modal.
Can be extracted to `_components/*-drawer.tsx` or inline in page-content.

```tsx
import {
  Sheet, SheetContent, SheetHeader,
  SheetTitle, SheetFooter,
} from '$/components/sheet'

const [isFilterOpen, setIsFilterOpen] = useState(false)
const [activeFilter, setActiveFilter] = useState({ status: '', role: '' })
```

### Render

```tsx
{/* Trigger */}
<Button intent="secondary" rounded="large" onClick={() => setIsFilterOpen(true)}>
  Filter
</Button>

<Sheet open={isFilterOpen} onOpenChange={setIsFilterOpen}>
  <SheetContent side="right" className="w-[400px] flex flex-col gap-0 p-0">
    {/* Header fixed */}
    <SheetHeader className="shrink-0 border-b border-gray-200 px-6 py-4">
      <SheetTitle>Filter</SheetTitle>
    </SheetHeader>

    {/* Scrollable content */}
    <div className="flex-1 min-h-0 overflow-y-auto px-6 py-4 space-y-4">
      <Select
        label="Status"
        options={statusOptions}
        value={statusOptions.find(o => o.value === activeFilter.status)}
        onChange={(option) => setActiveFilter(prev => ({ ...prev, status: option?.value ?? '' }))}
        getOptionLabel={(o) => o.label}
        getOptionValue={(o) => o.value}
        rounded="large"
        intent="clean"
      />
    </div>

    {/* Footer fixed */}
    <SheetFooter className="shrink-0 border-t border-gray-200 px-6 py-4">
      <Button
        intent="warning"
        rounded="large"
        textOnly
        onClick={() => {
          setActiveFilter({ status: '', role: '' })
          setQueryParams({ status: undefined, role: undefined, page: 1 })
          setIsFilterOpen(false)
        }}
      >
        Reset
      </Button>
      <Button
        intent="primary"
        rounded="large"
        onClick={() => {
          setQueryParams({
            status: activeFilter.status || undefined,
            role: activeFilter.role || undefined,
            page: 1,
          })
          setIsFilterOpen(false)
        }}
      >
        Apply
      </Button>
    </SheetFooter>
  </SheetContent>
</Sheet>
```

**Rules:**
- `side="right"` for filter/detail panels
- Same structure as Dialog: header + scrollable content + footer
- Trigger with button — **not** via URL param
- Apply filter → update URL params via `setQueryParams`, then close drawer

---

## SweetAlert2 Delete — `preConfirm` + Promise

```tsx
const handleOnDelete = (id: string) => {
  Swal.fire({
    text: 'Are you sure you want to delete this data?',
    icon: 'warning',
    showCancelButton: true,
    confirmButtonColor: '#e42c2c',
    cancelButtonColor: '#3278A0',
    confirmButtonText: 'Yes, delete!',
    cancelButtonText: 'Cancel',
    preConfirm: () => {
      Swal.showLoading()
      return new Promise((resolve, reject) => {
        deleteMutate(id, {
          onSuccess: () => resolve(null),
          onError: () => reject(),
        })
      })
    },
    didOpen: () => {
      const cancelButton = Swal.getCancelButton()
      cancelButton!.style.order = '-1'  // cancel on left
    },
  })
    .then((result) => {
      if (result.isConfirmed) {
        toast.success('Data deleted successfully!')
        setQueryParams({ page: 1 })
        refetch()
      }
    })
    .catch(() => {
      Swal.hideLoading()
      Swal.close()
      toast.error('Failed to delete data. Please try again.')
    })
}
```

**Rules:**
- `deleteMutate` = `mutate` (not `mutateAsync`) from `useFeatureDeleteOne()`
- `preConfirm` + `new Promise` — **don't** use `async/await` at `Swal.fire()` level
- `.then()/.catch()` chain outside
- `toast.success()`/`toast.error()` from `react-hot-toast` — Swal only for confirmation

---

## Additional Rules

- `'use client'` only in Client Components (usually in `_components/`)
- `page.tsx` doesn't need `'use client'`
- One `_components/` per route — not shared with other routes
- Component name = PascalCase, file name = kebab-case
- All files must end with newline (EOF)
