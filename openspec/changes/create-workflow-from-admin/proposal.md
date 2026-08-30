## Why

Admin Console saat ini hanya menampilkan daftar workflow (read-only). Tidak ada cara bagi user untuk membuat workflow baru dari UI — satu-satunya cara adalah via API langsung. Ini memblokir operator dan admin yang tidak punya akses langsung ke backend.

## What Changes

- Tambah tombol **"New Workflow"** di halaman `/admin/workflows` (hanya tampil jika user punya permission `workflow:create`)
- Tambah dialog/drawer **Create Workflow** dengan form: `slug`, `name`, `description` (optional)
- Setelah berhasil create, redirect ke `/state-builder/{id}` agar user langsung bisa mulai authoring
- Gunakan `useWorkflowsCreate` hook yang sudah ada tapi belum dipakai di frontend
- Tambah panduan visual `Tenant → Project → Workflow → Builder` pada Admin Console
- Tampilkan scope tenant dan project yang sedang dipakai, termasuk penjelasan bahwa workflow baru saat ini masuk ke `Default Project`

## Capabilities

### New Capabilities

- `web/admin-console-management/workflow-create`: Form + dialog untuk membuat workflow baru dari Admin Console, dengan redirect ke Builder setelah berhasil.

### Modified Capabilities

- `web/admin-console-management`: Workflow inventory requirement diperluas — halaman inventory kini juga menyediakan entry point untuk membuat workflow baru bagi user dengan `workflow:create` permission.

## Impact

- `apps/web/app/admin/workflows/workflows-page-content.tsx` — tambah tombol + render dialog
- `apps/web/app/admin/workflows/_components/create-workflow-dialog.tsx` — komponen baru
- `apps/web/components/admin-console/admin-flow-guide.tsx` — panduan scope dan langkah kerja Admin Console
- `apps/web/app/admin/admin-page-content.tsx` dan `apps/web/app/admin/tenant/tenant-page-content.tsx` — tampilkan panduan alur
- `packages/schemas/workflow.ts` — `createWorkflowSchema` sudah ada, tidak perlu perubahan
- `packages/types/workflow-response.ts` — `WorkflowResponse` sudah ada, tidak perlu perubahan
- `apps/web/hooks/transactions/use-workflow/use-create-workflow.ts` — sudah ada, tidak perlu perubahan

## Non-goals

- Bukan fitur untuk edit atau publish workflow dari Admin Console (tetap di Builder)
- Tidak membuat project CRUD atau project switcher; UI menjelaskan penggunaan `Default Project` saat ini
- Tidak menambah field `projectId` ke form — dikosongkan (opsional di schema)
- Tidak membuat halaman detail workflow baru di Admin Console
