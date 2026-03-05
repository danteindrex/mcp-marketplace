'use client'

import { useEffect, useState } from 'react'
import { use } from 'react'
import Link from 'next/link'
import { ArrowLeft, Copy } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { AppShell } from '@/components/app-shell'
import { fetchServerPricing } from '@/lib/api-client'
import { toast } from 'sonner'

export default function PricingPage({ params }: { params: Promise<{ serverId: string }> }) {
  const { serverId } = use(params)
  const [data, setData] = useState<any>(null)

  useEffect(() => {
    fetchServerPricing(serverId).then(setData)
  }, [serverId])

  if (!data) return <AppShell role="merchant"><div className="p-6">Loading...</div></AppShell>

  return (
    <AppShell role="merchant">
      <div className="max-w-4xl mx-auto p-6 space-y-6">
        <Link href="/merchant/servers" className="flex items-center gap-2 text-muted-foreground hover:text-foreground transition-colors w-fit"><ArrowLeft className="w-4 h-4" />Back to Servers</Link>
        <div><h1 className="text-3xl font-bold mb-2">Pricing Configuration</h1><p className="text-muted-foreground">Set up how users will pay for your server</p></div>

        <Card className="p-8 space-y-4">
          <h2 className="text-xl font-bold">Current Pricing</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div><p className="text-xs text-muted-foreground mb-1">Pricing Type</p><p className="font-medium capitalize">{data.pricing.type}</p></div>
            <div><p className="text-xs text-muted-foreground mb-1">Amount</p><p className="font-medium">{Number(data.pricing.amount).toFixed(2)}</p></div>
          </div>
          <div className="border-t border-border pt-4 space-y-3">
            <p className="text-sm font-semibold">x402 Configuration</p>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
              <div><p className="text-muted-foreground">Version</p><p>{data.pricing.x402.version}</p></div>
              <div><p className="text-muted-foreground">Network</p><p>{data.pricing.x402.network}</p></div>
              <div><p className="text-muted-foreground">Asset</p><p>{data.pricing.x402.asset}</p></div>
              <div><p className="text-muted-foreground">CAIP-2</p><p className="font-mono">{data.pricing.x402.caip2}</p></div>
            </div>
            <Button variant="outline" onClick={() => { navigator.clipboard.writeText(JSON.stringify(data.pricing.x402)); toast.success('x402 config copied') }}><Copy className="w-4 h-4 mr-2" />Copy x402 Config</Button>
          </div>
        </Card>
      </div>
    </AppShell>
  )
}
