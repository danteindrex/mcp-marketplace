'use client'

import { useState, useEffect } from 'react'
import { use } from 'react'
import Link from 'next/link'
import { ArrowLeft, Download, AlertCircle } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { AppShell } from '@/components/app-shell'
import { fetchServerObservability } from '@/lib/api-client'
import { Text } from '@/components/retroui/text'
import { BarChart } from '@/components/retroui/charts/BarChart'
import { LineChart } from '@/components/retroui/charts/LineChart'

interface ObservabilityMetrics {
  p50LatencyMs: number
  p95LatencyMs: number
  errorRate: number
  insufficientScopeCount: number
  totalRequests: number
  history?: Array<{
    window: string
    errors: number
    latency: number
    timestamp: string
  }>
}

export default function ObservabilityPage({ params }: { params: Promise<{ serverId: string }> }) {
  const { serverId } = use(params)
  const [dateRange, setDateRange] = useState('24h')
  const [data, setData] = useState<{ metrics: ObservabilityMetrics } | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    fetchServerObservability(serverId)
      .then((result) => setData(result as { metrics: ObservabilityMetrics } | null))
      .catch((err) => setError(err instanceof Error ? err.message : 'Failed to load observability data'))
  }, [serverId])

  if (error) return <AppShell role="merchant"><div className="p-6"><div className="flex items-center gap-2 text-red-500"><AlertCircle className="w-5 h-5" /><Text variant="body">Error loading observability: {error}</Text></div></div></AppShell>
  if (!data) return <AppShell role="merchant"><div className="p-6">Loading...</div></AppShell>

  const bars = [
    { p: 'p50', latency: data.metrics.p50LatencyMs },
    { p: 'p95', latency: data.metrics.p95LatencyMs },
  ]

  // Use real historical data from API if available, otherwise show empty state
  const trend = data.metrics.history || []
  const hasHistoricalData = trend.length > 0

  return (
    <AppShell role="merchant">
      <div className="p-6 space-y-6">
        <Link href="/merchant/servers" className="flex items-center gap-2 text-muted-foreground hover:text-foreground transition-colors w-fit"><ArrowLeft className="w-4 h-4" />Back to Servers</Link>
        <div className="flex items-center justify-between">
          <div><Text variant="h3" className="mb-2">Observability</Text><Text variant="body" className="text-muted-foreground">Monitor server performance and tool usage</Text></div>
          <div className="flex gap-2"><select value={dateRange} onChange={e => setDateRange(e.target.value)} className="px-3 py-2 rounded-md border border-input bg-background text-sm"><option value="24h">Last 24 Hours</option><option value="7d">Last 7 Days</option><option value="30d">Last 30 Days</option></select><Button variant="outline" size="sm"><Download className="w-4 h-4 mr-2" />Export</Button></div>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <Card className="p-6"><Text variant="small" className="mb-2 text-muted-foreground">p50 Latency</Text><Text variant="h3">{data.metrics.p50LatencyMs}ms</Text></Card>
          <Card className="p-6"><Text variant="small" className="mb-2 text-muted-foreground">p95 Latency</Text><Text variant="h3">{data.metrics.p95LatencyMs}ms</Text></Card>
          <Card className="p-6"><Text variant="small" className="mb-2 text-muted-foreground">Error Rate</Text><Text variant="h3">{(data.metrics.errorRate * 100).toFixed(2)}%</Text></Card>
          <Card className="p-6"><Text variant="small" className="mb-2 text-muted-foreground">Insufficient Scope</Text><Text variant="h3">{data.metrics.insufficientScopeCount}</Text></Card>
        </div>

        <Card className="p-6">
          <Text variant="h6" className="mb-4">Latency Percentiles</Text>
          <BarChart
            data={bars}
            index="p"
            categories={['latency']}
          />
        </Card>

        <Card className="p-6">
          <Text variant="h6" className="mb-4">Recent Error/Latency Trend</Text>
          {hasHistoricalData ? (
            <LineChart
              data={trend}
              index="window"
              categories={['errors', 'latency']}
            />
          ) : (
            <div className="text-center py-12 text-muted-foreground">
              <Text variant="small">No historical trend data available yet.</Text>
              <Text variant="caption" className="mt-1">Error and latency trends will appear as data is collected over time.</Text>
            </div>
          )}
        </Card>
      </div>
    </AppShell>
  )
}
