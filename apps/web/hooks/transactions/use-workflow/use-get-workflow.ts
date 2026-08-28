import { tenantConfig } from "$/configs/tenant"
import { apiRouters, queryKeys } from "$/constants"
import { axios } from "$/services/axios"
import type { ErrorResponse } from "$/types/generals"
import type { WorkflowResponse } from "@openstate/types"
import { pathVariable } from "@openstate/utils"
import { useQuery } from "@tanstack/react-query"
import type { AxiosError } from "axios"

type UseGetWorkflowArgs = {
  id: string
  projectId?: string
  enabled?: boolean
}

const getWorkflow = async (id: string, projectId?: string) => {
  const headers: Record<string, string> = {
    "X-Tenant-ID": tenantConfig.tenantId,
  }
  if (projectId) headers["X-Project-ID"] = projectId

  const result = await axios<WorkflowResponse>({
    method: "GET",
    url: pathVariable(apiRouters.workflows.show, { id }),
    headers,
  })

  return result
}

const useGetWorkflow = (args: UseGetWorkflowArgs) => {
  const { id, projectId, enabled = true } = args

  const query = useQuery<
    WorkflowResponse,
    ErrorResponse<AxiosError>,
    WorkflowResponse,
    [string, string, string | undefined]
  >({
    queryKey: [queryKeys.workflows.get, id, projectId],
    queryFn: () => getWorkflow(id, projectId),
    enabled: enabled && !!id,
  })

  return {
    ...query,
  }
}

export default useGetWorkflow
export { useGetWorkflow as useWorkflowsGet }
