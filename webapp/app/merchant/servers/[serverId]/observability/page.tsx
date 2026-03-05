'use client'

import { useState } from 'react'
import Link from 'next/link'
import { ArrowLeft, Download } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { AppShell } from '@/components/app-shell'
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
  PieChart,
  Pie,
  Cell,
} from 'recharts'

// Sample metrics data
const invocationData = [
  { time: '00:00', invocations: 145, errors: 3 },
  { time: '04:00', invocations: 89, errors: 1 },
  { time: '08:00', invocations: 234, errors: 5 },
  { time: '12:00', invocations: 412, errors: 8 },
  { time: '16:00', invocations: 356, errors: 6 },
  { time: '20:00', invocations: 278, errors: 4 },
  { time: '23:59', invocations: 156, errors: 2 },
]

const latencyData = [
  { p: 'p50', latency: 45 },
  { p: 'p75', latency: 82 },
  { p: 'p90', latency: 156 },
  { p: 'p95', latency: 234 },
  { p: 'p99', latency: 567 },
]

const errorRates = [
  { name: 'Success', value: 9542 },
  { name: '4xx Errors', value: 287 },
  { name: '5xx Errors', value: 171 },
]

const toolMetrics = [
  { tool: 'query', calls: 3456, avgLatency: 87, errorRate: 0.02 },
  { tool: 'execute', calls: 2134, avgLatency: 156, errorRate: 0.04 },
  { tool: 'analyze', calls: 1234, avgLatency: 234, errorRate: 0.01 },
  { tool: 'transform', calls: 987, avgLatency: 45, errorRate: 0.00 },
  { tool: 'validate', calls: 456, avgLatency: 23, errorRate: 0.00 },
]

const colors = ['#10b981', '#ef4444', '#f97316']

