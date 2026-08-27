"use client"

import { useCallback, useEffect, useMemo, useState } from "react"
import type { ChangeEvent } from "react"

import { ActionsDropdown } from "$/components/actions-dropdown"
import { LoadingSpinner } from "$/components/loading-spinner"
import type { ColumnDef } from "$/components/table"
import {
	useExamplesDataTable,
	useExamplesDeleteOne,
	useExamplesGetOne,
	useExamplesInsertOne,
	useExamplesUpdateOne,
} from "$/hooks/transactions/use-examples"
import { useQueryParam } from "$/hooks/utility/use-query-param"
import {
	type ExampleSchemaProps,
	exampleSchema,
	getExampleLabel,
} from "$/schemas/example"
import type { ExampleResponseProps } from "@vibecoding-starter/types/example-response"
import { debounce } from "@vibecoding-starter/utils"
import { zodResolver } from "@hookform/resolvers/zod"
import { useSearchParams } from "next/navigation"
import { useForm } from "react-hook-form"
import { toast } from "react-hot-toast"
import Swal from "sweetalert2"
import { ExamplesFormDialog } from "./examples-form-dialog"
import { ExamplesTableCard } from "./examples-table-card"
import { ExamplesToolbar } from "./examples-toolbar"

const initialFormData: ExampleSchemaProps = {
	name: "",
	code: "",
	type: "EXAMPLE_TYPE_A",
	isActive: true,
}

