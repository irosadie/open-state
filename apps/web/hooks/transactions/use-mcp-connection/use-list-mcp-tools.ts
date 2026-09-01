import { queryKeys } from "$/constants"
import { useQuery } from "@tanstack/react-query"
import type { MCPToolCatalogError } from "./use-mcp-connection-common"
import { listMCPTools } from "./use-mcp-connection-common"

const useListMCPTools = (
  projectId: string | undefined,
  connectionId: string | undefined,
  enabled = true,
) =>
  useQuery({
    queryKey: [queryKeys.mcpConnections.tools, projectId, connectionId],
    queryFn: () => listMCPTools(projectId as string, connectionId as string),
    enabled: enabled && !!projectId && !!connectionId,
  }) as ReturnType<
    typeof useQuery<
      Awaited<ReturnType<typeof listMCPTools>>,
      MCPToolCatalogError
    >
  >

export default useListMCPTools
export { useListMCPTools }
