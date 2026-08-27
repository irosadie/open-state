import { PanelCard } from "$/components/panel-card"
import { type ColumnDef, Table } from "$/components/table"
import type { ExampleResponseProps } from "@vibecoding-starter/types/example-response"

type ExamplesTableCardProps = {
  data: ExampleResponseProps[]
  columns: ColumnDef<ExampleResponseProps>[]
  isLoading: boolean
  currentPage: number
  limit: number
  totalData: number
  onPageChange: (page: number) => void
}

export function ExamplesTableCard({
  data,
  columns,
  isLoading,
  currentPage,
  limit,
  totalData,
  onPageChange,
}: ExamplesTableCardProps) {
  return (
    <PanelCard>
      <Table
        data={data}
        columns={columns}
        intent="clean"
        isLoading={isLoading}
        isShowPagination={true}
        wrapperClassName="overflow-x-auto"
        thClassName="whitespace-nowrap"
        pagination={{
          page: currentPage,
          limit,
          totalData,
          onPageClick(newPage) {
            onPageChange(newPage)
          },
        }}
      />
    </PanelCard>
  )
}
