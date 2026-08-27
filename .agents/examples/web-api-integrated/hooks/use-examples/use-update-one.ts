// @ts-nocheck — for example files only. DO NOT use in production code.
import { useMutation } from '@tanstack/react-query'
import { axios } from '$/services/axios'
import { apiRouters, queryKeys } from '$/constants'
import { ErrorResponse } from '$/types/generals'
import { AxiosError } from 'axios'
import type { ExampleSchemaProps as PayloadProps } from '$/schemas/example'
import type { ExampleResponseProps as ResponseProps } from '@vibecoding-starter/types/example-response'
import { pathVariable } from '$/utils/path-variable'

type UpdateParamsProps = {
	id: string
	payload: PayloadProps
}

const updateOne = async ({ id, payload }: UpdateParamsProps) => {
	const result = await axios<ResponseProps>({
		method: 'PUT',
		url: pathVariable(apiRouters.examples.update, { id }),
		data: payload,
	})

	return result
}

const useUpdateOne = () => {
	const mutation = useMutation<
		ResponseProps,
		ErrorResponse<AxiosError>,
		UpdateParamsProps,
		UpdateParamsProps
	>({
		mutationKey: [queryKeys.examples.update],
		mutationFn: updateOne,
	})

	return {
		...mutation,
	}
}

export default useUpdateOne
export { useUpdateOne as useExamplesUpdateOne }
