import { tenantConfig } from "$/configs/tenant"
import { apiRouters, queryKeys } from "$/constants"
import { axios } from "$/services/axios"
import type { ErrorResponse } from "$/types/generals"
import { instanceResponseSchema } from "@openstate/schemas"
import type { InstanceResponse } from "@openstate/types"
import { pathVariable } from "@openstate/utils"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import type { AxiosError } from "axios"

type InstanceCommand = "suspend" | "resume" | "retry"

const commandRouters: Record<
  InstanceCommand,
  (typeof apiRouters.admin)["suspendInstance"]
> = {
  suspend: apiRouters.admin.suspendInstance,
  resume: apiRouters.admin.resumeInstance,
  retry: apiRouters.admin.retryInstance,
}

const commandKeys: Record<InstanceCommand, string> = {
  suspend: queryKeys.admin.suspendInstance,
  resume: queryKeys.admin.resumeInstance,
  retry: queryKeys.admin.retryInstance,
}

type InstanceCommandArgs = { id: string; command: InstanceCommand }

const runInstanceCommand = async ({
  id,
  command,
}: InstanceCommandArgs): Promise<InstanceResponse> => {
  const result = await axios<InstanceResponse>({
    method: "POST",
    url: pathVariable(commandRouters[command], { id }),
    headers: { "X-Tenant-ID": tenantConfig.tenantId },
  })
  return instanceResponseSchema.parse(result)
}

const useInstanceCommand = (command: InstanceCommand) => {
  const queryClient = useQueryClient()
  return useMutation<
    InstanceResponse,
    ErrorResponse<AxiosError>,
    Omit<InstanceCommandArgs, "command">
  >({
    mutationKey: [commandKeys[command]],
    mutationFn: (args) => runInstanceCommand({ ...args, command }),
    onSuccess: (_data, variables) => {
      void queryClient.invalidateQueries({
        queryKey: [queryKeys.admin.instances],
      })
      void queryClient.invalidateQueries({
        queryKey: [queryKeys.runtimeInstances.list],
      })
      void queryClient.invalidateQueries({
        queryKey: [queryKeys.runtimeInstances.get, variables.id],
      })
      void queryClient.invalidateQueries({ queryKey: [queryKeys.admin.events] })
    },
  })
}

export default useInstanceCommand
export { useInstanceCommand as useAdminInstanceCommand }
