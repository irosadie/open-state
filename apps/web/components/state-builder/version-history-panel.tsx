import type {
  WorkflowDiffResponse,
  WorkflowVersionResponse,
} from "@openstate/types"

interface VersionHistoryPanelProps {
  versions: WorkflowVersionResponse[]
  isLoading: boolean
  error: string | null
  baseVersion: number | null
  targetVersion: number | null
  onBaseVersionChange: (version: number | null) => void
  onTargetVersionChange: (version: number | null) => void
  diff: WorkflowDiffResponse | undefined
  isDiffLoading: boolean
  onClose: () => void
}

function DiffGroup({
  label,
  group,
}: {
  label: string
  group: WorkflowDiffResponse["nodes"]
}) {
  const total = group.added.length + group.removed.length + group.changed.length
  if (total === 0) return null
  return (
    <section className="space-y-1.5">
      <h4 className="text-xs font-semibold uppercase tracking-wide text-slate-500">
        {label}
      </h4>
      {group.added.map((item) => (
        <p key={`added-${item.id}`} className="text-xs text-emerald-700">
          + {item.id}
        </p>
      ))}
      {group.removed.map((item) => (
        <p key={`removed-${item.id}`} className="text-xs text-red-700">
          − {item.id}
        </p>
      ))}
      {group.changed.map((item) => (
        <p key={`changed-${item.id}`} className="text-xs text-amber-700">
          ~ {item.id}: {item.changedFields?.join(", ") || "definition"}
        </p>
      ))}
    </section>
  )
}

export function VersionHistoryPanel({
  versions,
  isLoading,
  error,
  baseVersion,
  targetVersion,
  onBaseVersionChange,
  onTargetVersionChange,
  diff,
  isDiffLoading,
  onClose,
}: VersionHistoryPanelProps) {
  return (
    <aside
      className="absolute right-3 top-3 z-30 flex max-h-[calc(100%-1.5rem)] w-80 flex-col overflow-hidden rounded-lg border border-slate-200 bg-white shadow-xl"
      data-testid="version-history-panel"
    >
      <div className="flex items-center justify-between border-b border-slate-200 px-4 py-3">
        <div>
          <h3 className="text-sm font-semibold text-slate-800">
            Version history
          </h3>
          <p className="text-[11px] text-slate-400">
            Published snapshots are read-only
          </p>
        </div>
        <button
          type="button"
          onClick={onClose}
          className="rounded px-2 py-1 text-xs text-slate-500 hover:bg-slate-100"
        >
          Tutup
        </button>
      </div>
      <div className="sb-scroll space-y-4 overflow-y-auto p-4">
        {isLoading ? (
          <p className="text-xs text-slate-500">Memuat history…</p>
        ) : null}
        {error ? <p className="text-xs text-red-600">{error}</p> : null}
        {!isLoading && !error && versions.length === 0 ? (
          <p className="text-xs text-slate-500">Belum ada versi published.</p>
        ) : null}
        {versions.length > 0 ? (
          <div className="space-y-2" data-testid="version-list">
            {versions.map((version) => (
              <div
                key={version.id}
                className="flex items-center justify-between rounded border border-slate-100 px-3 py-2"
              >
                <div>
                  <p className="text-xs font-medium text-slate-700">
                    v{version.versionNo} {version.isCurrent ? "· current" : ""}
                  </p>
                  <p className="text-[11px] text-slate-400">
                    {new Date(version.createdAt).toLocaleString()}
                  </p>
                </div>
                <span className="text-[11px] text-slate-500">
                  {version.status}
                </span>
              </div>
            ))}
          </div>
        ) : null}
        {versions.length >= 2 ? (
          <div className="space-y-2 border-t border-slate-100 pt-3">
            <p className="text-xs font-semibold text-slate-700">
              Compare versions
            </p>
            <label className="block text-[11px] text-slate-500">
              Base
              <select
                value={baseVersion ?? ""}
                onChange={(event) =>
                  onBaseVersionChange(Number(event.target.value) || null)
                }
                className="mt-1 w-full rounded border border-slate-200 px-2 py-1.5 text-xs"
              >
                <option value="">Pilih versi</option>
                {versions.map((version) => (
                  <option key={`base-${version.id}`} value={version.versionNo}>
                    v{version.versionNo}
                  </option>
                ))}
              </select>
            </label>
            <label className="block text-[11px] text-slate-500">
              Target
              <select
                value={targetVersion ?? ""}
                onChange={(event) =>
                  onTargetVersionChange(Number(event.target.value) || null)
                }
                className="mt-1 w-full rounded border border-slate-200 px-2 py-1.5 text-xs"
              >
                <option value="">Pilih versi</option>
                {versions.map((version) => (
                  <option
                    key={`target-${version.id}`}
                    value={version.versionNo}
                  >
                    v{version.versionNo}
                  </option>
                ))}
              </select>
            </label>
            {isDiffLoading ? (
              <p className="text-xs text-slate-500">Menghitung diff…</p>
            ) : null}
            {diff ? (
              <div
                className="space-y-3 rounded bg-slate-50 p-3"
                data-testid="version-diff"
              >
                <p className="text-xs font-medium text-slate-700">
                  v{diff.baseVersion} → v{diff.targetVersion}
                </p>
                <DiffGroup label="Nodes" group={diff.nodes} />
                <DiffGroup label="Transitions" group={diff.transitions} />
                {diff.nodes.added.length === 0 &&
                diff.nodes.removed.length === 0 &&
                diff.nodes.changed.length === 0 &&
                diff.transitions.added.length === 0 &&
                diff.transitions.removed.length === 0 &&
                diff.transitions.changed.length === 0 ? (
                  <p className="text-xs text-slate-500">Tidak ada perubahan.</p>
                ) : null}
              </div>
            ) : null}
          </div>
        ) : null}
      </div>
    </aside>
  )
}
