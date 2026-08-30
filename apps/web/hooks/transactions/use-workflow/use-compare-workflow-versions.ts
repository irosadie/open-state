import { tenantConfig } from "$/configs/tenant"
import { apiRouters, queryKeys } from "$/constants"
import { axios } from "$/services/axios"
import type { WorkflowDiffResponse } from "@openstate/types"
import { pathVariable } from "@openstate/utils"
import { useQuery } from "@tanstack/react-query"

export function useCompareWorkflowVersions(
  id: string,
  baseVersion: number | null,
  targetVersion: number | null,
  projectId?: string,
) {
  return useQuery({
    queryKey: [
      queryKeys.workflows.compare,
      id,
      baseVersion,
      targetVersion,
      projectId,
    ],
    enabled:
      Boolean(id) &&
      baseVersion !== null &&
      targetVersion !== null &&
      baseVersion > 0 &&
      targetVersion > 0 &&
      baseVersion !== targetVersion,
    queryFn: () =>
      axios<WorkflowDiffResponse>({
        method: "GET",
        url: pathVariable(apiRouters.workflows.compare, { id }),
        headers: {
          "X-Tenant-ID": tenantConfig.tenantId,
          ...(projectId ? { "X-Project-ID": projectId } : {}),
        },
        params: { baseVersion, targetVersion },
      }),
  })
}
