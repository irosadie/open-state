import { usePathname, useRouter, useSearchParams } from "next/navigation"

const useQueryParam = () => {
  const searchParams = useSearchParams()
  const pathname = usePathname()
  const router = useRouter()

  const setQueryParams = (valuesToUpdate: Record<string, unknown>) => {
    const params = new URLSearchParams(searchParams)

    for (const [key, value] of Object.entries(valuesToUpdate)) {
      if (value !== undefined) {
        params.set(key, value as string)
      } else {
        params.delete(key)
      }
    }

    const queryString = params.toString()
    const updatedPath = queryString ? `${pathname}?${queryString}` : pathname

    router.push(updatedPath)
  }

  return {
    setQueryParams,
  }
}

export default useQueryParam
