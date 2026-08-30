import { tenantConfig } from "$/configs/tenant"
import { apiRouters, queryKeys } from "$/constants"
import { axios } from "$/services/axios"
import type { WorkflowVersionResponse } from "@openstate/types"
import { pathVariable } from "@openstate/utils"
import { useQuery } from "@tanstack/react-query"

export function useGetWorkflowVersion(
  id: string,
  versionNo: number,
  projectId?: string,
) {
  return useQuery({
    queryKey: [queryKeys.workflows.version, id, versionNo, projectId],
    enabled: Boolean(id) && versionNo > 0,
    queryFn: () =>
      axios<WorkflowVersionResponse>({
        method: "GET",
        url: pathVariable(apiRouters.workflows.version, { id, versionNo }),
        headers: {
          "X-Tenant-ID": tenantConfig.tenantId,
          ...(projectId ? { "X-Project-ID": projectId } : {}),
        },
      }),
  })
}
