// @ts-nocheck — for example files only. DO NOT use in production code.
import { useMutation } from '@tanstack/react-query'
import { axios } from '$/services/axios'
import { apiRouters, queryKeys } from '$/constants'
import { ErrorResponse } from '$/types/generals'
import { AxiosError } from 'axios'
import type { ExampleSchemaProps as PayloadProps } from '$/schemas/example'
import type { ExampleResponseProps as ResponseProps } from '@vibecoding-starter/types/example-response'

const insertOne = async (payload: PayloadProps) => {
	const result = await axios<ResponseProps>({
		method: 'POST',
		url: apiRouters.examples.insert,
		data: payload,
	})

	return result
}

const useInsertOne = () => {
	const mutation = useMutation<
		ResponseProps,
		ErrorResponse<AxiosError>,
		PayloadProps,
		unknown
	>({
		mutationKey: [queryKeys.examples.insert],
		mutationFn: insertOne,
	})

	return {
		...mutation,
	}
}

export default useInsertOne
export { useInsertOne as useExamplesInsertOne }
