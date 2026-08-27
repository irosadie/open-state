// @ts-nocheck — for example files only. DO NOT use in production code.
import { QueryFunctionContext, useQuery } from '@tanstack/react-query'
import type { ExampleResponseProps as DataTypeProps } from '@vibecoding-starter/types/example-response'
import { axios } from '$/services/axios'
import { apiRouters, queryKeys } from '$/constants'
import { DataTableResponse, ErrorResponse } from '$/types/generals'
import { AxiosError } from 'axios'

const DEFAULT_LIMIT = 10

type DataTableFilterProps = {
	search?: string
	type?: string
	isActive?: boolean
}

type UseDataTableProps = {
	isAutoFetch?: boolean
	key?: string
	page?: number
	filter?: DataTableFilterProps
	limit?: number
}

type DataTableQueryKey = [
	string,
	{ page: number; limit: number; filter?: DataTableFilterProps },
]

const fetchDataTable = async (
	args: QueryFunctionContext<DataTableQueryKey>,
) => {
	const [, { page, limit, filter }] = args.queryKey

	const result = await axios<DataTableResponse<DataTypeProps>>({
		method: 'GET',
		url: apiRouters.examples.index,
		params: {
			...filter,
			search: filter?.search?.trim() || undefined,
			page: page,
			limit: limit,
		},
	})

	return result
}

const useDataTable = (args?: UseDataTableProps) => {
	const {
		key = queryKeys.examples.index,
		page = 1,
		filter,
		limit = DEFAULT_LIMIT,
		isAutoFetch,
	} = args || {}

	const dataTable = useQuery<
		DataTableResponse<DataTypeProps>,
		ErrorResponse<AxiosError>,
		DataTableResponse<DataTypeProps>,
		DataTableQueryKey
	>({
		queryKey: [key, { page, limit, filter }],
		enabled: isAutoFetch,
		queryFn: fetchDataTable,
	})

	return {
		limit: limit,
		refetch: dataTable.refetch,
		data: dataTable.data?.list,
		pagination: dataTable.data?.meta.pagination,
		error: dataTable.error,
		isLoading: dataTable.isLoading,
	}
}

export default useDataTable
export { useDataTable as useExamplesDataTable }
