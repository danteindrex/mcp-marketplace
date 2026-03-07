'use client'

import { useEffect, useState } from 'react'
import { AppShell } from '@/components/app-shell'
import { Card } from '@/components/ui/card'
import { fetchAdminPaymentsOverview } from '@/lib/api-client'

export default function AdminPaymentsPage() {
  const [data, setData] = useState<any>({
    intentCount: 0,
    settledCount: 0,
    pendingCount: 0,
    failedCount: 0,
    settledVolumeUsdc: 0,
    byMethod: {},
    x402: {},
    methods: [],
  })

  useEffect(() => {
    fetchAdminPaymentsOverview().then(setData)
  }, [])

  return (
    <AppShell role="admin">
      <div className="p-6 space-y-6">
        <div>
          <h1 className="text-3xl font-bold mb-2">Payments Control Center</h1>
          <p className="text-muted-foreground">Monitor x402 settlement health, methods, and integration readiness.</p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <Card className="p-5"><p className="text-xs text-muted-foreground">Total Intents</p><p className="text-3xl font-bold">{data.intentCount}</p></Card>
          <Card className="p-5"><p className="text-xs text-muted-foreground">Settled</p><p className="text-3xl font-bold">{data.settledCount}</p></Card>
          <Card className="p-5"><p className="text-xs text-muted-foreground">Pending</p><p className="text-3xl font-bold">{data.pendingCount}</p></Card>
          <Card className="p-5"><p className="text-xs text-muted-foreground">Volume (USDC)</p><p className="text-3xl font-bold">{Number(data.settledVolumeUsdc || 0).toFixed(2)}</p></Card>
        </div>

        <Card className="p-6 space-y-3">
          <h2 className="text-xl font-bold">x402 Integration Status</h2>
          <p className="text-sm">Mode: <span className="font-semibold">{data.x402?.mode || 'mock'}</span></p>
          <p className="text-sm">Facilitator URL: <span className="font-mono text-xs">{data.x402?.facilitatorUrl || 'not configured'}</span></p>
          <p className="text-sm">API Key Set: <span className="font-semibold">{data.x402?.apiKeySet ? 'yes' : 'no'}</span></p>
        </Card>

        <Card className="p-6">
          <h2 className="text-xl font-bold mb-3">Supported Payment Methods</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
            {(data.methods || []).map((method: any) => (
              <div key={method.id} className="border border-border rounded-md p-3">
                <p className="font-semibold">{method.displayName}</p>
                <p className="text-xs text-muted-foreground mt-1">{method.integration}</p>
                <p className="text-xs mt-1">{method.configured ? 'Configured' : 'Not configured'}</p>
                {method.docs ? <p className="text-xs text-muted-foreground mt-1">{method.docs}</p> : null}
              </div>
            ))}
          </div>
        </Card>
      </div>
    </AppShell>
  )
}
