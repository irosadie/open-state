import { queryKeys } from "$/constants"
import type { ErrorResponse } from "$/types/generals"
import type { MCPToolOptionListResponse } from "@openstate/types"
import { useQuery } from "@tanstack/react-query"
import type { AxiosError } from "axios"
import { listProjectMCPToolOptions } from "./use-project-mcp-binding-common"

const useListProjectMCPToolOptions = (
  projectId: string | undefined,
  enabled = true,
) =>
  useQuery<MCPToolOptionListResponse, ErrorResponse<AxiosError>>({
    queryKey: [queryKeys.projectMCPBindings.options, projectId],
    queryFn: () => listProjectMCPToolOptions(projectId as string),
    enabled: enabled && Boolean(projectId),
  })

export default useListProjectMCPToolOptions
