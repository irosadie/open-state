import { StateBuilder } from "$/components/state-builder"

interface StateBuilderWorkflowPageProps {
  params: Promise<{ workflowId: string }>
  searchParams: Promise<{ projectId?: string | string[] }>
}

export default async function StateBuilderWorkflowPage({
  params,
  searchParams,
}: StateBuilderWorkflowPageProps) {
  const { workflowId } = await params
  const query = await searchParams
  const projectId = Array.isArray(query.projectId)
    ? query.projectId[0]
    : query.projectId
  return (
    <main className="h-screen w-screen overflow-hidden">
      <StateBuilder workflowId={workflowId} projectId={projectId} />
    </main>
  )
}
