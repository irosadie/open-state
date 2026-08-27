# Guide: Web Component (`apps/web/components/`)

## Folder Contract

✅ Allowed:
- Accept props, render JSX
- `useState`, `useEffect` for local UI state
- Import from `utils/`, `types/`, `constants/`

❌ Forbidden:
- Call `axios` or `fetch` directly
- Import data-fetching hooks from `hooks/`
- Hardcode API URL or query key
- Business logic

---

## Folder Structure

All components are flat in `components/` — no subfolders.

```
components/
├── button.tsx             → wrapper dengan prop `intent`, `loading`, `rounded`, `leftIcon`
├── input.tsx              → wrapper dengan prop `label`, `error`, `leftIcon`, `intent`, `rounded`
├── dialog.tsx             → Dialog, DialogContent, DialogHeader, DialogTitle, dll.
├── sheet.tsx              → Sheet, SheetContent, SheetHeader, SheetTitle, dll.
├── select.tsx             → custom select dengan prop `options`, `getOptionLabel`, `getOptionValue`
├── radio-group.tsx        → wrapper dengan prop `data`, `getDataLabel`, `getDataValue`
├── textarea.tsx           → wrapper dengan prop `label`, `rounded`, `intent`
├── table.tsx              → custom: data + columns + pagination built-in
├── actions-dropdown.tsx   → dropdown menu untuk action baris tabel
├── panel-card.tsx         → card wrapper untuk konten halaman panel
├── status-badge.tsx       → badge aktif/nonaktif dari boolean
└── loading-spinner.tsx    → standalone loading indicator
```

---

## Component Types

### 1. Wrapper (primitive + project props)

```tsx
// components/button.tsx
'use client'

interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  intent?: 'primary' | 'warning' | 'danger' | 'secondary'
  size?: 'small' | 'medium' | 'large'
  rounded?: 'default' | 'large' | 'full'
  loading?: boolean
  textOnly?: boolean
  leftIcon?: React.ReactNode
}

export function Button({ intent = 'primary', loading, leftIcon, children, disabled, ...props }: ButtonProps) {
  return (
    <button disabled={disabled || loading} {...props}>
      {loading ? <Spinner /> : leftIcon}
      {children}
    </button>
  )
}
```

```tsx
// components/input.tsx
'use client'

interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  label?: string
  error?: string
  leftIcon?: React.ReactNode
  intent?: 'default' | 'clean'
  rounded?: 'default' | 'large'
}

export function Input({ label, error, leftIcon, ...props }: InputProps) {
  return (
    <div className="space-y-1">
      {label && <label className="text-sm font-medium">{label}</label>}
      <div className="relative">
        {leftIcon && <span className="absolute left-3 top-1/2 -translate-y-1/2">{leftIcon}</span>}
        <input className={leftIcon ? 'pl-10' : ''} {...props} />
      </div>
      {error && <p className="text-xs text-red-500">{error}</p>}
    </div>
  )
}
```

### 2. Composite (combination of multiple elements)

```tsx
// components/actions-dropdown.tsx
'use client'

interface Action {
  label: string
  onClick: () => void
  destructive?: boolean
}

export function ActionsDropdown({ actions }: { actions: Action[] }) {
  // dropdown menu with list of actions
}
```

```tsx
// components/status-badge.tsx
interface StatusBadgeProps {
  status: boolean
  activeLabel?: string
  inactiveLabel?: string
}

export function StatusBadge({ status, activeLabel = 'Active', inactiveLabel = 'Inactive' }: StatusBadgeProps) {
  return (
    <span className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium ${
      status ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-600'
    }`}>
      {status ? activeLabel : inactiveLabel}
    </span>
  )
}
```

---

## When to Put Where

| Condition | Location |
|---|---|
| Component used in only 1 route | `app/**/_components/` |
| Component used in >1 route | `components/` |

---

## Additional Rules

- One file = one main component
- Export named — not default export
- Always forward native HTML props (`...props`) in wrappers
- Files must end with newline (EOF)
