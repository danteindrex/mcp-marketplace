'use client'

import { useState, useMemo } from 'react'
import Link from 'next/link'
import { LayoutGrid, List, Star } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { TableToolbar } from '@/components/table-toolbar'
import { mockServers } from '@/lib/mock-data'

type ViewMode = 'grid' | 'list'

const categoryOptions = [
  { value: 'data', label: 'Data' },
  { value: 'automation', label: 'Automation' },
  { value: 'ai', label: 'AI' },
  { value: 'integration', label: 'Integration' },
  { value: 'other', label: 'Other' },
]

const pricingOptions = [
  { value: 'free', label: 'Free' },
  { value: 'subscription', label: 'Subscription' },
  { value: 'flat', label: 'Flat Rate' },
  { value: 'x402', label: 'Per-Call (X402)' },
]

export default function MarketplacePage() {
  const [view, setView] = useState<ViewMode>('grid')
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedCategory, setSelectedCategory] = useState('')
  const [selectedPricing, setSelectedPricing] = useState('')
  const [sortBy, setSortBy] = useState('installs')

  // Filter and search servers
  const filteredServers = useMemo(() => {
    let results = mockServers

    // Search filter
    if (searchQuery) {
      const query = searchQuery.toLowerCase()
      results = results.filter(
        s =>
          s.name.toLowerCase().includes(query) ||
          s.description.toLowerCase().includes(query) ||
          s.author.toLowerCase().includes(query)
      )
    }

    // Category filter
    if (selectedCategory) {
      results = results.filter(s => s.category === selectedCategory)
    }

    // Pricing filter
    if (selectedPricing) {
      results = results.filter(s => s.pricingType === selectedPricing)
    }

    // Sort
    switch (sortBy) {
      case 'rating':
        results.sort((a, b) => b.rating - a.rating)
        break
      case 'installs':
        results.sort((a, b) => b.installCount - a.installCount)
        break
      case 'newest':
        results.sort((a, b) => b.lastUpdated.getTime() - a.lastUpdated.getTime())
        break
      case 'name':
        results.sort((a, b) => a.name.localeCompare(b.name))
        break
    }

    return results
  }, [searchQuery, selectedCategory, selectedPricing, sortBy])

  return (
    <div className="flex flex-col min-h-screen bg-background">
      {/* Page Header */}
      <div className="border-b border-border bg-background/50 backdrop-blur sticky top-0 z-40">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <h1 className="text-4xl font-bold mb-2">MCP Marketplace</h1>
          <p className="text-muted-foreground">Discover and install verified MCP servers</p>
        </div>
      </div>

      {/* Main Content */}
      <main className="flex-1 max-w-7xl mx-auto w-full px-4 sm:px-6 lg:px-8 py-8">
        <div className="space-y-6">
          {/* Toolbar */}
          <div className="space-y-4">
            <TableToolbar
              searchPlaceholder="Search servers by name, description, or author..."
              onSearch={setSearchQuery}
              filters={[
                {
                  name: 'category',
                  label: 'Category',
                  options: categoryOptions,
                  onFilter: setSelectedCategory,
                },
                {
                  name: 'pricing',
                  label: 'Pricing',
                  options: pricingOptions,
                  onFilter: setSelectedPricing,
                },
              ]}
              onExport={() => {
                const csv = filteredServers
                  .map(
                    s =>
                      `"${s.name}","${s.author}","${s.category}","${s.rating}","${s.installCount}"`
                  )
                  .join('\n')
                const blob = new Blob([csv], { type: 'text/csv' })
                const url = window.URL.createObjectURL(blob)
                const a = document.createElement('a')
                a.href = url
                a.download = 'servers.csv'
                a.click()
              }}
            />

            {/* Additional Controls */}
            <div className="flex flex-col sm:flex-row gap-4 items-start sm:items-center justify-between">
              <div className="flex items-center gap-2">
                <label className="text-sm text-muted-foreground">Sort by:</label>
                <select
                  value={sortBy}
                  onChange={e => setSortBy(e.target.value)}
                  className="px-3 py-2 rounded-md border border-input bg-background text-sm"
                >
                  <option value="installs">Most Installed</option>
                  <option value="rating">Highest Rated</option>
                  <option value="newest">Newest</option>
                  <option value="name">Name (A-Z)</option>
                </select>
              </div>

              <div className="flex items-center gap-2">
                <Button
                  variant={view === 'grid' ? 'default' : 'outline'}
                  size="sm"
                  onClick={() => setView('grid')}
                  aria-label="Grid view"
                >
                  <LayoutGrid className="w-4 h-4" />
                </Button>
                <Button
                  variant={view === 'list' ? 'default' : 'outline'}
                  size="sm"
                  onClick={() => setView('list')}
                  aria-label="List view"
                >
                  <List className="w-4 h-4" />
                </Button>
              </div>
            </div>
          </div>

          {/* Results */}
          <div className="text-sm text-muted-foreground">
            Showing {filteredServers.length} of {mockServers.length} servers
          </div>

          {/* Servers Grid */}
          {filteredServers.length === 0 ? (
            <div className="text-center py-16">
              <p className="text-muted-foreground mb-4">No servers found matching your criteria.</p>
              <Button
                variant="outline"
                onClick={() => {
                  setSearchQuery('')
                  setSelectedCategory('')
                  setSelectedPricing('')
                }}
              >
                Clear Filters
              </Button>
            </div>
          ) : view === 'grid' ? (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              {filteredServers.map(server => (
                <Link key={server.id} href={`/marketplace/${server.slug}`}>
                  <Card className="h-full p-6 hover:border-primary transition-colors cursor-pointer group flex flex-col">
                    <div className="flex items-start justify-between mb-4">
                      <div className="flex-1">
                        <h3 className="font-semibold group-hover:text-primary transition-colors line-clamp-1">
                          {server.name}
                        </h3>
                        <p className="text-sm text-muted-foreground">{server.author}</p>
                      </div>
                      {server.verified && (
                        <span className="ml-2 text-xs bg-green-500/20 text-green-700 dark:text-green-400 px-2 py-1 rounded whitespace-nowrap">
                          Verified
                        </span>
                      )}
                    </div>

                    <p className="text-sm text-muted-foreground mb-4 flex-1 line-clamp-2">{server.description}</p>

                    <div className="space-y-3 border-t border-border pt-3">
                      <div className="flex items-center justify-between text-sm">
                        <span className="text-muted-foreground">{server.category}</span>
                        <span className="font-medium">{server.pricingType === 'free' ? 'Free' : `$${server.pricingAmount || 'Custom'}`}</span>
                      </div>
                      <div className="flex items-center justify-between text-sm text-muted-foreground">
                        <span className="flex items-center gap-1">
                          <Star className="w-4 h-4 fill-amber-400 text-amber-400" />
                          {server.rating}
                        </span>
                        <span>{server.installCount.toLocaleString()} installs</span>
                      </div>
                    </div>
                  </Card>
                </Link>
              ))}
            </div>
          ) : (
            <div className="space-y-2">
              {filteredServers.map(server => (
                <Link key={server.id} href={`/marketplace/${server.slug}`}>
                  <Card className="p-4 hover:border-primary transition-colors cursor-pointer group">
                    <div className="flex items-center justify-between gap-4">
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2 mb-1">
                          <h3 className="font-semibold group-hover:text-primary transition-colors truncate">
                            {server.name}
                          </h3>
                          {server.verified && (
                            <span className="text-xs bg-green-500/20 text-green-700 dark:text-green-400 px-2 py-1 rounded whitespace-nowrap">
                              Verified
                            </span>
                          )}
                        </div>
                        <p className="text-sm text-muted-foreground truncate">{server.description}</p>
                      </div>
                      <div className="flex items-center gap-6 text-sm text-muted-foreground whitespace-nowrap">
                        <span className="hidden sm:inline">{server.author}</span>
                        <span className="flex items-center gap-1">
                          <Star className="w-4 h-4 fill-amber-400 text-amber-400" />
                          {server.rating}
                        </span>
                        <span className="hidden sm:inline">{server.installCount.toLocaleString()}</span>
                        <span className="font-medium">{server.pricingType === 'free' ? 'Free' : `$${server.pricingAmount || 'Custom'}`}</span>
                      </div>
                    </div>
                  </Card>
                </Link>
              ))}
            </div>
          )}
        </div>
      </main>
    </div>
  )
}
