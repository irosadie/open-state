import type { ChangeEvent } from "react"

import { Plus, Search } from "lucide-react"

import { Button } from "$/components/button"
import { Input } from "$/components/input"

type ExamplesToolbarProps = {
	searchValue: string
	onSearchChange: (event: ChangeEvent<HTMLInputElement>) => void
	onCreate: () => void
}

export function ExamplesToolbar({
	searchValue,
	onSearchChange,
	onCreate,
}: ExamplesToolbarProps) {
	return (
		<div className="flex items-center justify-between gap-4 w-full">
			<div className="w-full">
				<Input
					intent="clean"
					rounded="large"
					leftIcon={<Search className="h-5 w-5 text-gray-400" />}
					placeholder="Search examples..."
					onChange={onSearchChange}
					value={searchValue}
				/>
			</div>
			<Button
				intent="primary"
				size="medium"
				leftIcon={<Plus className="h-4 w-4" />}
				onClick={onCreate}
				className="whitespace-nowrap"
				rounded="large"
			>
				Add New
			</Button>
		</div>
	)
}
