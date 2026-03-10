'use client'

import { use, useEffect, useState } from 'react'
import Link from 'next/link'
import { ArrowLeft } from 'lucide-react'
import { AppShell } from '@/components/app-shell'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import {
  fetchServerBuilder,
  updateServerBuilder,
  type ServerBuilderConfig,
  type ServerBuilderTool,
} from '@/lib/api-client'

function formatToolCatalog(tools: ServerBuilderTool[]) {
  return JSON.stringify(tools || [], null, 2)
}

export default function BuilderPage({ params }: { params: Promise<{ serverId: string }> }) {
  const { serverId } = use(params)
  const [data, setData] = useState<ServerBuilderConfig | null>(null)
  const [framework, setFramework] = useState('')
  const [template, setTemplate] = useState('')
  const [instructions, setInstructions] = useState('')
  const [scopeMappings, setScopeMappings] = useState('')
  const [toolCatalogJSON, setToolCatalogJSON] = useState('[]')
  const [saving, setSaving] = useState(false)
  const [message, setMessage] = useState('')

  useEffect(() => {
    fetchServerBuilder(serverId).then(payload => {
      setData(payload)
      setFramework(payload.framework || '')
      setTemplate(payload.template || '')
      setInstructions(payload.instructions || '')
      setScopeMappings((payload.scopeMappings || []).join(', '))
      setToolCatalogJSON(formatToolCatalog(payload.toolCatalog || []))
    })
  }, [serverId])

  async function handleSave() {
    setSaving(true)
    setMessage('')
    try {
      const parsedTools = JSON.parse(toolCatalogJSON) as ServerBuilderTool[]
      const next = await updateServerBuilder(serverId, {
        framework: framework.trim(),
        template: template.trim(),
        instructions: instructions.trim(),
        scopeMappings: scopeMappings.split(',').map(scope => scope.trim()).filter(Boolean),
        toolCatalog: parsedTools,
      })
      setData(next)
      setFramework(next.framework || '')
      setTemplate(next.template || '')
      setInstructions(next.instructions || '')
      setScopeMappings((next.scopeMappings || []).join(', '))
      setToolCatalogJSON(formatToolCatalog(next.toolCatalog || []))
      setMessage('Builder configuration saved.')
    } catch (error) {
      setMessage(error instanceof Error ? error.message : 'Failed to save builder configuration')
    } finally {
      setSaving(false)
    }
  }

  if (!data) {
    return <AppShell role="merchant"><div className="p-6">Loading...</div></AppShell>
  }

  return (
    <AppShell role="merchant">
      <div className="p-6 max-w-6xl mx-auto space-y-6">
        <Link
          href="/merchant/servers"
          className="flex items-center gap-2 text-muted-foreground hover:text-foreground transition-colors w-fit"
        >
          <ArrowLeft className="w-4 h-4" />
          Back to Servers
        </Link>

        <div>
          <h1 className="text-3xl font-bold mb-2">Agent Builder</h1>
          <p className="text-muted-foreground">
            Configure the MCP tool catalog, scope mapping, and implementation notes for this server.
          </p>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <Card className="p-6 space-y-4">
            <div className="space-y-2">
              <label className="text-sm font-medium">Framework</label>
              <Input value={framework} onChange={e => setFramework(e.target.value)} />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">Template</label>
              <Input value={template} onChange={e => setTemplate(e.target.value)} />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">Instructions</label>
              <Textarea value={instructions} onChange={e => setInstructions(e.target.value)} rows={8} />
            </div>
          </Card>

          <Card className="p-6 space-y-4">
            <div className="space-y-2">
              <label className="text-sm font-medium">Required Scopes</label>
              <Input
                value={scopeMappings}
                onChange={e => setScopeMappings(e.target.value)}
                placeholder="read:data, write:data"
              />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">Tool Catalog JSON</label>
              <Textarea
                value={toolCatalogJSON}
                onChange={e => setToolCatalogJSON(e.target.value)}
                rows={14}
                className="font-mono text-xs"
              />
            </div>
            <p className="text-xs text-muted-foreground">
              Last edited by {data.lastEditedBy || 'nobody yet'}
              {data.lastEditedAt ? ` at ${new Date(data.lastEditedAt).toLocaleString()}` : ''}
            </p>
            <Button onClick={handleSave} disabled={saving}>
              {saving ? 'Saving...' : 'Save Builder Config'}
            </Button>
            {message ? <p className="text-sm text-muted-foreground">{message}</p> : null}
          </Card>
        </div>
      </div>
    </AppShell>
  )
}
