import { StateBuilder } from "$/components/state-builder"

export const metadata = {
  title: "State Builder",
  description: "Visual workflow & conversation state builder",
}

type StateBuilderPageProps = {
  searchParams: Promise<{ projectId?: string | string[] }>
}

export default async function StateBuilderPage({
  searchParams,
}: StateBuilderPageProps) {
  const query = await searchParams
  const projectId = Array.isArray(query.projectId)
    ? query.projectId[0]
    : query.projectId

  return (
    <main className="h-screen w-screen overflow-hidden">
      <StateBuilder projectId={projectId} />
    </main>
  )
}
