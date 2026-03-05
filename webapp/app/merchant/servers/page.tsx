'use client'

import { useState, useMemo, useEffect } from 'react'
import Link from 'next/link'
import { Eye, Edit, Upload, Plus, TrendingUp, Users, Zap } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { AppShell } from '@/components/app-shell'
import { TableToolbar } from '@/components/table-toolbar'
import { fetchMerchantServers, type Server } from '@/lib/api-client'

const statusOptions = [
  { value: 'published', label: 'Published' },
  { value: 'draft', label: 'Draft' },
  { value: 'deprecated', label: 'Deprecated' },
]

export default function MerchantServersPage() {
  const [servers, setServers] = useState<Server[]>([])
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedStatus, setSelectedStatus] = useState('')
  const [sortBy, setSortBy] = useState('installs')

  useEffect(() => {
    fetchMerchantServers().then(items => setServers(items))
      .catch(() => {
        setServers([])
      })
  }, [])

  const filteredServers = useMemo(() => {
    let results = [...servers]

    if (searchQuery) {
      const query = searchQuery.toLowerCase()
      results = results.filter(
        s =>
          s.name.toLowerCase().includes(query) ||
          s.description.toLowerCase().includes(query)
      )
    }

    if (selectedStatus) {
      // Mock status - in real app would come from API
      const statusMap: Record<string, string[]> = {
        published: ['srv_001', 'srv_002', 'srv_003', 'srv_005'],
        draft: ['srv_004'],
        deprecated: [],
      }
      results = results.filter(s => statusMap[selectedStatus]?.includes(s.id))
    }

    switch (sortBy) {
      case 'revenue':
        results.sort((a, b) => (b.installCount * (b.pricingAmount || 0)) - (a.installCount * (a.pricingAmount || 0)))
        break
      case 'installs':
        results.sort((a, b) => b.installCount - a.installCount)
        break
      case 'rating':
        results.sort((a, b) => b.rating - a.rating)
        break
      case 'newest':
        results.sort((a, b) => new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime())
        break
    }

    return results
  }, [servers, searchQuery, selectedStatus, sortBy])

  return (
    <AppShell role="merchant">
      <div className="p-6 space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold mb-2">Your Servers</h1>
            <p className="text-muted-foreground">Manage and monitor your published MCP servers</p>
          </div>
          <Button asChild>
            <Link href="/merchant/servers/new/import-docker">
              <Plus className="w-4 h-4 mr-2" />
              New Server
            </Link>
          </Button>
        </div>

        {/* Stats */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <Card className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">Total Servers</p>
                <p className="text-3xl font-bold mt-1">{servers.length}</p>
              </div>
              <Zap className="w-8 h-8 text-primary opacity-50" />
            </div>
          </Card>
          <Card className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">Total Installs</p>
                <p className="text-3xl font-bold mt-1">
                  {servers.reduce((sum, s) => sum + s.installCount, 0).toLocaleString()}
                </p>
              </div>
              <Users className="w-8 h-8 text-primary opacity-50" />
            </div>
          </Card>
          <Card className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">Monthly Revenue</p>
                <p className="text-3xl font-bold mt-1">
                  $
                  {servers
                    .filter(s => s.pricingType !== 'free')
                    .reduce((sum, s) => sum + (s.installCount * (s.pricingAmount || 0) * 0.3), 0)
                    .toFixed(0)}
                </p>
              </div>
              <TrendingUp className="w-8 h-8 text-primary opacity-50" />
            </div>
          </Card>
        </div>

        {/* Toolbar */}
        <TableToolbar
          searchPlaceholder="Search servers by name..."
          onSearch={setSearchQuery}
          filters={[
            {
              name: 'status',
              label: 'Status',
              options: statusOptions,
              onFilter: setSelectedStatus,
            },
          ]}
        />

        {/* Additional Controls */}
        <div className="flex items-center justify-between">
          <label className="text-sm text-muted-foreground">
            Showing {filteredServers.length} server{filteredServers.length !== 1 ? 's' : ''}
          </label>
          <select
            value={sortBy}
            onChange={e => setSortBy(e.target.value)}
            className="px-3 py-2 rounded-md border border-input bg-background text-sm"
          >
            <option value="installs">Most Installed</option>
            <option value="revenue">Most Revenue</option>
            <option value="rating">Highest Rated</option>
            <option value="newest">Newest</option>
          </select>
        </div>

        {/* Servers Table */}
        <div className="space-y-2">
          {filteredServers.length === 0 ? (
            <Card className="p-8 text-center">
              <p className="text-muted-foreground mb-4">No servers found</p>
              <Button asChild>
                <Link href="/merchant/servers/new/import-docker">Create Your First Server</Link>
              </Button>
            </Card>
          ) : (
            filteredServers.map(server => {
              const status = ['srv_001', 'srv_002', 'srv_003', 'srv_005'].includes(server.id)
                ? 'published'
                : 'draft'
              const monthlyRevenue = server.pricingType !== 'free'
                ? Math.floor(server.installCount * (server.pricingAmount || 0) * 0.3)
                : 0

              return (
                <Card key={server.id} className="p-6 hover:border-primary transition-colors">
                  <div className="grid grid-cols-1 md:grid-cols-6 gap-4 items-center">
                    {/* Name & Status */}
                    <div className="md:col-span-2">
                      <h3 className="font-semibold mb-1">{server.name}</h3>
                      <div className="flex items-center gap-2">
                        <div
                          className={`w-2 h-2 rounded-full ${
                            status === 'published' ? 'bg-green-500' : 'bg-amber-500'
                          }`}
                        />
                        <span className="text-xs font-medium capitalize text-muted-foreground">
                          {status}
                        </span>
                      </div>
                    </div>

                    {/* Installs */}
                    <div>
                      <p className="text-sm text-muted-foreground mb-1">Installs</p>
                      <p className="font-semibold">{server.installCount.toLocaleString()}</p>
                    </div>

                    {/* Rating */}
                    <div>
                      <p className="text-sm text-muted-foreground mb-1">Rating</p>
                      <p className="font-semibold">⭐ {server.rating}</p>
                    </div>

                    {/* Revenue */}
                    <div>
                      <p className="text-sm text-muted-foreground mb-1">Est. Revenue</p>
                      <p className="font-semibold">${monthlyRevenue}</p>
                    </div>

                    {/* Actions */}
                    <div className="flex gap-2">
                      <Button
                        size="sm"
                        variant="outline"
                        asChild
                        title="View details"
                      >
                        <Link href={`/marketplace/${server.slug}`}>
                          <Eye className="w-4 h-4" />
                        </Link>
                      </Button>
                      <Button
                        size="sm"
                        variant="outline"
                        asChild
                        title="Edit server"
                      >
                        <Link href={`/merchant/servers/${server.id}/builder`}>
                          <Edit className="w-4 h-4" />
                        </Link>
                      </Button>
                    </div>
                  </div>
                </Card>
              )
            })
          )}
        </div>
      </div>
    </AppShell>
  )
}