export default function ExamplesPageContent() {
	const searchParams = useSearchParams()
	const currentPage = searchParams.get("page") || "1"
	const query = searchParams.get("q") || ""
	const editId = searchParams.get("edit")

	const [isDialogOpen, setIsDialogOpen] = useState(false)
	const [editingId, setEditingId] = useState<string | null>(null)
	const [tempQuery, setTempQuery] = useState<string>(query)

	const { setQueryParams } = useQueryParam()

	const {
		data: examples,
		isLoading,
		refetch,
		pagination: paginationData,
		limit,
	} = useExamplesDataTable({
		filter: {
			search: tempQuery,
		},
		page: Number(currentPage),
		isAutoFetch: true,
	})

	const { data: editData, isLoading: isEditLoading } = useExamplesGetOne({
		id: editId || "",
	})

	const { mutateAsync: insertMutateAsync, isPending: isInsertPending } =
		useExamplesInsertOne()
	const { mutateAsync: updateMutateAsync, isPending: isUpdatePending } =
		useExamplesUpdateOne()

	const isPending = isInsertPending || isUpdatePending
	const { mutate: deleteMutate } = useExamplesDeleteOne()

	const {
		register,
		handleSubmit,
		reset,
		formState: { errors, isValid },
	} = useForm<ExampleSchemaProps>({
		mode: "onChange",
		resolver: zodResolver(exampleSchema),
		defaultValues: initialFormData,
	})

	const handleOpenDialog = useCallback(
		(editing = false, data?: ExampleResponseProps) => {
			if (editing && data) {
				setEditingId(data.id)
				reset({
					name: data.name,
					code: data.code,
					type: data.type,
					isActive: data.isActive,
				})
			} else {
				setEditingId(null)
				reset(initialFormData)
			}
			setIsDialogOpen(true)
		},
		[reset],
	)

	const handleCloseDialog = () => {
		setIsDialogOpen(false)
		setEditingId(null)
		reset(initialFormData)
	}

	useEffect(() => {
		if (editData) {
			handleOpenDialog(true, editData)
		}
	}, [editData, handleOpenDialog])

	const onSubmit = async (data: ExampleSchemaProps) => {
		if (editingId) {
			updateMutateAsync(
				{ id: editingId, payload: data },
				{
					onSuccess: () => {
						toast.success("Example updated successfully")
					},
					onError: ({ message }) => {
						toast.error(
							message || "Failed to update example. Please try again.",
						)
					},
					onSettled: () => {
						handleCloseDialog()
						refetch()
					},
				},
			)
		} else {
			insertMutateAsync(data, {
				onSuccess: () => {
					toast.success("Example created successfully")
				},
				onError: ({ message }) => {
					toast.error(message || "Failed to create example. Please try again.")
				},
				onSettled: () => {
					handleCloseDialog()
					refetch()
				},
			})
		}
	}

	const debouncedSearch = useMemo(
		() =>
			debounce((value: string) => {
				setQueryParams({ q: value, page: 1 })
			}, 250),
		[setQueryParams],
	)

	const handleOnSearch = (e: ChangeEvent<HTMLInputElement>) => {
		const value = e.target.value

		setTempQuery(value)
		debouncedSearch(value)
	}

	useEffect(() => {
		return () => {
			debouncedSearch.cancel()
		}
	}, [debouncedSearch])

	const handleOnDelete = useCallback(
		(id: string) => {
			Swal.fire({
				text: "Apakah yakin menghapus data ini?",
				icon: "warning",
				showCancelButton: true,
				confirmButtonColor: "#e42c2c",
				cancelButtonColor: "#3278A0",
				confirmButtonText: "Ya, hapus!",
				cancelButtonText: "Batal",
				preConfirm: () => {
					Swal.showLoading()
					return new Promise((resolve, reject) => {
						deleteMutate(id, {
							onSuccess: () => resolve(null),
							onError: () => reject(),
						})
					})
				},
				didOpen: () => {
					const cancelButton = Swal.getCancelButton()

					cancelButton?.style.order = "-1"
				},
			})
				.then((result) => {
					if (result.isConfirmed) {
						toast.success("Data deleted successfully!")
						setQueryParams({ page: 1 })
						refetch()
					}
				})
				.catch(() => {
					Swal.hideLoading()
					Swal.close()
					toast.error("Failed to delete data. Please try again.")
				})
		},
		[deleteMutate, setQueryParams, refetch],
	)

	const columns: ColumnDef<ExampleResponseProps>[] = useMemo(
		() => [
			{
				accessorKey: "name",
				header: "Name",
				cell: (info) => (
					<span className="text-sm font-medium text-gray-900">
						{info.getValue() as string}
					</span>
				),
			},
			{
				accessorKey: "code",
				header: "Code",
				cell: (info) => (
					<span className="text-sm text-gray-600">
						{info.getValue() as string}
					</span>
				),
			},
			{
				accessorKey: "type",
				header: "Type",
				cell: (info) => (
					<span className="text-sm text-gray-600">
						{getExampleLabel(info.getValue() as ExampleResponseProps["type"])}
					</span>
				),
			},
			{
				accessorKey: "isActive",
				header: "Status",
				cell: (info) => {
					const isActive = info.getValue() as boolean

					return (
						<span
							className={`text-xs font-medium px-2 py-1 rounded ${isActive ? "bg-green-100 text-green-700" : "bg-gray-100 text-gray-600"}`}
						>
							{isActive ? "Active" : "Inactive"}
						</span>
					)
				},
			},
			{
				id: "actions",
				header: "",
				cell: (info) => (
					<ActionsDropdown
						actions={[
							{
								label: "Edit",
								onClick: () => handleOpenDialog(true, info.row.original),
							},
							{
								label: "Delete",
								onClick: () => handleOnDelete(info.row.original.id),
								destructive: true,
							},
						]}
					/>
				),
			},
		],
		[handleOpenDialog, handleOnDelete],
	)

	if (isEditLoading) {
		return <LoadingSpinner />
	}

	return (
		<div className="space-y-6 p-8">
			<ExamplesToolbar
				searchValue={tempQuery}
				onSearchChange={handleOnSearch}
				onCreate={() => handleOpenDialog(false)}
			/>

			<ExamplesTableCard
				data={examples || []}
				columns={columns}
				isLoading={isLoading}
				currentPage={Number(currentPage)}
				limit={limit}
				totalData={paginationData?.total || 0}
				onPageChange={(newPage) =>
					setQueryParams({
						page: newPage,
					})
				}
			/>

			<ExamplesFormDialog
				open={isDialogOpen}
				onOpenChange={setIsDialogOpen}
				editingId={editingId}
				register={register}
				errors={errors}
				isPending={isPending}
				isValid={isValid}
				onCancel={handleCloseDialog}
				onFormSubmit={(event) => {
					void handleSubmit(onSubmit)(event)
				}}
			/>
		</div>
	)
}
