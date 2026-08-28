import { tenantConfig } from "$/configs/tenant"
import { apiRouters, queryKeys } from "$/constants"
import { axios } from "$/services/axios"
import type { ErrorResponse } from "$/types/generals"
import type { UpdateWorkflowSchemaProps } from "@openstate/schemas"
import type { WorkflowResponse } from "@openstate/types"
import { pathVariable } from "@openstate/utils"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import type { AxiosError } from "axios"

type UpdateWorkflowVariables = {
  id: string
  projectId?: string
  payload: UpdateWorkflowSchemaProps
}

const updateWorkflow = async ({
  id,
  projectId,
  payload,
}: UpdateWorkflowVariables) => {
  const headers: Record<string, string> = {
    "X-Tenant-ID": tenantConfig.tenantId,
  }
  if (projectId) headers["X-Project-ID"] = projectId

  const result = await axios<WorkflowResponse>({
    method: "PATCH",
    url: pathVariable(apiRouters.workflows.update, { id }),
    headers,
    data: payload,
  })

  return result
}

const useUpdateWorkflow = () => {
  const queryClient = useQueryClient()

  const mutation = useMutation<
    WorkflowResponse,
    ErrorResponse<AxiosError>,
    UpdateWorkflowVariables
  >({
    mutationKey: [queryKeys.workflows.update],
    mutationFn: updateWorkflow,
    onSuccess: (data) => {
      void queryClient.invalidateQueries({
        queryKey: [queryKeys.workflows.list],
      })
      void queryClient.invalidateQueries({
        queryKey: [queryKeys.workflows.get, data.id],
      })
    },
  })

  return {
    ...mutation,
  }
}

export default useUpdateWorkflow
export { useUpdateWorkflow as useWorkflowsUpdate }
