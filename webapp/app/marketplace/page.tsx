'use client'

import { useState, useMemo, useEffect } from 'react'
import Link from 'next/link'
import { LayoutGrid, List, Star } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { TableToolbar } from '@/components/table-toolbar'
import { fetchCurrentUser, fetchServers, type CurrentUser, type Server } from '@/lib/api-client'
import { Breadcrumb, Button as RetroButton } from '@/components/retroui'
import { LightModeOnly, DarkModeOnly } from '@/components/theme-aware'

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
  const [isLoading, setIsLoading] = useState(true)
  const [servers, setServers] = useState<Server[]>([])
  const [currentUser, setCurrentUser] = useState<CurrentUser | null>(null)
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedCategory, setSelectedCategory] = useState('')
  const [selectedPricing, setSelectedPricing] = useState('')
  const [sortBy, setSortBy] = useState('installs')

  useEffect(() => {
    setIsLoading(true)
    fetchServers().then(setServers)
      .catch(() => {
        setServers([])
      })
      .finally(() => {
        setIsLoading(false)
      })
  }, [])

  useEffect(() => {
    fetchCurrentUser().then(setCurrentUser)
  }, [])

  // Filter and search servers
  const filteredServers = useMemo(() => {
    let results = [...servers]

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
        results.sort((a, b) => new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime())
        break
      case 'name':
        results.sort((a, b) => a.name.localeCompare(b.name))
        break
    }

    return results
  }, [searchQuery, selectedCategory, selectedPricing, sortBy])

  const hasActiveFilters = Boolean(searchQuery || selectedCategory || selectedPricing)
  const resultSummary = isLoading
    ? 'Loading servers...'
    : hasActiveFilters
      ? `Showing ${filteredServers.length} results (${servers.length} total)`
      : `Showing ${servers.length} servers`
  const dashboardPath =
    currentUser?.role === 'admin'
      ? '/admin/tenants'
      : currentUser?.role === 'merchant'
        ? '/merchant/onboarding'
        : '/buyer/dashboard'

  return (
    <div className="flex flex-col min-h-screen grid-pattern bg-background">
      {/* Page Header */}
      <div className="border-b-2 border-foreground bg-background/90 backdrop-blur sticky top-0 z-40">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
            <div className="mb-4 flex items-center justify-between gap-3">
              <Breadcrumb>
                <Breadcrumb.List>
                <Breadcrumb.Item>
                  <Breadcrumb.Link asChild>
                    <Link href="/">Home</Link>
                  </Breadcrumb.Link>
                </Breadcrumb.Item>
                <Breadcrumb.Separator />
                <Breadcrumb.Item>
                  <Breadcrumb.Page>Marketplace</Breadcrumb.Page>
                </Breadcrumb.Item>
              </Breadcrumb.List>
            </Breadcrumb>
            <div className="flex items-center gap-2">
              {currentUser ? (
                <span className="hidden md:inline text-xs font-semibold text-muted-foreground">
                  Signed in as {currentUser.email}
                </span>
              ) : (
                <Button asChild size="sm" variant="outline" className="whitespace-nowrap">
                  <Link href="/login">Login</Link>
                </Button>
              )}
              <Button asChild size="sm" className="button-coral-solid whitespace-nowrap">
                <Link href={currentUser ? dashboardPath : '/'}>{currentUser ? 'Dashboard' : 'Return Home'}</Link>
              </Button>
            </div>
          </div>
          <h1 className="text-4xl font-black uppercase mb-2">MCP Marketplace</h1>
          <p className="text-muted-foreground font-medium">Discover and install verified MCP servers</p>
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
                <label className="text-sm text-muted-foreground font-semibold">Sort by:</label>
                <select
                  value={sortBy}
                  onChange={e => setSortBy(e.target.value)}
                  className="px-3 py-2 rounded-md border-2 border-foreground bg-background text-sm shadow-[3px_3px_0px_hsl(var(--shadow-color))]"
                >
                  <option value="installs">Most Installed</option>
                  <option value="rating">Highest Rated</option>
                  <option value="newest">Newest</option>
                  <option value="name">Name (A-Z)</option>
                </select>
              </div>

              <DarkModeOnly>
                <div className="flex items-center gap-2">
                  <Button
                    size="sm"
                    onClick={() => setView('grid')}
                    aria-label="Grid view"
                    className={view === 'grid' ? 'button-coral-solid' : 'border-2 border-foreground shadow-[3px_3px_0px_hsl(var(--shadow-color))]'}
                  >
                    <LayoutGrid className="w-4 h-4" />
                  </Button>
                  <Button
                    size="sm"
                    onClick={() => setView('list')}
                    aria-label="List view"
                    className={view === 'list' ? 'button-coral-solid' : 'border-2 border-foreground shadow-[3px_3px_0px_hsl(var(--shadow-color))]'}
                  >
                    <List className="w-4 h-4" />
                  </Button>
                </div>
              </DarkModeOnly>
              <LightModeOnly>
                <div className="flex items-center gap-2">
                  <RetroButton
                    size="sm"
                    onClick={() => setView('grid')}
                    aria-label="Grid view"
                    className={view === 'grid' ? 'button-coral-solid' : 'border-2 border-foreground shadow-[3px_3px_0px_hsl(var(--shadow-color))]'}
                  >
                    <LayoutGrid className="w-4 h-4" />
                  </RetroButton>
                  <RetroButton
                    size="sm"
                    onClick={() => setView('list')}
                    aria-label="List view"
                    className={view === 'list' ? 'button-coral-solid' : 'border-2 border-foreground shadow-[3px_3px_0px_hsl(var(--shadow-color))]'}
                  >
                    <List className="w-4 h-4" />
                  </RetroButton>
                </div>
              </LightModeOnly>
            </div>
          </div>

          {/* Results */}
          <div className="text-sm text-muted-foreground font-semibold">{resultSummary}</div>

          {/* Servers Grid */}
          {filteredServers.length === 0 ? (
            <div className="text-center py-16 brutal-surface">
              <p className="text-muted-foreground mb-4">No servers found matching your criteria.</p>
              <DarkModeOnly>
                <Button
                  onClick={() => {
                    setSearchQuery('')
                    setSelectedCategory('')
                    setSelectedPricing('')
                  }}
                  className="button-coral-solid"
                >
                  Clear Filters
                </Button>
              </DarkModeOnly>
              <LightModeOnly>
                <RetroButton
                  onClick={() => {
                    setSearchQuery('')
                    setSelectedCategory('')
                    setSelectedPricing('')
                  }}
                  className="button-coral-solid"
                >
                  Clear Filters
                </RetroButton>
              </LightModeOnly>
            </div>
          ) : view === 'grid' ? (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              {filteredServers.map(server => (
                <Link key={server.id} href={`/marketplace/${server.slug}`}>
                  <Card className="h-full p-6 border-2 border-foreground shadow-[4px_4px_0px_hsl(var(--shadow-color))] hover:translate-x-[-2px] hover:translate-y-[-2px] hover:shadow-[6px_6px_0px_hsl(var(--shadow-color))] transition-all cursor-pointer group flex flex-col">
                    <div className="flex items-start justify-between mb-4">
                      <div className="flex-1">
                        <h3 className="font-black uppercase group-hover:text-primary transition-colors line-clamp-1">
                          {server.name}
                        </h3>
                        <p className="text-sm text-muted-foreground font-semibold">{server.author}</p>
                      </div>
                      {server.verified && (
                        <span className="ml-2 text-xs bg-green-500/20 text-green-700 dark:text-green-400 px-2 py-1 rounded whitespace-nowrap">
                          Verified
                        </span>
                      )}
                    </div>

                    <p className="text-sm text-muted-foreground mb-4 flex-1 line-clamp-2 font-medium">{server.description}</p>

                    <div className="space-y-3 border-t-2 border-foreground/40 pt-3">
                      <div className="flex items-center justify-between text-sm font-semibold">
                        <span className="text-muted-foreground">{server.category}</span>
                        <span className="font-medium">{server.pricingType === 'free' ? 'Free' : `$${server.pricingAmount || 'Custom'}`}</span>
                      </div>
                      <div className="flex items-center justify-between text-sm text-muted-foreground font-semibold">
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
                  <Card className="p-4 border-2 border-foreground shadow-[4px_4px_0px_hsl(var(--shadow-color))] hover:translate-x-[-2px] hover:translate-y-[-2px] hover:shadow-[6px_6px_0px_hsl(var(--shadow-color))] transition-all cursor-pointer group">
                    <div className="flex items-center justify-between gap-4">
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2 mb-1">
                          <h3 className="font-black uppercase group-hover:text-primary transition-colors truncate">
                            {server.name}
                          </h3>
                          {server.verified && (
                            <span className="text-xs bg-green-500/20 text-green-700 dark:text-green-400 px-2 py-1 rounded whitespace-nowrap">
                              Verified
                            </span>
                          )}
                        </div>
                        <p className="text-sm text-muted-foreground truncate font-medium">{server.description}</p>
                      </div>
                      <div className="flex items-center gap-6 text-sm text-muted-foreground whitespace-nowrap font-semibold">
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
