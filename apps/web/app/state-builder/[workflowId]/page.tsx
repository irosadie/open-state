import { StateBuilder } from "$/components/state-builder"

interface StateBuilderWorkflowPageProps {
  params: Promise<{ workflowId: string }>
}

export default async function StateBuilderWorkflowPage({
  params,
}: StateBuilderWorkflowPageProps) {
  const { workflowId } = await params
  return (
    <main className="h-screen w-screen overflow-hidden">
      <StateBuilder workflowId={workflowId} />
    </main>
  )
}
