export type ProjectStatus = "ACTIVE" | "ARCHIVED"

export type ProjectResponse = {
  id: string
  tenantId: string
  name: string
  slug: string
  status: ProjectStatus
  createdAt: string
  updatedAt: string
}

export type ProjectListResponse = ProjectResponse[]
