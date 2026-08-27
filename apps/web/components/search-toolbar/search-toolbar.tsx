"use client"
import { Button } from "$/components/button"
import { Input } from "$/components/input"
import { Filter, Search } from "lucide-react"

interface SearchToolbarProps {
  searchQuery: string
  setSearchQuery: (query: string) => void
  searchPlaceholder?: string
  showFilterButton?: boolean
  onFilterClick?: () => void
}

export function SearchToolbar({
  searchQuery,
  setSearchQuery,
  searchPlaceholder = "Search...",
  showFilterButton = false,
  onFilterClick,
}: SearchToolbarProps) {
  return (
    <div className="flex flex-col sm:flex-row gap-3">
      <div className="flex-1">
        <Input
          placeholder={searchPlaceholder}
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          leftIcon={<Search className="w-4 h-4" />}
        />
      </div>

      {showFilterButton && (
        <Button intent="secondary" rounded="medium" onClick={onFilterClick}>
          <Filter className="w-4 h-4 mr-2" />
          Filter
        </Button>
      )}
    </div>
  )
}
