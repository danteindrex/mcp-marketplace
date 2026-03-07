'use client'

import { useCallback, useEffect, useState } from 'react'
import { use } from 'react'
import Link from 'next/link'
import { ArrowLeft } from 'lucide-react'
import { AppShell } from '@/components/app-shell'
import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { deployMerchantServer, fetchServerDeployments, type ServerLifecycle } from '@/lib/api-client'
import { toast } from 'sonner'

export default function DeploymentsPage({ params }: { params: Promise<{ serverId: string }> }) {
  const { serverId } = use(params)
  const [items, setItems] = useState<any[]>([])
  const [lifecycle, setLifecycle] = useState<ServerLifecycle | null>(null)
  const [isDeploying, setIsDeploying] = useState(false)

  const load = useCallback(async () => {
    const data = await fetchServerDeployments(serverId)
    setItems(data.items || [])
    setLifecycle(data.lifecycle || null)
  }, [serverId])

  useEffect(() => {
    load().catch(() => {
      setItems([])
      setLifecycle(null)
    })
  }, [load])

  const handleDeploy = async () => {
    setIsDeploying(true)
    try {
      await deployMerchantServer(serverId, { deploymentTarget: 'us-west-1' })
      toast.success('Agent deployed and saved as marketplace draft')
      await load()
    } catch (e: any) {
      toast.error(e?.message || 'Deploy failed')
    } finally {
      setIsDeploying(false)
    }
  }

  return (
    <AppShell role="merchant">
      <div className="p-6 max-w-5xl mx-auto space-y-6">
        <Link href="/merchant/servers" className="flex items-center gap-2 text-muted-foreground hover:text-foreground transition-colors w-fit"><ArrowLeft className="w-4 h-4" />Back to Servers</Link>
        <div><h1 className="text-3xl font-bold mb-2">Deployments</h1><p className="text-muted-foreground">Manage region, replicas, and rollout status.</p></div>

        <Card className="p-6 space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div>
              <p className="text-xs text-muted-foreground">Deployment</p>
              <p className="font-medium capitalize">{lifecycle?.deploymentStatus || 'not_deployed'}</p>
            </div>
            <div>
              <p className="text-xs text-muted-foreground">Marketplace</p>
              <p className="font-medium capitalize">{lifecycle?.marketplaceStatus || 'draft'}</p>
            </div>
            <div>
              <p className="text-xs text-muted-foreground">Publish Readiness</p>
              <p className="font-medium">{lifecycle?.canPublish ? 'Ready' : 'Price required'}</p>
            </div>
          </div>
          <div className="flex gap-3">
            <Button onClick={handleDeploy} disabled={isDeploying}>
              {isDeploying ? 'Deploying...' : 'Deploy Now'}
            </Button>
            <Button variant="outline" asChild>
              <Link href={`/merchant/servers/${serverId}/pricing`}>Set Price / Publish</Link>
            </Button>
          </div>
        </Card>

        <div className="space-y-3">
          {items.length === 0 && (
            <Card className="p-6">
              <p className="text-sm text-muted-foreground">No active deployments yet.</p>
            </Card>
          )}
          {items.map(dep => (
            <Card key={dep.id} className="p-6">
              <div className="grid grid-cols-1 md:grid-cols-6 gap-4 items-center">
                <div><p className="text-xs text-muted-foreground">Environment</p><p className="font-medium capitalize">{dep.environment}</p></div>
                <div><p className="text-xs text-muted-foreground">Region</p><p className="font-medium">{dep.region}</p></div>
                <div><p className="text-xs text-muted-foreground">Replicas</p><p className="font-medium">{dep.replicas}</p></div>
                <div><p className="text-xs text-muted-foreground">Transport</p><p className="font-medium uppercase">{dep.transport}</p></div>
                <div><p className="text-xs text-muted-foreground">Status</p><p className="font-medium">{dep.status}</p></div>
                <Button variant="outline" size="sm" disabled>Rollback</Button>
              </div>
            </Card>
          ))}
        </div>
      </div>
    </AppShell>
  )
}
