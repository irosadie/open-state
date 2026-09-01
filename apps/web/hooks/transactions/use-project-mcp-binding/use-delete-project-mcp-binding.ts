import { queryKeys } from "$/constants"
import type { ErrorResponse } from "$/types/generals"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import type { AxiosError } from "axios"
import {
  type ProjectMCPBindingMutation,
  deleteProjectMCPBinding,
} from "./use-project-mcp-binding-common"

const useDeleteProjectMCPBinding = () => {
  const queryClient = useQueryClient()
  return useMutation<
    void,
    ErrorResponse<AxiosError>,
    Pick<ProjectMCPBindingMutation, "projectId" | "capabilityId">
  >({
    mutationKey: [queryKeys.projectMCPBindings.delete],
    mutationFn: deleteProjectMCPBinding,
    onSuccess: (_data, variables) => {
      void queryClient.invalidateQueries({
        queryKey: [queryKeys.projectMCPBindings.list, variables.projectId],
      })
    },
  })
}

export default useDeleteProjectMCPBinding
