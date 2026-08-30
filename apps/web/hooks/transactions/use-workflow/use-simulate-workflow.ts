import { tenantConfig } from "$/configs/tenant"
import { apiRouters, queryKeys } from "$/constants"
import { axios } from "$/services/axios"
import type { ErrorResponse } from "$/types/generals"
import { simulationWorkflowSchema } from "@openstate/schemas"
import type { SimulationResultResponse } from "@openstate/types"
import { useMutation } from "@tanstack/react-query"
import type { AxiosError } from "axios"

export type SimulateWorkflowPayload = {
  definition: Record<string, unknown>
  initialContext?: Record<string, unknown>
  events?: Array<{
    type: string
    payload?: Record<string, unknown>
  }>
}

const simulateWorkflow = async (payload: SimulateWorkflowPayload) => {
  const validated = simulationWorkflowSchema.parse(payload)
  return axios<SimulationResultResponse>({
    method: "POST",
    url: apiRouters.workflows.simulate,
    headers: { "X-Tenant-ID": tenantConfig.tenantId },
    data: validated,
  })
}

const useSimulateWorkflow = () => {
  const mutation = useMutation<
    SimulationResultResponse,
    ErrorResponse<AxiosError>,
    SimulateWorkflowPayload
  >({
    mutationKey: [queryKeys.workflows.simulate],
    mutationFn: simulateWorkflow,
  })

  return {
    ...mutation,
  }
}

export default useSimulateWorkflow
export { useSimulateWorkflow as useWorkflowsSimulate }
