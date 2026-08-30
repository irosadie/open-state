import { Suspense } from "react"
import RuntimeInstanceDetailPageContent from "./runtime-instance-detail-page-content"

export const metadata = {
  title: "Runtime Instance",
  description: "Inspect a persisted workflow runtime instance",
}

type RuntimeInstanceDetailPageProps = {
  params: Promise<{ id: string }>
}

export default async function RuntimeInstanceDetailPage({
  params,
}: RuntimeInstanceDetailPageProps) {
  const { id } = await params
  return (
    <Suspense fallback={null}>
      <RuntimeInstanceDetailPageContent id={id} />
    </Suspense>
  )
}
