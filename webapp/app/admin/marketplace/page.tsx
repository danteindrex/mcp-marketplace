'use client'

import { useEffect, useState } from 'react'
import { AppShell } from '@/components/app-shell'
import { Card } from '@/components/ui/card'
import { Text } from '@/components/retroui/Text'
import { fetchAdminMarketplaceInsights, type AdminMarketplaceInsight } from '@/lib/api-client'

export default function AdminMarketplacePage() {
  const [data, setData] = useState<{
    items: AdminMarketplaceInsight[]
    count: number
    totalInstalls: number
    popular: AdminMarketplaceInsight[]
  }>({ items: [], count: 0, totalInstalls: 0, popular: [] })

  useEffect(() => {
    fetchAdminMarketplaceInsights().then(setData).catch(() => setData({ items: [], count: 0, totalInstalls: 0, popular: [] }))
  }, [])

  return (
    <AppShell role="admin">
      <div className="p-6 space-y-6">
        <div>
          <Text variant="h3" className="mb-2">Marketplace Insights</Text>
          <Text variant="body" className="text-muted-foreground">See install traction, popular MCPs, and payment activity across the catalog.</Text>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <Card className="p-5"><Text variant="caption" className="text-muted-foreground">Tracked MCPs</Text><Text variant="h3">{data.count}</Text></Card>
          <Card className="p-5"><Text variant="caption" className="text-muted-foreground">Total Installs</Text><Text variant="h3">{data.totalInstalls}</Text></Card>
          <Card className="p-5"><Text variant="caption" className="text-muted-foreground">Popular MCPs</Text><Text variant="h3">{data.popular.length}</Text></Card>
        </div>

        <Card className="p-6">
          <Text variant="h5" className="mb-4">Top MCPs</Text>
          <div className="space-y-3">
            {data.popular.length === 0 ? <Text variant="small" className="text-muted-foreground">No install data yet.</Text> : null}
            {data.popular.map(item => (
              <div key={item.id} className="border border-border rounded-md p-3 grid grid-cols-1 md:grid-cols-5 gap-3 text-sm">
                <div><Text variant="caption" className="text-muted-foreground">MCP</Text><Text variant="small">{item.name}</Text></div>
                <div><Text variant="caption" className="text-muted-foreground">Installs</Text><Text variant="small">{item.installCount}</Text></div>
                <div><Text variant="caption" className="text-muted-foreground">x402 Intents</Text><Text variant="small">{item.x402IntentCount}</Text></div>
                <div><Text variant="caption" className="text-muted-foreground">Settled USDC</Text><Text variant="small">{Number(item.settledVolumeUsdc || 0).toFixed(2)}</Text></div>
                <div><Text variant="caption" className="text-muted-foreground">Status</Text><Text variant="small">{item.status}</Text></div>
              </div>
            ))}
          </div>
        </Card>

        <Card className="p-6">
          <Text variant="h5" className="mb-4">All MCP Install Metrics</Text>
          <div className="space-y-3">
            {data.items.length === 0 ? <Text variant="small" className="text-muted-foreground">No marketplace server metrics available.</Text> : null}
            {data.items.map(item => (
              <div key={item.id} className="border border-border rounded-md p-3 grid grid-cols-1 md:grid-cols-6 gap-3 text-sm">
                <div><Text variant="caption" className="text-muted-foreground">Name</Text><Text variant="small">{item.name}</Text></div>
                <div><Text variant="caption" className="text-muted-foreground">Slug</Text><Text variant="small">{item.slug}</Text></div>
                <div><Text variant="caption" className="text-muted-foreground">Installs</Text><Text variant="small">{item.installCount}</Text></div>
                <div><Text variant="caption" className="text-muted-foreground">Pricing</Text><Text variant="small">{item.pricingType || 'n/a'}</Text></div>
                <div><Text variant="caption" className="text-muted-foreground">Deployment</Text><Text variant="small">{item.deploymentStatus || 'n/a'}</Text></div>
                <div><Text variant="caption" className="text-muted-foreground">Settled USDC</Text><Text variant="small">{Number(item.settledVolumeUsdc || 0).toFixed(2)}</Text></div>
              </div>
            ))}
          </div>
        </Card>
      </div>
    </AppShell>
  )
}
