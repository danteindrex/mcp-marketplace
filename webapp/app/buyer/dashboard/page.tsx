'use client'

import { useEffect, useState } from 'react'
import Link from 'next/link'
import { AlertCircle, TrendingUp, Zap, Clock, DollarSign, Plus } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { AppShell } from '@/components/app-shell'
import { LoadingState } from '@/components/empty-state'
import { fetchConnections, fetchBilling } from '@/lib/api-client'
import { Text } from '@/components/retroui/Text'
import { PieChart } from '@/components/retroui/charts/PieChart'
import { LineChart } from '@/components/retroui/charts/LineChart'

export default function BuyerDashboardPage() {
  const [isLoading, setIsLoading] = useState(true)
  const [connections, setConnections] = useState<any[]>([])
  const [billing, setBilling] = useState<any>({ monthlySpend: 0, currentBalance: 0, plan: 'free' })

  useEffect(() => {
    const loadData = async () => {
      setIsLoading(true)
      try {
        const [conns, bill] = await Promise.all([
          fetchConnections(),
          fetchBilling(),
        ])
        setConnections(conns as any)
        setBilling(bill as any)
      } finally {
        setIsLoading(false)
      }
    }

    loadData()
  }, [])

  if (isLoading) {
    return (
      <AppShell role="buyer">
        <LoadingState />
      </AppShell>
    )
  }

  const activeConnections = connections.filter(c => c.status === 'active').length
  const expiredConnections = connections.filter(c => c.status === 'expired').length
  const connectionStatus = [
    { name: 'Active', value: activeConnections },
    { name: 'Expired', value: expiredConnections },
    { name: 'Other', value: Math.max(0, connections.length - activeConnections - expiredConnections) },
  ].filter(item => item.value > 0)
  const usageTrend = connections.slice(0, 6).map((conn, idx) => ({
    name: `W${idx + 1}`,
    requests: conn.scopes.length * (idx + 1) * 3,
    spend: Number((billing.monthlySpend / Math.max(1, idx + 1)).toFixed(2)),
  }))

  return (
    <AppShell role="buyer">
      <div className="p-6 space-y-8">
        {/* Page Header */}
        <div>
          <Text variant="h3" className="mb-2">Dashboard</Text>
          <Text variant="body" className="text-muted-foreground">Welcome back! Here's an overview of your MCP servers and usage.</Text>
        </div>

        {/* Stats Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <Card className="p-6">
            <div className="flex items-center justify-between mb-2">
              <Text variant="small" className="text-muted-foreground">Active Connections</Text>
              <Zap className="w-4 h-4 text-primary" />
            </div>
            <Text variant="h3">{activeConnections}</Text>
            <Text variant="caption" className="text-muted-foreground mt-1">
              {expiredConnections > 0 ? `${expiredConnections} need renewal` : 'All active'}
            </Text>
          </Card>

          <Card className="p-6">
            <div className="flex items-center justify-between mb-2">
              <Text variant="small" className="text-muted-foreground">Monthly Spend</Text>
              <DollarSign className="w-4 h-4 text-primary" />
            </div>
            <Text variant="h3">${billing.monthlySpend.toFixed(2)}</Text>
            <Text variant="caption" className="text-muted-foreground mt-1">Billing cycle</Text>
          </Card>

          <Card className="p-6">
            <div className="flex items-center justify-between mb-2">
              <Text variant="small" className="text-muted-foreground">Account Balance</Text>
              <TrendingUp className="w-4 h-4 text-primary" />
            </div>
            <Text variant="h3">${billing.currentBalance.toFixed(2)}</Text>
            <Text variant="caption" className="text-muted-foreground mt-1">Available credit</Text>
          </Card>

          <Card className="p-6">
            <div className="flex items-center justify-between mb-2">
              <Text variant="small" className="text-muted-foreground">Plan</Text>
              <Clock className="w-4 h-4 text-primary" />
            </div>
            <Text variant="h3" className="capitalize">{billing.plan}</Text>
            <Text variant="caption" className="text-muted-foreground mt-1">Current tier</Text>
          </Card>
        </div>

        {/* Alerts */}
        {expiredConnections > 0 && (
          <Card className="bg-amber-500/10 border-amber-200 dark:border-amber-800 p-4 flex items-start gap-4">
            <AlertCircle className="w-5 h-5 text-amber-600 dark:text-amber-500 flex-shrink-0 mt-0.5" />
            <div>
              <Text variant="small" className="text-amber-900 dark:text-amber-100">
                {expiredConnections} Connection{expiredConnections > 1 ? 's' : ''} Expired
              </Text>
              <Text variant="small" className="text-amber-800 dark:text-amber-200 mt-1">
                Some of your tokens have expired. Please renew them to maintain access.
              </Text>
              <Button
                size="sm"
                variant="outline"
                className="mt-3 bg-amber-50 dark:bg-amber-950 hover:bg-amber-100 dark:hover:bg-amber-900 text-amber-900 dark:text-amber-100 border-amber-200 dark:border-amber-800"
                asChild
              >
                <Link href="/buyer/connections">Manage Connections</Link>
              </Button>
            </div>
          </Card>
        )}

        {/* Recent Connections */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
          <Card className="p-6">
            <Text variant="h6" className="mb-4">Connection Status</Text>
            <PieChart
              data={connectionStatus}
              dataKey="value"
              nameKey="name"
            />
          </Card>
          <Card className="p-6">
            <Text variant="h6" className="mb-4">Usage Trend</Text>
            <LineChart
              data={usageTrend}
              index="name"
              categories={['requests', 'spend']}
            />
          </Card>
        </div>

        {/* Recent Connections */}
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <Text variant="h5">Recent Connections</Text>
            <Button size="sm" asChild>
              <Link href="/marketplace">
                <Plus className="w-4 h-4 mr-2" />
                Add Server
              </Link>
            </Button>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {connections.slice(0, 4).map(conn => (
              <Card key={conn.id} className="p-6 hover:border-primary transition-colors">
                <div className="flex items-start justify-between mb-4">
                  <div>
                    <Text variant="small">{conn.serverName}</Text>
                    <Text variant="small" className="text-muted-foreground">
                      Connected{' '}
                      {Math.floor(
                        (new Date().getTime() - conn.createdAt.getTime()) / (1000 * 60 * 60 * 24)
                      )}{' '}
                      days ago
                    </Text>
                  </div>
                  <div
                    className={`px-2 py-1 rounded text-xs font-medium ${
                      conn.status === 'active'
                        ? 'bg-green-500/20 text-green-700 dark:text-green-400'
                        : 'bg-red-500/20 text-red-700 dark:text-red-400'
                    }`}
                  >
                    {conn.status === 'active' ? 'Active' : 'Expired'}
                  </div>
                </div>

                <div className="space-y-2 text-sm mb-4">
                  <div>
                    <Text variant="caption" className="text-muted-foreground mb-1">Scopes</Text>
                    <div className="flex flex-wrap gap-1">
                      {conn.scopes.slice(0, 2).map((scope: string) => (
                        <span key={scope} className="text-xs bg-muted px-2 py-1 rounded">
                          {scope}
                        </span>
                      ))}
                      {conn.scopes.length > 2 && (
                        <span className="text-xs bg-muted px-2 py-1 rounded">+{conn.scopes.length - 2}</span>
                      )}
                    </div>
                  </div>
                </div>

                {conn.tokenExpiresAt && (
                  <Text variant="caption" className="text-muted-foreground mb-4">
                    Expires: {conn.tokenExpiresAt.toLocaleDateString()}
                  </Text>
                )}

                <Button variant="outline" size="sm" asChild className="w-full">
                  <Link href={`/buyer/connections#${conn.id}`}>Manage</Link>
                </Button>
              </Card>
            ))}
          </div>

          {connections.length === 0 && (
            <Card className="p-8 text-center">
              <Text variant="body" className="text-muted-foreground mb-4">No servers connected yet</Text>
              <Button asChild>
                <Link href="/marketplace">Browse Marketplace</Link>
              </Button>
            </Card>
          )}
        </div>

        {/* Quick Links */}
        <Card className="bg-gradient-to-br from-primary/5 to-primary/10 border-primary/20 p-8">
          <Text variant="h5" className="mb-4">Next Steps</Text>
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
            <Link href="/marketplace" className="group">
              <div className="p-4 rounded-lg bg-background/50 hover:bg-background transition-colors border border-border group-hover:border-primary">
                <Text variant="small" className="mb-1">Explore Marketplace</Text>
                <Text variant="caption" className="text-muted-foreground">Find new servers</Text>
              </div>
            </Link>
            <Link href="/buyer/billing" className="group">
              <div className="p-4 rounded-lg bg-background/50 hover:bg-background transition-colors border border-border group-hover:border-primary">
                <Text variant="small" className="mb-1">Manage Billing</Text>
                <Text variant="caption" className="text-muted-foreground">View invoices & plans</Text>
              </div>
            </Link>
            <Link href="/buyer/connections" className="group">
              <div className="p-4 rounded-lg bg-background/50 hover:bg-background transition-colors border border-border group-hover:border-primary">
                <Text variant="small" className="mb-1">All Connections</Text>
                <Text variant="caption" className="text-muted-foreground">Manage & rotate tokens</Text>
              </div>
            </Link>
          </div>
        </Card>
      </div>
    </AppShell>
  )
}
