import { PanelCard } from "$/components/panel-card"

export default function HomeContent() {
  return (
    <main className="mx-auto flex min-h-screen w-full max-w-5xl flex-col gap-6 px-6 py-12">
      <PanelCard className="rounded-3xl p-2" noPadding>
        <div className="p-8">
          <p className="text-sm font-semibold uppercase tracking-[0.3em] text-brand-dark">
            vibecoding-starter
          </p>
          <h1 className="mt-3 text-3xl font-bold text-slate-900">
            Starter kosong untuk mulai vibe coding dari nol
          </h1>
          <p className="mt-3 max-w-3xl text-slate-600">
            Repo ini sengaja dimulai tanpa feature aktif. Yang disediakan adalah
            scaffold web, API, worker, dokumentasi, dan sistem agent supaya
            fitur pertama bisa dibangun dengan flow yang rapi.
          </p>
        </div>
      </PanelCard>

      <section className="grid gap-4 md:grid-cols-3">
        <PanelCard
          className="rounded-3xl"
          title="Web"
          description="Next.js App Router"
        >
          <p className="text-xl font-semibold text-slate-900">
            Homepage starter sudah siap dipakai
          </p>
        </PanelCard>

        <PanelCard
          className="rounded-3xl"
          title="API"
          description="Hono + Clean Architecture"
        >
          <p className="text-xl font-semibold text-slate-900">
            Baseline route ` / ` dan ` /health ` aktif
          </p>
        </PanelCard>

        <PanelCard
          className="rounded-3xl"
          title="Worker"
          description="Background runtime scaffold"
        >
          <p className="text-xl font-semibold text-slate-900">
            Belum ada queue aktif secara default
          </p>
        </PanelCard>
      </section>

      <PanelCard
        className="rounded-3xl bg-slate-900 text-slate-100 ring-0"
        noPadding
      >
        <div className="flex flex-col gap-4 p-8 md:flex-row md:items-center md:justify-between">
          <div>
            <p className="text-sm font-semibold uppercase tracking-[0.2em] text-slate-300">
              Start Here
            </p>
            <h2 className="mt-4 text-2xl font-semibold">
              Mulai dari flow, bukan dari demo
            </h2>
            <ul className="mt-2 space-y-2 text-sm text-slate-300">
              <li>Ketik `Mulai` untuk onboarding session.</li>
              <li>Buat PRD/TRD fitur pertama di `docs/features/`.</li>
              <li>
                Gunakan skill flow untuk bootstrap tiket dan implementasi.
              </li>
            </ul>
          </div>
          <div className="rounded-2xl bg-white/10 px-5 py-4 text-sm text-slate-200">
            Tidak ada route demo yang dikunci di starter ini.
          </div>
        </div>
      </PanelCard>
    </main>
  )
}
