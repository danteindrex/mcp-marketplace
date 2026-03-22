import Link from 'next/link'
import { cookies } from 'next/headers'
import { redirect } from 'next/navigation'
import { ExternalLink, Plus, Workflow } from 'lucide-react'
import { AppShell } from '@/components/app-shell'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Text } from '@/components/retroui/text'
import type { Server } from '@/lib/api-client'
import { getServerApiBase } from '@/lib/api-base'

type RuntimeConfig = { n8n?: { url?: string } }

async function loadRuntimeConfig(apiBase: string) {
  try {
    const res = await fetch(`${apiBase}/v1/runtime-config`, { cache: 'no-store' })
    if (!res.ok) {
      return null
    }
    return (await res.json()) as RuntimeConfig
  } catch {
    return null
  }
}

async function loadMerchantServers(apiBase: string, token: string, enabled: boolean) {
  if (!enabled) {
    return [] as Server[]
  }
  try {
    const res = await fetch(`${apiBase}/v1/merchant/servers`, {
      cache: 'no-store',
      headers: { Cookie: `mcp_access_token=${token}` },
    })
    if (!res.ok) {
      return [] as Server[]
    }
    const body = (await res.json()) as { items?: Server[] }
    return body.items || []
  } catch {
    return [] as Server[]
  }
}

export default async function AgentBuilderPage() {
  const cookieStore = await cookies()
  const token = cookieStore.get('mcp_access_token')?.value
  const role = cookieStore.get('mcp_active_role')?.value
  const resolvedRole = role === 'buyer' || role === 'merchant' || role === 'admin' ? role : null

  if (!token || !resolvedRole) {
    redirect('/login?next=/settings/agent-builder')
  }

  const apiBase = getServerApiBase()
  const runtime = await loadRuntimeConfig(apiBase)
  const merchantServers = await loadMerchantServers(apiBase, token, resolvedRole === 'merchant' || resolvedRole === 'admin')
  const n8nURL = runtime?.n8n?.url || ''

  return (
    <AppShell role={resolvedRole}>
      <div className="p-6 space-y-6">
        <div>
          <Text variant="h3" className="mb-2">Agent Builder</Text>
          <Text variant="body" className="text-muted-foreground">
            Use the shared n8n workspace and manage the saved builder configuration for each real MCP server.
          </Text>
        </div>

        <Tabs defaultValue="workspace" className="space-y-4">
          <TabsList>
            <TabsTrigger value="workspace">Builder Workspace</TabsTrigger>
            <TabsTrigger value="servers">Saved Server Configs</TabsTrigger>
          </TabsList>

          <TabsContent value="workspace">
            <Card className="p-6 space-y-4">
              <div className="flex items-start gap-3">
                <Workflow className="w-5 h-5 mt-0.5 text-primary" />
                <div>
                  <Text variant="h6">Launch n8n</Text>
                  <Text variant="small" className="text-muted-foreground">
                    This opens the live n8n builder configured in Admin Integrations. Use it to design real automations, then save the MCP-facing catalog on each server record.
                  </Text>
                </div>
              </div>
              <div className="flex flex-wrap gap-3">
                <Button asChild>
                  <Link href={n8nURL} target="_blank" rel="noopener noreferrer">
                    <ExternalLink className="w-4 h-4 mr-2" />
                    Open n8n In New Tab
                  </Link>
                </Button>
                {(resolvedRole === 'merchant' || resolvedRole === 'admin') ? (
                  <Button asChild variant="outline">
                    <Link href="/merchant/servers/new/import-docker">
                      <Plus className="w-4 h-4 mr-2" />
                      Create New MCP Server
                    </Link>
                  </Button>
                ) : null}
              </div>
              <Text variant="small" className="font-mono break-all">{n8nURL}</Text>
            </Card>
          </TabsContent>

          <TabsContent value="servers">
            <div className="grid gap-4">
              {(resolvedRole === 'merchant' || resolvedRole === 'admin') ? merchantServers.map(server => (
                <Card key={server.id} className="p-5 flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
                  <div className="space-y-1">
                    <Text variant="h6">{server.name}</Text>
                    <Text variant="small" className="text-muted-foreground">{server.description}</Text>
                    <Text variant="caption" className="text-muted-foreground">
                      Scopes: {server.builder?.scopeMappings?.length || server.requiredScopes?.length || 0} · Tools: {server.builder?.toolCatalog?.length || 0}
                    </Text>
                  </div>
                  <Button asChild variant="outline">
                    <Link href={`/merchant/servers/${server.id}/builder`}>Open Builder Config</Link>
                  </Button>
                </Card>
              )) : (
                <Card className="p-6 text-sm text-muted-foreground">
                  Buyer accounts can launch the shared builder workspace but cannot save server build definitions. Switch to a merchant or admin session to manage server builder configs.
                </Card>
              )}
              {(resolvedRole === 'merchant' || resolvedRole === 'admin') && merchantServers.length === 0 ? (
                <Card className="p-6 text-sm text-muted-foreground">
                  No MCP servers exist for this account yet. Create one first, then its builder configuration will appear here.
                </Card>
              ) : null}
            </div>
          </TabsContent>
        </Tabs>
      </div>
    </AppShell>
  )
}
