// @ts-nocheck — for example files only. DO NOT use in production code.
import { useQuery } from '@tanstack/react-query'
import { axios } from '$/services/axios'
import { apiRouters, queryKeys } from '$/constants'
import { ErrorResponse } from '$/types/generals'
import { AxiosError } from 'axios'
import type { ExampleResponseProps as ResponseProps } from '@vibecoding-starter/types/example-response'
import { pathVariable } from '$/utils/path-variable'

type UseGetOneProps = {
	id: string
}

const getOne = async (id: string) => {
	const result = await axios<ResponseProps>({
		method: 'GET',
		url: pathVariable(apiRouters.examples.show, { id }),
	})

	return result
}

const useGetOne = ({ id }: UseGetOneProps) => {
	const query = useQuery<
		ResponseProps,
		ErrorResponse<AxiosError>,
		ResponseProps,
		[string, string]
	>({
		queryKey: [queryKeys.examples.get, id],
		queryFn: () => getOne(id),
		enabled: !!id,
	})

	return {
		...query,
	}
}

export default useGetOne
export { useGetOne as useExamplesGetOne }
