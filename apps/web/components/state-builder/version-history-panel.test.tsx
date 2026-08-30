import { render, screen } from "@testing-library/react"
import { describe, expect, it, vi } from "vitest"
import { VersionHistoryPanel } from "./version-history-panel"

const baseProps = {
  isLoading: false,
  error: null,
  baseVersion: null,
  targetVersion: null,
  onBaseVersionChange: vi.fn(),
  onTargetVersionChange: vi.fn(),
  diff: undefined,
  isDiffLoading: false,
  onClose: vi.fn(),
}

describe("VersionHistoryPanel", () => {
  it("shows an explicit empty state when no versions are published", () => {
    render(<VersionHistoryPanel {...baseProps} versions={[]} />)

    expect(screen.getByText("Belum ada versi published.")).toBeTruthy()
    expect(screen.queryByText("Compare versions")).toBeNull()
  })

  it("renders version metadata and semantic diff groups", () => {
    render(
      <VersionHistoryPanel
        {...baseProps}
        baseVersion={1}
        targetVersion={2}
        versions={[
          {
            id: "v2",
            workflowId: "wf-1",
            versionNo: 2,
            status: "PUBLISHED",
            isCurrent: true,
            createdAt: "2026-01-02T00:00:00Z",
            updatedAt: "2026-01-02T00:00:00Z",
            definition: {},
          },
          {
            id: "v1",
            workflowId: "wf-1",
            versionNo: 1,
            status: "PUBLISHED",
            isCurrent: false,
            createdAt: "2026-01-01T00:00:00Z",
            updatedAt: "2026-01-01T00:00:00Z",
            definition: {},
          },
        ]}
        diff={{
          workflowId: "wf-1",
          baseVersion: 1,
          targetVersion: 2,
          nodes: {
            added: [{ id: "payment" }],
            removed: [{ id: "legacy" }],
            changed: [{ id: "start", changedFields: ["name", "policy"] }],
          },
          transitions: { added: [], removed: [], changed: [] },
        }}
      />,
    )

    expect(screen.getByText("v2 · current")).toBeTruthy()
    expect(screen.getAllByText("v1").length).toBeGreaterThan(0)
    expect(screen.getByText("+ payment")).toBeTruthy()
    expect(screen.getByText("− legacy")).toBeTruthy()
    expect(screen.getByText("~ start: name, policy")).toBeTruthy()
  })
})
