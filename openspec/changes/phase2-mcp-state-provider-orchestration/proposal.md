## Why

OpenState saat ini sudah dapat menemukan intent dan workflow, tetapi belum memberi LLM kontrak yang tegas tentang MCP provider mana yang wajib dipakai pada setiap state. Akibatnya, LLM belum menerima hubungan yang terstruktur antara state, capability, provider MCP `http://localhost:8031/mcp`, dan transition yang boleh dilakukan.

Phase 2 melengkapi alur dua-MCP agar State MCP menjadi gatekeeper yang dapat menjelaskan requirement provider kepada LLM dan memastikan workflow tidak maju sebelum capability wajib menghasilkan data yang tervalidasi.

## What Changes

- Tambahkan server-level instructions pada State MCP yang menjelaskan bahwa OpenState adalah sumber kebenaran intent, state, dan transition.
- Tambahkan kontrak `requiredCapabilities` pada hasil resolusi intent dan pembacaan current state.
- Petakan logical capability ke nama/alias provider MCP server dan tool yang konkret, misalnya `padel-provider` + `padel.check_available`.
- Pisahkan identitas provider MCP yang stabil dari endpoint lokal/deployment sehingga workflow tidak bergantung pada URL localhost atau meneruskan koneksi MCP.
- Sediakan status requirement dan execution evidence untuk setiap capability wajib dalam workflow instance.
- Tolak proposal transition jika capability wajib pada state belum berhasil dijalankan atau hasilnya belum tersimpan pada context instance.
- Hubungkan requirement capability ke nama/alias MCP provider yang sudah dikonfigurasi oleh host/LLM; State MCP tidak meneruskan atau membuka koneksi provider.
- Pastikan provider mock yang dikonfigurasi pada MCP host mengembalikan tool domain-nya sendiri melalui `tools/list`, bukan tool State MCP.
- Tambahkan smoke test curl untuk koneksi LLM ke dua MCP: discovery State MCP, discovery provider MCP, pemanggilan provider, dan transition State MCP.
- Dokumentasikan perbedaan konfigurasi development (`8030` dan `8031`) dengan deployment production.

## Capabilities

### New Capabilities

- `mcp/state-provider-contract`: Kontrak machine-readable yang menjelaskan provider MCP dan tool yang diperlukan oleh intent/state, termasuk requirement, tujuan, input mapping, output mapping, dan execution evidence.

### Modified Capabilities

- `mcp/server-runtime`: State MCP mengirimkan server instructions dan metadata provider requirement yang dapat dipahami LLM melalui kontrak MCP.
- `mcp/orchestrator-tools`: `resolve_intent` dan `get_current_state` mengembalikan required capabilities; transition hanya dapat dilakukan setelah requirement state terpenuhi.

## Impact

- `apps/api/internal/interfaces/mcp/`: server instructions, response contract, provider requirement metadata, dan enforcement boundary.
- `apps/api/internal/domain/engine/` dan `apps/api/internal/application/services/`: validasi requirement capability sebelum event transition.
- `apps/api/internal/domain/capability/`: provider server alias, concrete tool mapping, dan execution evidence.
- `apps/api/internal/interfaces/http/` dan `apps/web/`: admin contract/UI untuk memilih provider server alias dan tool tanpa memasukkan endpoint ke workflow.
- `apps/mcp-provider-mock/`: runtime verification pada konfigurasi provider mock lokal dan contract tests untuk padel, doctor, dan food provider tools.
- Workflow definition, capability registry, dan binding data: penambahan metadata provider/tool tanpa menyimpan secret.
- MCP client/LLM host configuration: wajib mendaftarkan State MCP dan provider MCP sebagai dua server terpisah; endpoint diputuskan oleh host, bukan workflow.

## Non-goals

- Menghubungkan provider mock ke sistem production atau database production.
- Membuat State MCP meneruskan atau membuka koneksi ke MCP provider atas nama LLM.
- Mengubah State MCP menjadi penyimpan data pihak ketiga.
- Menghapus `invoke_capability`; tool tersebut tetap menjadi jalur terotorisasi untuk mode eksekusi yang dikelola OpenState.
- Membuat sistem credential vault production baru; Phase 2 hanya menggunakan reference/configuration yang sudah tersedia.
