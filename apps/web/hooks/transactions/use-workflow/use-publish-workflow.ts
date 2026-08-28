import { tenantConfig } from "$/configs/tenant"
import { apiRouters, queryKeys } from "$/constants"
import { axios } from "$/services/axios"
import type { ErrorResponse } from "$/types/generals"
import type { PublishWorkflowSchemaProps } from "@openstate/schemas"
import type { WorkflowVersionResponse } from "@openstate/types"
import { pathVariable } from "@openstate/utils"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import type { AxiosError } from "axios"

type PublishWorkflowVariables = {
  id: string
  projectId?: string
  payload: PublishWorkflowSchemaProps
}

const publishWorkflow = async ({
  id,
  projectId,
  payload,
}: PublishWorkflowVariables) => {
  const headers: Record<string, string> = {
    "X-Tenant-ID": tenantConfig.tenantId,
  }
  if (projectId) headers["X-Project-ID"] = projectId

  const result = await axios<WorkflowVersionResponse>({
    method: "POST",
    url: pathVariable(apiRouters.workflows.publish, { id }),
    headers,
    data: payload,
  })

  return result
}

const usePublishWorkflow = () => {
  const queryClient = useQueryClient()

  const mutation = useMutation<
    WorkflowVersionResponse,
    ErrorResponse<AxiosError>,
    PublishWorkflowVariables
  >({
    mutationKey: [queryKeys.workflows.publish],
    mutationFn: publishWorkflow,
    onSuccess: (data) => {
      void queryClient.invalidateQueries({
        queryKey: [queryKeys.workflows.list],
      })
      void queryClient.invalidateQueries({
        queryKey: [queryKeys.workflows.get, data.workflowId],
      })
      void queryClient.invalidateQueries({
        queryKey: [queryKeys.workflows.versions, data.workflowId],
      })
    },
  })

  return {
    ...mutation,
  }
}

export default usePublishWorkflow
export { usePublishWorkflow as useWorkflowsPublish }
