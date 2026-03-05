'use client'

import { useState, useEffect } from 'react'
import { TrendingUp, Download } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { AppShell } from '@/components/app-shell'
import { TableToolbar } from '@/components/table-toolbar'
import { fetchMerchantRevenue } from '@/lib/api-client'
import { BarChart } from '@/components/retroui/charts/BarChart'
import { AreaChart } from '@/components/retroui/charts/AreaChart'
import { LineChart } from '@/components/retroui/charts/LineChart'

export default function RevenuePage() {
  const [dateRange, setDateRange] = useState('6m')
  const [searchQuery, setSearchQuery] = useState('')
  const [data, setData] = useState<any>({ totalRevenue: 0, totalCustomers: 0, servers: [], trend: [] })

  useEffect(() => {
    fetchMerchantRevenue().then(setData)
  }, [])

  const filteredServers = (data.servers || []).filter((s: any) => s.name.toLowerCase().includes(searchQuery.toLowerCase()))

  return (
    <AppShell role="merchant">
      <div className="p-6 space-y-6">
        <div className="flex items-center justify-between">
          <div><h1 className="text-3xl font-bold mb-2">Revenue Dashboard</h1><p className="text-muted-foreground">Track earnings from your MCP servers</p></div>
          <div className="flex gap-2">
            <select value={dateRange} onChange={e => setDateRange(e.target.value)} className="px-3 py-2 rounded-md border border-input bg-background text-sm"><option value="1m">Last Month</option><option value="3m">Last 3 Months</option><option value="6m">Last 6 Months</option></select>
            <Button variant="outline" size="sm"><Download className="w-4 h-4 mr-2" />Export</Button>
          </div>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <Card className="p-6"><p className="text-sm text-muted-foreground mb-2">Total Revenue</p><p className="text-3xl font-bold">${Number(data.totalRevenue).toLocaleString()}</p><p className="text-xs text-green-600 dark:text-green-400 mt-2 flex items-center gap-1"><TrendingUp className="w-3 h-3" />+22% from last period</p></Card>
          <Card className="p-6"><p className="text-sm text-muted-foreground mb-2">Active Customers</p><p className="text-3xl font-bold">{data.totalCustomers}</p></Card>
          <Card className="p-6"><p className="text-sm text-muted-foreground mb-2">Avg Revenue / Server</p><p className="text-3xl font-bold">${data.servers?.length ? (Number(data.totalRevenue) / data.servers.length).toLocaleString() : 0}</p></Card>
          <Card className="p-6"><p className="text-sm text-muted-foreground mb-2">Payout Status</p><p className="text-3xl font-bold text-green-600 dark:text-green-400">Processed</p></Card>
        </div>

        <Card className="p-6">
          <h2 className="text-lg font-semibold mb-4">Revenue Trend</h2>
          <BarChart
            data={data.trend || []}
            index="month"
            categories={['subscriptions', 'perCall']}
            stacked
          />
        </Card>

        <div className="grid grid-cols-1 xl:grid-cols-2 gap-4">
          <Card className="p-6">
            <h2 className="text-lg font-semibold mb-4">Cumulative Revenue</h2>
            <AreaChart
              data={(data.trend || []).map((row: any, idx: number) => ({
                ...row,
                total: Number(row.subscriptions || 0) + Number(row.perCall || 0) + idx * 40,
              }))}
              index="month"
              categories={['total']}
            />
          </Card>

          <Card className="p-6">
            <h2 className="text-lg font-semibold mb-4">Subscriptions vs Per-Call</h2>
            <LineChart
              data={data.trend || []}
              index="month"
              categories={['subscriptions', 'perCall']}
            />
          </Card>
        </div>

        <div className="space-y-4">
          <h2 className="text-lg font-semibold mb-4">Revenue by Server</h2>
          <TableToolbar searchPlaceholder="Search by server name..." onSearch={setSearchQuery} />
          <div className="space-y-2">
            {filteredServers.map((server: any) => (
              <Card key={server.id} className="p-6">
                <div className="grid grid-cols-1 md:grid-cols-5 gap-4 items-center">
                  <div><h3 className="font-semibold">{server.name}</h3><p className="text-sm text-muted-foreground">{server.customers} customers</p></div>
                  <div><p className="text-sm text-muted-foreground mb-1">Revenue</p><p className="text-2xl font-bold">${Number(server.revenue).toLocaleString()}</p></div>
                  <div><p className="text-sm text-muted-foreground mb-1">Trend</p><p className="text-lg font-semibold text-green-600 dark:text-green-400">{server.trend}</p></div>
                  <div><p className="text-sm text-muted-foreground mb-1">Avg / Customer</p><p className="text-lg font-semibold">${server.customers ? (Number(server.revenue) / server.customers).toFixed(2) : '0.00'}</p></div>
                  <Button variant="outline" size="sm" className="w-full md:w-auto">View Details</Button>
                </div>
              </Card>
            ))}
          </div>
        </div>
      </div>
    </AppShell>
  )
}
