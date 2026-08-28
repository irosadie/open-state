import { Suspense } from "react"
import CapabilityDetailPageContent from "./capability-detail-page-content"

export const metadata = {
  title: "Capability Detail",
  description: "View, edit, bind, and test a capability",
}

type CapabilityDetailPageProps = {
  params: Promise<{ id: string }>
}

export default async function CapabilityDetailPage({
  params,
}: CapabilityDetailPageProps) {
  const { id } = await params

  return (
    <Suspense fallback={null}>
      <CapabilityDetailPageContent id={id} />
    </Suspense>
  )
}
