import { queryKeys } from "$/constants"
import type { ErrorResponse } from "$/types/generals"
import { upsertProjectCapabilityMCPBindingSchema } from "@openstate/schemas"
import type { ProjectCapabilityMCPBindingResponse } from "@openstate/types"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import type { AxiosError } from "axios"
import {
  type ProjectMCPBindingMutation,
  upsertProjectMCPBinding,
} from "./use-project-mcp-binding-common"

const useUpsertProjectMCPBinding = () => {
  const queryClient = useQueryClient()
  return useMutation<
    ProjectCapabilityMCPBindingResponse,
    ErrorResponse<AxiosError>,
    ProjectMCPBindingMutation
  >({
    mutationKey: [queryKeys.projectMCPBindings.upsert],
    mutationFn: (variables) =>
      upsertProjectMCPBinding({
        ...variables,
        ...upsertProjectCapabilityMCPBindingSchema.parse({
          connectionId: variables.connectionId,
          toolId: variables.toolId,
        }),
      }),
    onSuccess: (_data, variables) => {
      void queryClient.invalidateQueries({
        queryKey: [queryKeys.projectMCPBindings.list, variables.projectId],
      })
    },
  })
}

export default useUpsertProjectMCPBinding
