## 1. Create Workflow Dialog Component

- [x] 1.1 Buat file `apps/web/app/admin/workflows/_components/create-workflow-dialog.tsx` dengan form fields: `slug`, `name`, `description` (optional)
- [x] 1.2 Implementasi validasi client-side menggunakan `createWorkflowSchema` dari `@openstate/schemas`
- [x] 1.3 Integrasikan `useWorkflowsCreate` hook — tampilkan loading state di tombol submit saat mutation pending
- [x] 1.4 Tampilkan field-level errors dari Zod validation
- [x] 1.5 Tampilkan error dari backend (API error) di dalam dialog
- [x] 1.6 Setelah sukses, panggil `router.push(`/state-builder/${data.id}`)` untuk redirect ke Builder

## 2. Update Workflow Inventory Page

- [x] 2.1 Import dan render `CreateWorkflowDialog` di `workflows-page-content.tsx`
- [x] 2.2 Tambah state `isCreateOpen` (boolean) untuk mengontrol visibility dialog
- [x] 2.3 Tambah tombol "New Workflow" di area header inventory — hanya render jika `authorization.hasPermission("workflow:create")`
- [x] 2.4 Pastikan tombol tidak tampil jika user tidak punya `workflow:create` permission
- [x] 2.5 Tampilkan setup path `Tenant → Project → Workflow → Builder` pada workflow inventory
- [x] 2.6 Tampilkan tenant scope dan penjelasan bahwa workflow baru masuk ke `Default Project`

## 3. Clarify Admin Console Flow

- [x] 3.1 Buat shared `AdminFlowGuide` untuk menjelaskan urutan tenant, project, workflow, dan Builder
- [x] 3.2 Render `AdminFlowGuide` pada overview dan tenant settings
- [x] 3.3 Jelaskan secara eksplisit bahwa project switcher/CRUD belum tersedia

## 4. Verification

- [x] 4.1 Tambah test untuk setup path dan `Default Project` copy
- [x] 4.2 Jalankan `cd apps/web && bun run typecheck` — pastikan tidak ada type error
- [x] 4.3 Jalankan `cd apps/web && bun run lint` — pastikan tidak ada lint error (no any, no console, no unused imports)
