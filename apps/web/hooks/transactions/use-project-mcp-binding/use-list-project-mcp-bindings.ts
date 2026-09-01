import { queryKeys } from "$/constants"
import type { ErrorResponse } from "$/types/generals"
import type { ProjectCapabilityMCPBindingListResponse } from "@openstate/types"
import { useQuery } from "@tanstack/react-query"
import type { AxiosError } from "axios"
import { listProjectMCPBindings } from "./use-project-mcp-binding-common"

const useListProjectMCPBindings = (
  projectId: string | undefined,
  enabled = true,
) =>
  useQuery<ProjectCapabilityMCPBindingListResponse, ErrorResponse<AxiosError>>({
    queryKey: [queryKeys.projectMCPBindings.list, projectId],
    queryFn: () => listProjectMCPBindings(projectId as string),
    enabled: enabled && Boolean(projectId),
  })

export default useListProjectMCPBindings
