import { Suspense } from "react"
import ExamplesPageContent from "./_components/examples-page-content"
import { ExamplesPageLoading } from "./_components/examples-page-loading"

export default function ExamplesPage() {
	return (
		<Suspense fallback={<ExamplesPageLoading />}>
			<ExamplesPageContent />
		</Suspense>
	)
}
