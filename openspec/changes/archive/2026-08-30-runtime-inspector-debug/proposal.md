## Why

Operator belum dapat melihat eksekusi workflow yang sudah berjalan tanpa memakai
tool MCP atau membaca data internal secara manual. Runtime Inspector diperlukan
agar operator dapat memahami instance, state aktif, context, dan timeline; Debug
View diperlukan agar masalah per turn dapat ditelusuri melalui jejak yang dimiliki
OpenState sendiri.

## What Changes

- **NEW** — API Runtime Inspector yang tenant-scoped untuk mencantumkan dan
  mengambil detail instance: workflow/version yang dipin, lifecycle, state aktif,
  context yang sudah disanitasi, event timeline, dan correlation id.
- **NEW** — penyimpanan append-only debug trace per turn dari tahap yang dikelola
  OpenState: intent resolution, state/transition decision, context compilation,
  serta request/result MCP capability.
- **NEW** — metadata trace provider eksternal untuk LLM agent dan RAG: provider
  alias, operation/reference id, status, durasi, correlation id, dan ringkasan
  tersanitasi. Inspector tidak memanggil, mengautentikasi ke, atau membaca sistem
  LLM/RAG pihak ketiga secara langsung.
- **NEW** — halaman Runtime Inspector dan Debug View di Admin Console, dengan
  daftar instance, detail state/context/timeline, serta urutan trace per turn.
- **NEW** — permission read terpisah untuk runtime dan debug trace, dengan
  redaksi wajib untuk secret, credential, PII sensitif, raw prompt/response, dan
  dokumen RAG.

## Capabilities

### New Capabilities

- `backend/runtime-inspector`: read model dan API terautentikasi untuk daftar dan
  detail runtime instance beserta state, context, dan timeline yang tenant-scoped.
- `backend/runtime-debug-trace`: kontrak serta penyimpanan jejak per turn yang
  menghubungkan aktivitas OpenState dengan metadata referensi provider eksternal
  tanpa mengakses provider tersebut.
- `web/runtime-inspector`: halaman operator untuk mencari instance, memeriksa
  runtime state/context/timeline, dan melihat Debug View yang telah disanitasi.

### Modified Capabilities

- `auth/role-permission`: the role-permission matrix gains a least-privilege
  `debug:read` permission for the sanitized Debug View, while Runtime Inspector
  continues to require the existing `instance:read` permission.

## Impact

- **Backend:** entity/repository/migration/sqlc untuk debug trace; DTO, service,
  controller, route, authorization, dan OpenAPI untuk query runtime inspector.
  Data instance, state, history, context, audit correlation, dan trace projection
  dipakai sebagai sumber internal.
- **Frontend:** Zod schemas, response types, constants, React Query hooks, dan
  route Admin Console untuk list/detail inspector serta debug timeline.
- **Integrations:** adapter LLM/RAG/MCP hanya dapat mengirim metadata terstruktur
  yang sudah disanitasi ke OpenState. Tidak ada SDK, credential, atau network call
  baru dari inspector ke provider pihak ketiga.

## Non-Goals

- Menjalankan simulasi workflow; itu ditangani change
  `simulation-sandbox-workflow`.
- Mengubah keputusan state machine, guard, atau mengeksekusi ulang event.
- Menghubungi dashboard/API LLM, RAG, MCP provider, atau observability collector
  pihak ketiga dari aplikasi.
- Menyimpan atau menampilkan raw prompt, raw model response, dokumen retrieval,
  secret, credential, maupun PII sensitif.
- Menggantikan OpenTelemetry/OTLP atau UI observability eksternal untuk analisis
  infrastruktur tingkat rendah.

## Dependencies

- Runtime engine, persistence instance/event/context, dan MCP orchestration dari
  Epic #2–#4 yang sudah tersedia.
- Correlation id dan tracing dari Epic #6 untuk menghubungkan audit/runtime trace.
- Change ini dapat berjalan independen dari `simulation-sandbox-workflow`; keduanya
  hanya berbagi definisi workflow dan aturan redaksi.
