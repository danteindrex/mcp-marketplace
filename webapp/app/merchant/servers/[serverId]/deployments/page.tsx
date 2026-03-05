'use client'

import { useEffect, useState } from 'react'
import { use } from 'react'
import Link from 'next/link'
import { ArrowLeft } from 'lucide-react'
import { AppShell } from '@/components/app-shell'
import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { fetchServerDeployments } from '@/lib/api-client'

export default function DeploymentsPage({ params }: { params: Promise<{ serverId: string }> }) {
  const { serverId } = use(params)
  const [items, setItems] = useState<any[]>([])

  useEffect(() => {
    fetchServerDeployments(serverId).then(r => setItems(r.items || []))
  }, [serverId])

  return (
    <AppShell role="merchant">
      <div className="p-6 max-w-5xl mx-auto space-y-6">
        <Link href="/merchant/servers" className="flex items-center gap-2 text-muted-foreground hover:text-foreground transition-colors w-fit"><ArrowLeft className="w-4 h-4" />Back to Servers</Link>
        <div><h1 className="text-3xl font-bold mb-2">Deployments</h1><p className="text-muted-foreground">Manage region, replicas, and rollout status.</p></div>

        <div className="space-y-3">
          {items.map(dep => (
            <Card key={dep.id} className="p-6">
              <div className="grid grid-cols-1 md:grid-cols-6 gap-4 items-center">
                <div><p className="text-xs text-muted-foreground">Environment</p><p className="font-medium capitalize">{dep.environment}</p></div>
                <div><p className="text-xs text-muted-foreground">Region</p><p className="font-medium">{dep.region}</p></div>
                <div><p className="text-xs text-muted-foreground">Replicas</p><p className="font-medium">{dep.replicas}</p></div>
                <div><p className="text-xs text-muted-foreground">Transport</p><p className="font-medium uppercase">{dep.transport}</p></div>
                <div><p className="text-xs text-muted-foreground">Status</p><p className="font-medium">{dep.status}</p></div>
                <Button variant="outline" size="sm">Rollback</Button>
              </div>
            </Card>
          ))}
        </div>
      </div>
    </AppShell>
  )
}
