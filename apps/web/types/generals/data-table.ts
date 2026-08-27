export type Meta = {
  pagination: {
    total: number
    currentPage: number
    perPage: number
    lastPage: number
  }
  cursor: {
    nextCursor: string | null
    hasMore: boolean
  }
}

export type DataTableResponse<T> = {
  list: T[]
  meta: Meta
}
