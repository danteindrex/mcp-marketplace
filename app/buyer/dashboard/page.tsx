'use client'

import { useEffect, useState } from 'react'
import Link from 'next/link'
import { AlertCircle, TrendingUp, Zap, Clock, DollarSign, Plus } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { AppShell } from '@/components/app-shell'
import { LoadingState } from '@/components/empty-state'
import { fetchConnections, fetchBilling } from '@/lib/api-client'
import { mockConnections, mockBilling } from '@/lib/mock-data'

const userId = 'user_001' // Mock current user

export default function BuyerDashboardPage() {
  const [isLoading, setIsLoading] = useState(true)
  const [connections, setConnections] = useState(mockConnections)
  const [billing, setBilling] = useState(mockBilling[0])

  useEffect(() => {
    const loadData = async () => {
      setIsLoading(true)
      try {
        const [conns, bill] = await Promise.all([
          fetchConnections(userId),
          fetchBilling(userId),
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

  return (
    <AppShell role="buyer">
      <div className="p-6 space-y-8">
        {/* Page Header */}
        <div>
          <h1 className="text-3xl font-bold mb-2">Dashboard</h1>
          <p className="text-muted-foreground">Welcome back! Here's an overview of your MCP servers and usage.</p>
        </div>

        {/* Stats Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <Card className="p-6">
            <div className="flex items-center justify-between mb-2">
              <p className="text-sm text-muted-foreground">Active Connections</p>
              <Zap className="w-4 h-4 text-primary" />
            </div>
            <p className="text-3xl font-bold">{activeConnections}</p>
            <p className="text-xs text-muted-foreground mt-1">
              {expiredConnections > 0 ? `${expiredConnections} need renewal` : 'All active'}
            </p>
          </Card>

          <Card className="p-6">
            <div className="flex items-center justify-between mb-2">
              <p className="text-sm text-muted-foreground">Monthly Spend</p>
              <DollarSign className="w-4 h-4 text-primary" />
            </div>
            <p className="text-3xl font-bold">${billing.monthlySpend.toFixed(2)}</p>
            <p className="text-xs text-muted-foreground mt-1">Billing cycle</p>
          </Card>

          <Card className="p-6">
            <div className="flex items-center justify-between mb-2">
              <p className="text-sm text-muted-foreground">Account Balance</p>
              <TrendingUp className="w-4 h-4 text-primary" />
            </div>
            <p className="text-3xl font-bold">${billing.currentBalance.toFixed(2)}</p>
            <p className="text-xs text-muted-foreground mt-1">Available credit</p>
          </Card>

          <Card className="p-6">
            <div className="flex items-center justify-between mb-2">
              <p className="text-sm text-muted-foreground">Plan</p>
              <Clock className="w-4 h-4 text-primary" />
            </div>
            <p className="text-3xl font-bold capitalize">{billing.plan}</p>
            <p className="text-xs text-muted-foreground mt-1">Current tier</p>
          </Card>
        </div>

        {/* Alerts */}
        {expiredConnections > 0 && (
          <Card className="bg-amber-500/10 border-amber-200 dark:border-amber-800 p-4 flex items-start gap-4">
            <AlertCircle className="w-5 h-5 text-amber-600 dark:text-amber-500 flex-shrink-0 mt-0.5" />
            <div>
              <h3 className="font-semibold text-amber-900 dark:text-amber-100">
                {expiredConnections} Connection{expiredConnections > 1 ? 's' : ''} Expired
              </h3>
              <p className="text-sm text-amber-800 dark:text-amber-200 mt-1">
                Some of your tokens have expired. Please renew them to maintain access.
              </p>
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
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <h2 className="text-xl font-semibold">Recent Connections</h2>
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
                    <h3 className="font-semibold">{conn.serverName}</h3>
                    <p className="text-sm text-muted-foreground">
                      Connected{' '}
                      {Math.floor(
                        (new Date().getTime() - conn.createdAt.getTime()) / (1000 * 60 * 60 * 24)
                      )}{' '}
                      days ago
                    </p>
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
                    <p className="text-muted-foreground text-xs mb-1">Scopes</p>
                    <div className="flex flex-wrap gap-1">
                      {conn.scopes.slice(0, 2).map(scope => (
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
                  <p className="text-xs text-muted-foreground mb-4">
                    Expires: {conn.tokenExpiresAt.toLocaleDateString()}
                  </p>
                )}

                <Button variant="outline" size="sm" asChild className="w-full">
                  <Link href={`/buyer/connections#${conn.id}`}>Manage</Link>
                </Button>
              </Card>
            ))}
          </div>

          {connections.length === 0 && (
            <Card className="p-8 text-center">
              <p className="text-muted-foreground mb-4">No servers connected yet</p>
              <Button asChild>
                <Link href="/marketplace">Browse Marketplace</Link>
              </Button>
            </Card>
          )}
        </div>

        {/* Quick Links */}
        <Card className="bg-gradient-to-br from-primary/5 to-primary/10 border-primary/20 p-8">
          <h2 className="text-xl font-semibold mb-4">Next Steps</h2>
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
            <Link href="/marketplace" className="group">
              <div className="p-4 rounded-lg bg-background/50 hover:bg-background transition-colors border border-border group-hover:border-primary">
                <p className="font-medium text-sm mb-1">Explore Marketplace</p>
                <p className="text-xs text-muted-foreground">Find new servers</p>
              </div>
            </Link>
            <Link href="/buyer/billing" className="group">
              <div className="p-4 rounded-lg bg-background/50 hover:bg-background transition-colors border border-border group-hover:border-primary">
                <p className="font-medium text-sm mb-1">Manage Billing</p>
                <p className="text-xs text-muted-foreground">View invoices & plans</p>
              </div>
            </Link>
            <Link href="/buyer/connections" className="group">
              <div className="p-4 rounded-lg bg-background/50 hover:bg-background transition-colors border border-border group-hover:border-primary">
                <p className="font-medium text-sm mb-1">All Connections</p>
                <p className="text-xs text-muted-foreground">Manage & rotate tokens</p>
              </div>
            </Link>
          </div>
        </Card>
      </div>
    </AppShell>
  )
}
