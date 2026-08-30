## Context

Halaman `/admin/workflows` saat ini hanya menampilkan daftar workflow (read-only). Hook `useWorkflowsCreate` dan schema `createWorkflowSchema` sudah tersedia tapi belum digunakan di frontend. Backend endpoint `POST /workflows` sudah ada. Tidak ada perubahan backend yang diperlukan.

## Goals / Non-Goals

**Goals:**
- Tambah tombol "New Workflow" di inventory page, hanya tampil untuk user dengan `workflow:create`
- Dialog form dengan field `slug`, `name`, `description` (optional)
- Validasi client-side via `createWorkflowSchema` (Zod)
- Submit via `useWorkflowsCreate`, redirect ke `/state-builder/{id}` setelah sukses
- Surface error dari backend di dalam dialog
- Tampilkan setup path bersama pada Admin Console: `Tenant → Project → Workflow → Builder`
- Tampilkan project scope eksplisit pada inventory workflow agar operator memahami bahwa workflow baru masuk ke `Default Project`

**Non-Goals:**
- Tidak mengubah backend
- Tidak menambah field `projectId` ke form
- Tidak membuat halaman detail workflow baru di Admin Console
- Tidak mengubah flow edit/publish (tetap di Builder)
- Tidak menambah project management; status project saat ini disampaikan sebagai context-only UI

## Decisions

### 1. Dialog bukan page baru

Gunakan dialog (`_components/create-workflow-dialog.tsx`) yang di-trigger dari tombol di inventory page — bukan navigasi ke route baru.

**Alasan:** Form create workflow sangat sederhana (3 field). Dialog cukup, tidak perlu full page. Pola ini konsisten dengan cara Admin Console menangani mutation lain (confirm dialog sebelum submit).

**Alternatif:** Dedicated `/admin/workflows/new` page — overkill untuk 3 field.

### 2. Gunakan `useWorkflowsCreate` yang sudah ada

Hook sudah menghandle invalidasi query list setelah sukses. Tidak perlu buat hook baru.

### 3. Redirect ke Builder setelah sukses via `router.push`

Setelah `mutateAsync` resolve, panggil `router.push(`/state-builder/${data.id}`)`. Dialog tidak perlu ditutup secara eksplisit karena navigasi sudah meninggalkan halaman.

**Alternatif:** Tutup dialog + refresh list — tidak sesuai spec, karena spec mewajibkan redirect ke Builder.

### 4. Setup path sebagai shared Admin Console component

Gunakan `AdminFlowGuide` di overview, tenant settings, dan workflow inventory.
Komponen ini menampilkan urutan kerja serta menandai `Default Project` sebagai
project otomatis saat ini. Langkah Project tidak dibuat sebagai link karena belum
ada project CRUD/switcher di frontend.

**Alasan:** Operator perlu memahami scope sebelum membuat workflow, sementara
menambahkan project management akan memperluas perubahan ke kontrak API dan RBAC.

### 4. Komponen dialog sebagai `_components/` private

Komponen hanya digunakan oleh route `/admin/workflows`, sehingga masuk ke `_components/` sesuai konvensi folder contract.

## Risks / Trade-offs

- **[Risk] Slug collision** → Backend akan return error conflict. Dialog menampilkan pesan error dari API, user bisa koreksi slug.
- **[Risk] Navigasi meninggalkan halaman sebelum invalidasi query selesai** → Tidak masalah karena list query di-invalidate oleh `useWorkflowsCreate.onSuccess`, dan halaman inventory akan fresh saat user kembali dari Builder.
- **[Risk] Project scope belum dapat dipilih** → UI menyatakan `Default Project` secara eksplisit dan tidak memberi kesan bahwa project switcher sudah tersedia.

## Migration Plan

Tidak ada migration data atau breaking change. Perubahan murni additive di frontend.
