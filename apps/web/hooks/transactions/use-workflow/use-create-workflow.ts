import { tenantConfig } from "$/configs/tenant"
import { apiRouters, queryKeys } from "$/constants"
import { axios } from "$/services/axios"
import type { ErrorResponse } from "$/types/generals"
import type { CreateWorkflowSchemaProps } from "@openstate/schemas"
import type { WorkflowResponse } from "@openstate/types"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import type { AxiosError } from "axios"

const createWorkflow = async (payload: CreateWorkflowSchemaProps) => {
  const result = await axios<WorkflowResponse>({
    method: "POST",
    url: apiRouters.workflows.create,
    headers: { "X-Tenant-ID": tenantConfig.tenantId },
    data: payload,
  })

  return result
}

const useCreateWorkflow = () => {
  const queryClient = useQueryClient()

  const mutation = useMutation<
    WorkflowResponse,
    ErrorResponse<AxiosError>,
    CreateWorkflowSchemaProps
  >({
    mutationKey: [queryKeys.workflows.create],
    mutationFn: createWorkflow,
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

export default useCreateWorkflow
export { useCreateWorkflow as useWorkflowsCreate }
