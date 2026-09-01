import { describe, expect, it } from "vitest"
import { projectListResponseSchema, projectResponseSchema } from "./project"

const project = {
  id: "project-1",
  tenantId: "tenant-1",
  name: "Padel",
  slug: "padel",
  status: "ACTIVE" as const,
  createdAt: "2026-08-31T00:00:00Z",
  updatedAt: "2026-08-31T00:00:00Z",
}

describe("project response schemas", () => {
  it("accepts a tenant project list", () => {
    expect(projectListResponseSchema.parse([project])).toEqual([project])
  })

  it("rejects an unsupported project status", () => {
    expect(() =>
      projectResponseSchema.parse({ ...project, status: "DELETED" }),
    ).toThrow()
  })
})
