import { PanelCard } from "$/components/panel-card"

const tableSkeletonRows = [1, 2, 3, 4, 5, 6]

export function ExamplesPageLoading() {
	return (
		<div className="space-y-6 p-8">
			<div className="flex w-full items-center justify-between gap-4">
				<div className="h-12 w-full animate-pulse rounded-large bg-main-100" />
				<div className="h-12 w-32 animate-pulse rounded-large bg-main-100" />
			</div>

			<PanelCard>
				<div className="space-y-4 p-4">
					<div className="h-9 w-full animate-pulse rounded-large bg-main-100" />
					{tableSkeletonRows.map((row) => (
						<div
							key={row}
							className="h-6 w-full animate-pulse rounded bg-main-100"
						/>
					))}
				</div>
			</PanelCard>
		</div>
	)
}
