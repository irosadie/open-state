# Context: Web Slicing

## Target Folder

```
apps/web/
├── app/                  → page.tsx + *-page-content.tsx + _components/
└── components/           → reusable UI (only when shared across pages)
```

## Page Structure (Required)

```
app/(group)/[feature]/
├── page.tsx                    → Server Component, Suspense wrapper only
├── [feature]-page-content.tsx  → Client Component, route orchestration
└── _components/                → Components private to this route
```

Use `_components/` for private dialogs, forms, tables, and drawers. Move components used by more than one route to `apps/web/components/`.

## Real Code Examples

### Example 1 — Basic CRUD (Template)
Location: `.agents/examples/web-slicing/nextjs-app-router/examples/`

Pattern: useDataTable + useQueryParam + debounce + SweetAlert preConfirm + ActionsDropdown

Files:
- `page.tsx` — Suspense wrapper
- `examples-content.tsx` — route orchestration: list, search, form dialog, delete confirm

**Use for:** CRUD pages with simple fields.

---

## Key Patterns

### `page.tsx` — Suspense Wrapper

```tsx
import { Suspense } from 'react'
import FeaturePageContent from './feature-page-content'
import { LoadingSpinner } from '$/components/loading-spinner'

export default function FeaturePage() {
  return (
    <Suspense fallback={<LoadingSpinner />}>
      <FeaturePageContent />
    </Suspense>
  )
}
```

### Search with `useQueryParam` + `debounce`

```tsx
const [tempQuery, setTempQuery] = useState(currentSearch)
const { setQueryParams } = useQueryParam()

const debouncedSearch = useMemo(
  () => debounce((value: string) => {
    setQueryParams({ q: value.trim() || undefined, page: 1 })
  }, 250),
  [setQueryParams],
)

const handleOnSearch = (e) => {
  setTempQuery(e.target.value)
  debouncedSearch(e.target.value)
}
```

### SweetAlert Delete — `preConfirm` + Promise

```tsx
const handleOnDelete = (id: string) => {
  Swal.fire({
    text: 'Are you sure you want to delete this data?',
    icon: 'warning',
    showCancelButton: true,
    confirmButtonColor: '#e42c2c',
    cancelButtonColor: '#3278A0',
    confirmButtonText: 'Yes, delete it!',
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
      Swal.getCancelButton()!.style.order = '-1'
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

### Table + Pagination

```tsx
<PanelCard>
  <Table
    data={items || []}
    columns={columns}
    intent="clean"
    isLoading={isLoading}
    isShowPagination={true}
    pagination={{
      page: currentPage,
      limit: limit,
      totalData: paginationData?.total || 0,
      onPageClick(newPage) { setQueryParams({ page: newPage }) },
    }}
  />
</PanelCard>
```

### ActionsDropdown

```tsx
<ActionsDropdown
  actions={[
    { label: 'Edit', onClick: () => handleOpenDialog(true, row) },
    { label: 'Delete', onClick: () => handleOnDelete(row.id), destructive: true },
  ]}
/>
```

## Tech Stack

| Technology | Use |
|-----------|-----|
| Next.js App Router | Routing, layout, page |
| Tailwind CSS | Styling |
| react-hook-form + zodResolver | Form state + validation |
| SweetAlert2 | Delete confirmation |
| react-hot-toast | Success/error notifications |
| Lucide React | Icons |
| useQueryParam | Filter & pagination in URL |
| debounce utility | Debounce search input |
