'use client'

import { useState } from 'react'
import { TrendingUp, Download } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { AppShell } from '@/components/app-shell'
import { TableToolbar } from '@/components/table-toolbar'
import {
  LineChart,
  Line,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from 'recharts'

const revenueData = [
  { month: 'Jan', revenue: 4200, subscriptions: 3200, perCall: 1000 },
  { month: 'Feb', revenue: 5100, subscriptions: 3800, perCall: 1300 },
  { month: 'Mar', revenue: 6300, subscriptions: 4200, perCall: 2100 },
  { month: 'Apr', revenue: 7800, subscriptions: 5100, perCall: 2700 },
  { month: 'May', revenue: 8900, subscriptions: 5800, perCall: 3100 },
  { month: 'Jun', revenue: 9200, subscriptions: 5900, perCall: 3300 },
]

const serverRevenue = [
  { name: 'PostgreSQL Assistant', revenue: 4200, customers: 145, trend: '+12%' },
  { name: 'GitHub Integration Suite', revenue: 2800, customers: 98, trend: '+8%' },
  { name: 'Document Analyzer', revenue: 1600, customers: 67, trend: '+15%' },
  { name: 'Email Campaign Manager', revenue: 1200, customers: 45, trend: '+3%' },
  { name: 'API Rate Limiter', revenue: 400, customers: 12, trend: '-2%' },
]

export default function RevenuePage() {
  const [dateRange, setDateRange] = useState('6m')
  const [searchQuery, setSearchQuery] = useState('')

  const filteredServers = serverRevenue.filter(s =>
    s.name.toLowerCase().includes(searchQuery.toLowerCase())
  )

  const totalRevenue = serverRevenue.reduce((sum, s) => sum + s.revenue, 0)
  const totalCustomers = serverRevenue.reduce((sum, s) => sum + s.customers, 0)

  return (
    <AppShell role="merchant">
      <div className="p-6 space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold mb-2">Revenue Dashboard</h1>
            <p className="text-muted-foreground">Track earnings from your MCP servers</p>
          </div>

          <div className="flex gap-2">
            <select
              value={dateRange}
              onChange={e => setDateRange(e.target.value)}
              className="px-3 py-2 rounded-md border border-input bg-background text-sm"
            >
              <option value="1m">Last Month</option>
              <option value="3m">Last 3 Months</option>
              <option value="6m">Last 6 Months</option>
              <option value="1y">Last Year</option>
              <option value="ytd">Year to Date</option>
            </select>
            <Button variant="outline" size="sm">
              <Download className="w-4 h-4 mr-2" />
              Export
            </Button>
          </div>
        </div>

        {/* Key Metrics */}
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <Card className="p-6">
            <p className="text-sm text-muted-foreground mb-2">Total Revenue</p>
            <p className="text-3xl font-bold">${totalRevenue.toLocaleString()}</p>
            <p className="text-xs text-green-600 dark:text-green-400 mt-2 flex items-center gap-1">
              <TrendingUp className="w-3 h-3" />
              +22% from last period
            </p>
          </Card>

          <Card className="p-6">
            <p className="text-sm text-muted-foreground mb-2">Active Customers</p>
            <p className="text-3xl font-bold">{totalCustomers}</p>
            <p className="text-xs text-green-600 dark:text-green-400 mt-2">Across all servers</p>
          </Card>

          <Card className="p-6">
            <p className="text-sm text-muted-foreground mb-2">Avg Revenue / Server</p>
            <p className="text-3xl font-bold">
              ${(totalRevenue / serverRevenue.length).toLocaleString()}
            </p>
            <p className="text-xs text-muted-foreground mt-2">Monthly average</p>
          </Card>

          <Card className="p-6">
            <p className="text-sm text-muted-foreground mb-2">Payout Status</p>
            <p className="text-3xl font-bold text-green-600 dark:text-green-400">Processed</p>
            <p className="text-xs text-muted-foreground mt-2">Next payout: 1 day</p>
          </Card>
        </div>

        {/* Revenue Trend Chart */}
        <Card className="p-6">
          <h2 className="text-lg font-semibold mb-4">Revenue Trend</h2>
          <ResponsiveContainer width="100%" height={300}>
            <BarChart data={revenueData}>
              <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" />
              <XAxis dataKey="month" stroke="var(--muted-foreground)" />
              <YAxis stroke="var(--muted-foreground)" />
              <Tooltip
                contentStyle={{
                  backgroundColor: 'var(--background)',
                  border: '1px solid var(--border)',
                }}
              />
              <Legend />
              <Bar dataKey="subscriptions" fill="var(--primary)" name="Subscriptions" stackId="a" />
              <Bar dataKey="perCall" fill="var(--chart-2)" name="Per-Call Revenue" stackId="a" />
            </BarChart>
          </ResponsiveContainer>
        </Card>

        {/* Revenue by Server */}
        <div className="space-y-4">
          <div>
            <h2 className="text-lg font-semibold mb-4">Revenue by Server</h2>
          </div>

          <TableToolbar
            searchPlaceholder="Search by server name..."
            onSearch={setSearchQuery}
          />

          <div className="space-y-2">
            {filteredServers.length === 0 ? (
              <Card className="p-8 text-center text-muted-foreground">
                <p>No servers found</p>
              </Card>
            ) : (
              filteredServers.map(server => (
                <Card key={server.name} className="p-6">
                  <div className="grid grid-cols-1 md:grid-cols-5 gap-4 items-center">
                    <div>
                      <h3 className="font-semibold">{server.name}</h3>
                      <p className="text-sm text-muted-foreground">
                        {server.customers} customer{server.customers !== 1 ? 's' : ''}
                      </p>
                    </div>

                    <div>
                      <p className="text-sm text-muted-foreground mb-1">Revenue</p>
                      <p className="text-2xl font-bold">${server.revenue.toLocaleString()}</p>
                    </div>

                    <div>
                      <p className="text-sm text-muted-foreground mb-1">Trend</p>
                      <p className={`text-lg font-semibold ${
                        server.trend.startsWith('+')
                          ? 'text-green-600 dark:text-green-400'
                          : 'text-red-600 dark:text-red-400'
                      }`}>
                        {server.trend}
                      </p>
                    </div>

                    <div>
                      <p className="text-sm text-muted-foreground mb-1">Avg / Customer</p>
                      <p className="text-lg font-semibold">
                        ${(server.revenue / server.customers).toFixed(2)}
                      </p>
                    </div>

                    <Button variant="outline" size="sm" className="w-full md:w-auto">
                      View Details
                    </Button>
                  </div>
                </Card>
              ))
            )}
          </div>
        </div>

        {/* Invoices & Payouts */}
        <Card className="p-6">
          <h2 className="text-lg font-semibold mb-4">Recent Payouts</h2>
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-border">
                  <th className="text-left py-3 px-4 font-semibold text-muted-foreground">Date</th>
                  <th className="text-left py-3 px-4 font-semibold text-muted-foreground">Period</th>
                  <th className="text-left py-3 px-4 font-semibold text-muted-foreground">Amount</th>
                  <th className="text-left py-3 px-4 font-semibold text-muted-foreground">Status</th>
                  <th className="text-left py-3 px-4 font-semibold text-muted-foreground">Action</th>
                </tr>
              </thead>
              <tbody>
                {[
                  { date: 'Jun 15, 2025', period: 'May 2025', amount: 8900, status: 'Paid' },
                  { date: 'May 15, 2025', period: 'Apr 2025', amount: 7800, status: 'Paid' },
                  { date: 'Apr 15, 2025', period: 'Mar 2025', amount: 6300, status: 'Paid' },
                  { date: 'Mar 15, 2025', period: 'Feb 2025', amount: 5100, status: 'Paid' },
                  { date: 'Feb 15, 2025', period: 'Jan 2025', amount: 4200, status: 'Paid' },
                ].map((payout, i) => (
                  <tr key={i} className="border-b border-border hover:bg-muted/50">
                    <td className="py-3 px-4">{payout.date}</td>
                    <td className="py-3 px-4">{payout.period}</td>
                    <td className="py-3 px-4 font-semibold">${payout.amount.toLocaleString()}</td>
                    <td className="py-3 px-4">
                      <span className="inline-flex items-center gap-1 text-xs bg-green-500/20 text-green-700 dark:text-green-400 px-2 py-1 rounded">
                        {payout.status}
                      </span>
                    </td>
                    <td className="py-3 px-4">
                      <Button variant="ghost" size="sm">
                        View Invoice
                      </Button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </Card>
      </div>
    </AppShell>
  )
}
