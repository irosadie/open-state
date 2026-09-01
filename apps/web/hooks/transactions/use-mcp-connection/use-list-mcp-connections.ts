import { queryKeys } from "$/constants"
import { useQuery } from "@tanstack/react-query"
import type { MCPConnectionError } from "./use-mcp-connection-common"
import { listMCPConnections } from "./use-mcp-connection-common"

const useListMCPConnections = (projectId: string | undefined, enabled = true) =>
  useQuery({
    queryKey: [queryKeys.mcpConnections.list, projectId],
    queryFn: () => listMCPConnections(projectId as string),
    enabled: enabled && !!projectId,
  }) as ReturnType<
    typeof useQuery<
      Awaited<ReturnType<typeof listMCPConnections>>,
      MCPConnectionError
    >
  >

export default useListMCPConnections