export default function ObservabilityPage({ params }: any) {
  const [dateRange, setDateRange] = useState('24h')

  return (
    <AppShell role="merchant">
      <div className="p-6 space-y-6">
        {/* Header */}
        <Link
          href="/merchant/servers"
          className="flex items-center gap-2 text-muted-foreground hover:text-foreground transition-colors w-fit"
        >
          <ArrowLeft className="w-4 h-4" />
          Back to Servers
        </Link>

        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold mb-2">Observability</h1>
            <p className="text-muted-foreground">Monitor server performance and tool usage</p>
          </div>

          <div className="flex gap-2">
            <select
              value={dateRange}
              onChange={e => setDateRange(e.target.value)}
              className="px-3 py-2 rounded-md border border-input bg-background text-sm"
            >
              <option value="24h">Last 24 Hours</option>
              <option value="7d">Last 7 Days</option>
              <option value="30d">Last 30 Days</option>
              <option value="90d">Last 90 Days</option>
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
            <p className="text-sm text-muted-foreground mb-2">Total Invocations</p>
            <p className="text-3xl font-bold">7,842</p>
            <p className="text-xs text-green-600 dark:text-green-400 mt-2">+12% from yesterday</p>
          </Card>
          <Card className="p-6">
            <p className="text-sm text-muted-foreground mb-2">Avg Latency</p>
            <p className="text-3xl font-bold">127ms</p>
            <p className="text-xs text-green-600 dark:text-green-400 mt-2">Optimal performance</p>
          </Card>
          <Card className="p-6">
            <p className="text-sm text-muted-foreground mb-2">Error Rate</p>
            <p className="text-3xl font-bold">1.8%</p>
            <p className="text-xs text-amber-600 dark:text-amber-400 mt-2">3 errors in last hour</p>
          </Card>
          <Card className="p-6">
            <p className="text-sm text-muted-foreground mb-2">Uptime</p>
            <p className="text-3xl font-bold">99.95%</p>
            <p className="text-xs text-green-600 dark:text-green-400 mt-2">Running smoothly</p>
          </Card>
        </div>

        {/* Charts */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* Invocations Chart */}
          <Card className="p-6">
            <h2 className="text-lg font-semibold mb-4">Invocations Over Time</h2>
            <ResponsiveContainer width="100%" height={300}>
              <LineChart data={invocationData}>
                <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" />
                <XAxis dataKey="time" stroke="var(--muted-foreground)" />
                <YAxis stroke="var(--muted-foreground)" />
                <Tooltip
                  contentStyle={{
                    backgroundColor: 'var(--background)',
                    border: '1px solid var(--border)',
                  }}
                />
                <Legend />
                <Line
                  type="monotone"
                  dataKey="invocations"
                  stroke="var(--primary)"
                  name="Invocations"
                  dot={false}
                />
                <Line
                  type="monotone"
                  dataKey="errors"
                  stroke="var(--destructive)"
                  name="Errors"
                  dot={false}
                />
              </LineChart>
            </ResponsiveContainer>
          </Card>

          {/* Latency Percentiles */}
          <Card className="p-6">
            <h2 className="text-lg font-semibold mb-4">Latency Percentiles (ms)</h2>
            <ResponsiveContainer width="100%" height={300}>
              <BarChart data={latencyData}>
                <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" />
                <XAxis dataKey="p" stroke="var(--muted-foreground)" />
                <YAxis stroke="var(--muted-foreground)" />
                <Tooltip
                  contentStyle={{
                    backgroundColor: 'var(--background)',
                    border: '1px solid var(--border)',
                  }}
                />
                <Bar dataKey="latency" fill="var(--primary)" name="Latency (ms)" />
              </BarChart>
            </ResponsiveContainer>
          </Card>

          {/* Error Distribution */}
          <Card className="p-6">
            <h2 className="text-lg font-semibold mb-4">Error Distribution</h2>
            <ResponsiveContainer width="100%" height={300}>
              <PieChart>
                <Pie
                  data={errorRates}
                  cx="50%"
                  cy="50%"
                  labelLine={false}
                  label={({ name, value }) => `${name}: ${value}`}
                  outerRadius={100}
                  fill="#8884d8"
                  dataKey="value"
                >
                  {errorRates.map((entry, index) => (
                    <Cell key={`cell-${index}`} fill={colors[index % colors.length]} />
                  ))}
                </Pie>
                <Tooltip
                  contentStyle={{
                    backgroundColor: 'var(--background)',
                    border: '1px solid var(--border)',
                  }}
                />
              </PieChart>
            </ResponsiveContainer>
          </Card>

          {/* Tool Performance */}
          <Card className="p-6">
            <h2 className="text-lg font-semibold mb-4">Tool Performance</h2>
            <div className="space-y-3">
              {toolMetrics.map(tool => (
                <div key={tool.tool} className="flex items-center justify-between p-3 bg-muted/50 rounded">
                  <div>
                    <p className="font-medium capitalize text-sm">{tool.tool}</p>
                    <p className="text-xs text-muted-foreground">
                      {tool.calls.toLocaleString()} calls
                    </p>
                  </div>
                  <div className="text-right">
                    <p className="text-sm font-medium">{tool.avgLatency}ms</p>
                    <p className="text-xs text-muted-foreground">
                      {(tool.errorRate * 100).toFixed(2)}% errors
                    </p>
                  </div>
                </div>
              ))}
            </div>
          </Card>
        </div>

        {/* Detailed Logs */}
        <Card className="p-6">
          <h2 className="text-lg font-semibold mb-4">Recent Errors</h2>
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-border">
                  <th className="text-left py-3 px-4 font-semibold text-muted-foreground">
                    Timestamp
                  </th>
                  <th className="text-left py-3 px-4 font-semibold text-muted-foreground">
                    Tool
                  </th>
                  <th className="text-left py-3 px-4 font-semibold text-muted-foreground">
                    Error Code
                  </th>
                  <th className="text-left py-3 px-4 font-semibold text-muted-foreground">
                    Message
                  </th>
                </tr>
              </thead>
              <tbody>
                {[
                  { time: '12:45:32', tool: 'execute', code: '503', msg: 'Service temporarily unavailable' },
                  { time: '12:32:15', tool: 'query', code: '400', msg: 'Invalid query parameter' },
                  { time: '12:18:42', tool: 'analyze', code: '500', msg: 'Internal server error' },
                  { time: '12:05:18', tool: 'transform', code: '429', msg: 'Rate limit exceeded' },
                  { time: '11:52:03', tool: 'validate', code: '400', msg: 'Validation failed' },
                ].map((error, i) => (
                  <tr key={i} className="border-b border-border hover:bg-muted/50">
                    <td className="py-3 px-4 font-mono text-xs">{error.time}</td>
                    <td className="py-3 px-4 capitalize">{error.tool}</td>
                    <td className="py-3 px-4">
                      <span className="inline-flex items-center gap-1 text-xs bg-red-500/20 text-red-700 dark:text-red-400 px-2 py-1 rounded">
                        {error.code}
                      </span>
                    </td>
                    <td className="py-3 px-4 text-muted-foreground">{error.msg}</td>
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
