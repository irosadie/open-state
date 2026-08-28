import { tenantConfig } from "$/configs/tenant"
import { apiRouters, queryKeys } from "$/constants"
import { axios } from "$/services/axios"
import type { ErrorResponse } from "$/types/generals"
import type { WorkflowResponse } from "@openstate/types"
import { useQuery } from "@tanstack/react-query"
import type { AxiosError } from "axios"

type UseListWorkflowsArgs = {
  projectId?: string
  enabled?: boolean
}

type ListWorkflowsQueryKey = [string, { projectId?: string }]

const listWorkflows = async (projectId?: string) => {
  const result = await axios<WorkflowResponse[]>({
    method: "GET",
    url: apiRouters.workflows.index,
    headers: { "X-Tenant-ID": tenantConfig.tenantId },
    params: { projectId: projectId || undefined },
  })

  return result
}

const useListWorkflows = (args?: UseListWorkflowsArgs) => {
  const { projectId, enabled = true } = args || {}

  const query = useQuery<
    WorkflowResponse[],
    ErrorResponse<AxiosError>,
    WorkflowResponse[],
    ListWorkflowsQueryKey
  >({
    queryKey: [queryKeys.workflows.list, { projectId }],
    queryFn: () => listWorkflows(projectId),
    enabled,
  })

  return {
    ...query,
  }
}

export default useListWorkflows
export { useListWorkflows as useWorkflowsList }
