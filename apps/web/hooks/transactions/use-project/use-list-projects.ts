import { tenantConfig } from "$/configs/tenant"
import { apiRouters, queryKeys } from "$/constants"
import { axios } from "$/services/axios"
import type { ErrorResponse } from "$/types/generals"
import { projectListResponseSchema } from "@openstate/schemas"
import type { ProjectListResponse } from "@openstate/types"
import { useQuery } from "@tanstack/react-query"
import type { AxiosError } from "axios"

type UseListProjectsArgs = {
  enabled?: boolean
}

const listProjects = async () => {
  const result = await axios<ProjectListResponse>({
    method: "GET",
    url: apiRouters.projects.index,
    headers: { "X-Tenant-ID": tenantConfig.tenantId },
  })

  return projectListResponseSchema.parse(result)
}

const useListProjects = (args?: UseListProjectsArgs) => {
  const { enabled = true } = args || {}

  return useQuery<ProjectListResponse, ErrorResponse<AxiosError>>({
    queryKey: [queryKeys.projects.list, tenantConfig.tenantId],
    queryFn: listProjects,
    enabled,
  })
}

export default useListProjects
export { useListProjects as useProjectsList }
