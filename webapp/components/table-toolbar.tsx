'use client'

import { ReactNode, useState } from 'react'
import { X, Download, Search } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

export interface TableToolbarProps {
  searchPlaceholder?: string
  onSearch?: (value: string) => void
  filters?: Array<{
    name: string
    label: string
    options: Array<{ value: string; label: string }>
    onFilter?: (value: string) => void
  }>
  onExport?: () => void
  children?: ReactNode
  isLoading?: boolean
}

export function TableToolbar({
  searchPlaceholder = 'Search...',
  onSearch,
  filters = [],
  onExport,
  children,
  isLoading = false,
}: TableToolbarProps) {
  const [searchValue, setSearchValue] = useState('')
  const [activeFilters, setActiveFilters] = useState<Record<string, string>>({})

  const handleSearch = (value: string) => {
    setSearchValue(value)
    onSearch?.(value)
  }

  const handleFilter = (filterName: string, value: string) => {
    const newFilters = { ...activeFilters }
    if (value === 'all') {
      delete newFilters[filterName]
    } else {
      newFilters[filterName] = value
    }
    setActiveFilters(newFilters)

    const filter = filters.find(f => f.name === filterName)
    filter?.onFilter?.(value === 'all' ? '' : value)
  }

  const handleClearFilters = () => {
    setSearchValue('')
    setActiveFilters({})
    onSearch?.('')
    filters.forEach(f => f.onFilter?.(''))
  }

  const hasActiveFilters = Object.keys(activeFilters).length > 0 || searchValue !== ''

  return (
    <div className="space-y-4">
      <div className="flex flex-col sm:flex-row gap-4 items-start sm:items-center justify-between">
        <div className="flex flex-col sm:flex-row gap-2 w-full sm:w-auto flex-1">
          {/* Search Input */}
          <div className="relative flex-1 sm:max-w-xs">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground w-4 h-4" />
            <Input
              placeholder={searchPlaceholder}
              value={searchValue}
              onChange={e => handleSearch(e.target.value)}
              className="pl-10"
              disabled={isLoading}
              aria-label={searchPlaceholder}
            />
          </div>

          {/* Filters */}
          {filters.map(filter => (
            <Select
              key={filter.name}
              value={activeFilters[filter.name] || 'all'}
              onValueChange={value => handleFilter(filter.name, value)}
              disabled={isLoading}
            >
              <SelectTrigger className="w-full sm:w-40" aria-label={`Filter by ${filter.label}`}>
                <SelectValue placeholder={filter.label} />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All {filter.label}</SelectItem>
                {filter.options.map(option => (
                  <SelectItem key={option.value} value={option.value}>
                    {option.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          ))}
        </div>

        <div className="flex gap-2 w-full sm:w-auto">
          {hasActiveFilters && (
            <Button
              variant="outline"
              size="sm"
              onClick={handleClearFilters}
              disabled={isLoading}
              className="w-full sm:w-auto"
            >
              <X className="w-4 h-4 mr-2" />
              Clear
            </Button>
          )}

          {onExport && (
            <Button
              variant="outline"
              size="sm"
              onClick={onExport}
              disabled={isLoading}
              className="w-full sm:w-auto"
            >
              <Download className="w-4 h-4 mr-2" />
              Export
            </Button>
          )}
        </div>
      </div>

      {/* Custom Children (additional filters/actions) */}
      {children}
    </div>
  )
}
