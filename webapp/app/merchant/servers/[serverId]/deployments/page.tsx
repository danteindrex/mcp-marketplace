'use client'

import { useCallback, useEffect, useState } from 'react'
import { use } from 'react'
import Link from 'next/link'
import { ArrowLeft } from 'lucide-react'
import { AppShell } from '@/components/app-shell'
import { Text } from '@/components/retroui/Text'
import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { deployMerchantServer, fetchServerDeployments, type ServerLifecycle } from '@/lib/api-client'
import { toast } from 'sonner'

export default function DeploymentsPage({ params }: { params: Promise<{ serverId: string }> }) {
  const { serverId } = use(params)
  const [items, setItems] = useState<any[]>([])
  const [lifecycle, setLifecycle] = useState<ServerLifecycle | null>(null)
  const [queue, setQueue] = useState<any>(null)
  const [isDeploying, setIsDeploying] = useState(false)
  const [target, setTarget] = useState<'local-docker' | 'external'>('local-docker')

  const load = useCallback(async () => {
    const data = await fetchServerDeployments(serverId)
    setItems(data.items || [])
    setLifecycle(data.lifecycle || null)
    setQueue(data.queue || null)
  }, [serverId])

  useEffect(() => {
    load().catch(() => {
      setItems([])
      setLifecycle(null)
      setQueue(null)
    })
  }, [load])

  useEffect(() => {
    if (lifecycle?.deploymentStatus !== 'deploy_queued') return
    const id = window.setInterval(() => {
      load().catch(() => undefined)
    }, 5000)
    return () => window.clearInterval(id)
  }, [lifecycle?.deploymentStatus, load])

  const handleDeploy = async () => {
    setIsDeploying(true)
    try {
      await deployMerchantServer(serverId, { deploymentTarget: target })
      toast.success(
        target === 'external'
          ? 'External runtime marked as deployed.'
          : 'Deploy queued. We will keep retrying automatically.',
      )
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
        <div><Text variant="h3" className="mb-2">Deployments</Text><Text variant="body" className="text-muted-foreground">Manage region, replicas, and rollout status.</Text></div>

        <Card className="p-6 space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div>
              <Text variant="caption" className="text-muted-foreground">Deployment</Text>
              <Text variant="body" className="capitalize">{lifecycle?.deploymentStatus || 'not_deployed'}</Text>
            </div>
            <div>
              <Text variant="caption" className="text-muted-foreground">Marketplace</Text>
              <Text variant="body" className="capitalize">{lifecycle?.marketplaceStatus || 'draft'}</Text>
            </div>
            <div>
              <Text variant="caption" className="text-muted-foreground">Publish Readiness</Text>
              <Text variant="body">{lifecycle?.canPublish ? 'Ready' : 'Price required'}</Text>
            </div>
          </div>
          {queue?.status && (
            <div className="text-sm rounded border border-border p-3">
              <Text variant="body" className="capitalize">Queue: {String(queue.status).replace('_', ' ')}</Text>
              <Text variant="small" className="text-muted-foreground">Attempts: {queue.attemptCount || 0}/{queue.maxAttempts || 0}</Text>
              {queue.lastError && <Text variant="caption" className="mt-1 text-destructive">{queue.lastError}</Text>}
            </div>
          )}
          <div className="flex flex-wrap gap-3 items-center">
            <select
              value={target}
              onChange={e => setTarget(e.target.value as 'local-docker' | 'external')}
              className="px-3 py-2 rounded-md border border-input bg-background text-sm"
            >
              <option value="local-docker">Managed deploy (Docker)</option>
              <option value="external">External hosted MCP URL</option>
            </select>
            <Button onClick={handleDeploy} disabled={isDeploying}>
              {isDeploying ? 'Deploying...' : target === 'external' ? 'Mark External Runtime Ready' : 'Deploy Now'}
            </Button>
            <Button variant="outline" asChild>
              <Link href={`/merchant/servers/${serverId}/pricing`}>Set Price / Publish</Link>
            </Button>
          </div>
        </Card>

        <div className="space-y-3">
          {items.length === 0 && (
            <Card className="p-6">
              <Text variant="small" className="text-muted-foreground">No active deployments yet.</Text>
            </Card>
          )}
          {items.map(dep => (
            <Card key={dep.id} className="p-6">
              <div className="grid grid-cols-1 md:grid-cols-6 gap-4 items-center">
                <div><Text variant="caption" className="text-muted-foreground">Environment</Text><Text variant="body" className="capitalize">{dep.environment}</Text></div>
                <div><Text variant="caption" className="text-muted-foreground">Region</Text><Text variant="body">{dep.region}</Text></div>
                <div><Text variant="caption" className="text-muted-foreground">Replicas</Text><Text variant="body">{dep.replicas}</Text></div>
                <div><Text variant="caption" className="text-muted-foreground">Transport</Text><Text variant="body" className="uppercase">{dep.transport}</Text></div>
                <div><Text variant="caption" className="text-muted-foreground">Status</Text><Text variant="body">{dep.status}</Text></div>
                <div><Text variant="caption" className="text-muted-foreground">URL</Text><Text variant="body" className="break-all">{dep.url || 'n/a'}</Text></div>
                <Button variant="outline" size="sm" disabled>Rollback</Button>
              </div>
            </Card>
          ))}
        </div>
      </div>
    </AppShell>
  )
}
