import { tenantConfig } from "$/configs/tenant"
import { apiRouters, queryKeys } from "$/constants"
import { axios } from "$/services/axios"
import type { ErrorResponse } from "$/types/generals"
import type { WorkflowVersionResponse } from "@openstate/types"
import { pathVariable } from "@openstate/utils"
import { useQuery } from "@tanstack/react-query"
import type { AxiosError } from "axios"

type UseListWorkflowVersionsArgs = {
  id: string
  projectId?: string
  enabled?: boolean
}

const listWorkflowVersions = async (id: string, projectId?: string) => {
  const headers: Record<string, string> = {
    "X-Tenant-ID": tenantConfig.tenantId,
  }
  if (projectId) headers["X-Project-ID"] = projectId

  const result = await axios<WorkflowVersionResponse[]>({
    method: "GET",
    url: pathVariable(apiRouters.workflows.versions, { id }),
    headers,
  })

  return result
}

const useListWorkflowVersions = (args: UseListWorkflowVersionsArgs) => {
  const { id, projectId, enabled = true } = args

  const query = useQuery<
    WorkflowVersionResponse[],
    ErrorResponse<AxiosError>,
    WorkflowVersionResponse[],
    [string, string, string | undefined]
  >({
    queryKey: [queryKeys.workflows.versions, id, projectId],
    queryFn: () => listWorkflowVersions(id, projectId),
    enabled: enabled && !!id,
  })

  return {
    ...query,
  }
}

export default useListWorkflowVersions
export { useListWorkflowVersions as useWorkflowsVersions }
