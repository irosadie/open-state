import type { FormEventHandler } from "react"

import type { FieldErrors, UseFormRegister } from "react-hook-form"

import { Button } from "$/components/button"
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from "$/components/dialog"
import { Input } from "$/components/input"
import { type ExampleSchemaProps, exampleLabels } from "$/schemas/example"

type ExamplesFormDialogProps = {
	open: boolean
	editingId: string | null
	register: UseFormRegister<ExampleSchemaProps>
	errors: FieldErrors<ExampleSchemaProps>
	isPending: boolean
	isValid: boolean
	onOpenChange: (open: boolean) => void
	onCancel: () => void
	onFormSubmit: FormEventHandler<HTMLFormElement>
}

export function ExamplesFormDialog({
	open,
	editingId,
	register,
	errors,
	isPending,
	isValid,
	onOpenChange,
	onCancel,
	onFormSubmit,
}: ExamplesFormDialogProps) {
	return (
		<Dialog open={open} onOpenChange={onOpenChange}>
			<DialogContent className="min-w-[40%] gap-0">
				<DialogHeader>
					<DialogTitle>
						{editingId ? "Edit Example" : "Add Example"}
					</DialogTitle>
					<DialogDescription>
						{editingId
							? "Update example details below."
							: "Fill in the example details below."}
					</DialogDescription>
				</DialogHeader>
				<form onSubmit={onFormSubmit} className="space-y-4">
					<Input
						label="Name"
						placeholder="Example name"
						required
						error={errors.name?.message}
						{...register("name")}
						rounded="large"
						intent="clean"
					/>
					<Input
						label="Code"
						placeholder="EXM-001"
						required
						error={errors.code?.message}
						{...register("code")}
						rounded="large"
						intent="clean"
					/>
					<div className="space-y-1.5">
						<label htmlFor="example-type" className="text-sm text-main-500">
							Type<span className="text-danger-500">*</span>
						</label>
						<select
							id="example-type"
							{...register("type")}
							className="h-12 w-full rounded-large border border-main-100 bg-white px-3 text-sm text-main-500"
						>
							{exampleLabels.map((item) => (
								<option key={item.value} value={item.value}>
									{item.label}
								</option>
							))}
						</select>
						{errors.type?.message ? (
							<span className="text-xs text-danger-500">
								{errors.type.message}
							</span>
						) : null}
					</div>
					<label className="flex items-center gap-2 text-sm text-main-500">
						<input type="checkbox" {...register("isActive")} />
						Active
					</label>
					<DialogFooter>
						<Button
							type="button"
							intent="warning"
							rounded="large"
							textOnly
							onClick={onCancel}
							disabled={isPending}
						>
							Cancel
						</Button>
						<Button
							type="submit"
							intent="primary"
							rounded="large"
							loading={isPending}
							disabled={!isValid || isPending}
						>
							{editingId ? "Update" : "Save"}
						</Button>
					</DialogFooter>
				</form>
			</DialogContent>
		</Dialog>
	)
}
