'use client'

import { useState, useEffect } from 'react'
import { use } from 'react'
import Link from 'next/link'
import { ArrowLeft, Download } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { AppShell } from '@/components/app-shell'
import { fetchServerObservability } from '@/lib/api-client'
import { BarChart } from '@/components/retroui/charts/BarChart'
import { LineChart } from '@/components/retroui/charts/LineChart'

export default function ObservabilityPage({ params }: { params: Promise<{ serverId: string }> }) {
  const { serverId } = use(params)
  const [dateRange, setDateRange] = useState('24h')
  const [data, setData] = useState<any>(null)

  useEffect(() => {
    fetchServerObservability(serverId).then(setData)
  }, [serverId])

  if (!data) return <AppShell role="merchant"><div className="p-6">Loading...</div></AppShell>

  const bars = [
    { p: 'p50', latency: data.metrics.p50LatencyMs },
    { p: 'p95', latency: data.metrics.p95LatencyMs },
  ]
  const trend = [
    { window: 'T-3', errors: Math.max(0, data.metrics.insufficientScopeCount - 2), latency: data.metrics.p50LatencyMs - 6 },
    { window: 'T-2', errors: Math.max(0, data.metrics.insufficientScopeCount - 1), latency: data.metrics.p50LatencyMs - 3 },
    { window: 'T-1', errors: data.metrics.insufficientScopeCount, latency: data.metrics.p50LatencyMs },
  ]

  return (
    <AppShell role="merchant">
      <div className="p-6 space-y-6">
        <Link href="/merchant/servers" className="flex items-center gap-2 text-muted-foreground hover:text-foreground transition-colors w-fit"><ArrowLeft className="w-4 h-4" />Back to Servers</Link>
        <div className="flex items-center justify-between">
          <div><h1 className="text-3xl font-bold mb-2">Observability</h1><p className="text-muted-foreground">Monitor server performance and tool usage</p></div>
          <div className="flex gap-2"><select value={dateRange} onChange={e => setDateRange(e.target.value)} className="px-3 py-2 rounded-md border border-input bg-background text-sm"><option value="24h">Last 24 Hours</option><option value="7d">Last 7 Days</option><option value="30d">Last 30 Days</option></select><Button variant="outline" size="sm"><Download className="w-4 h-4 mr-2" />Export</Button></div>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <Card className="p-6"><p className="text-sm text-muted-foreground mb-2">p50 Latency</p><p className="text-3xl font-bold">{data.metrics.p50LatencyMs}ms</p></Card>
          <Card className="p-6"><p className="text-sm text-muted-foreground mb-2">p95 Latency</p><p className="text-3xl font-bold">{data.metrics.p95LatencyMs}ms</p></Card>
          <Card className="p-6"><p className="text-sm text-muted-foreground mb-2">Error Rate</p><p className="text-3xl font-bold">{(data.metrics.errorRate * 100).toFixed(2)}%</p></Card>
          <Card className="p-6"><p className="text-sm text-muted-foreground mb-2">Insufficient Scope</p><p className="text-3xl font-bold">{data.metrics.insufficientScopeCount}</p></Card>
        </div>

        <Card className="p-6">
          <h2 className="text-lg font-semibold mb-4">Latency Percentiles</h2>
          <BarChart
            data={bars}
            index="p"
            categories={['latency']}
          />
        </Card>

        <Card className="p-6">
          <h2 className="text-lg font-semibold mb-4">Recent Error/Latency Trend</h2>
          <LineChart
            data={trend}
            index="window"
            categories={['errors', 'latency']}
          />
        </Card>
      </div>
    </AppShell>
  )
}
