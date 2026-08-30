import { z } from "zod"

export const simulationOutcomes = [
  "ENTERED",
  "TRANSITIONED",
  "REJECTED",
] as const

export const simulationOutcomeLabels = [
  { label: "Entered", value: "ENTERED" },
  { label: "Transitioned", value: "TRANSITIONED" },
  { label: "Rejected", value: "REJECTED" },
]

export const getSimulationOutcomeLabel = (
  value: (typeof simulationOutcomes)[number],
) => {
  return (
    simulationOutcomeLabels.find((label) => label.value === value)?.label ??
    value
  )
}

export const simulationFinalStatuses = ["RUNNING", "COMPLETED"] as const

export const simulationFinalStatusLabels = [
  { label: "Running", value: "RUNNING" },
  { label: "Completed", value: "COMPLETED" },
]

export const getSimulationFinalStatusLabel = (
  value: (typeof simulationFinalStatuses)[number],
) => {
  return (
    simulationFinalStatusLabels.find((label) => label.value === value)?.label ??
    value
  )
}

export const simulationEventSchema = z.object({
  type: z.string().trim().min(1, "Event type is required"),
  payload: z.record(z.string(), z.unknown()).optional().default({}),
})

export const simulationWorkflowSchema = z.object({
  definition: z.record(z.string(), z.unknown()),
  initialContext: z.record(z.string(), z.unknown()).optional().default({}),
  events: z.array(simulationEventSchema).max(100).default([]),
})

export type SimulationEventSchemaProps = z.infer<typeof simulationEventSchema>
export type SimulationWorkflowSchemaProps = z.infer<
  typeof simulationWorkflowSchema
>
