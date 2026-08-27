// @ts-nocheck — for example files only. DO NOT use in production code.
import { useMutation } from '@tanstack/react-query'
import { axios } from '$/services/axios'
import { apiRouters, queryKeys } from '$/constants'
import { ErrorResponse } from '$/types/generals'
import { AxiosError } from 'axios'
import { pathVariable } from '$/utils/path-variable'

const deleteOne = async (id: string) => {
	const result = await axios<unknown>({
		method: 'DELETE',
		url: pathVariable(apiRouters.examples.delete, { id }),
	})

	return result
}

const useDeleteOne = () => {
	const mutation = useMutation<
		unknown,
		ErrorResponse<AxiosError>,
		string,
		unknown
	>({
		mutationKey: [queryKeys.examples.delete],
		mutationFn: deleteOne,
	})

	return {
		...mutation,
		isLoading: mutation.isPending,
	}
}

export default useDeleteOne
export { useDeleteOne as useExamplesDeleteOne }
