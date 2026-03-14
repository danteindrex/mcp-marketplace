'use client'

import { use, useEffect, useState } from 'react'
import Link from 'next/link'
import { ArrowLeft } from 'lucide-react'
import { AppShell } from '@/components/app-shell'
import { Text } from '@/components/retroui/Text'
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
          <Text variant="h3" className="mb-2">Agent Builder</Text>
          <Text variant="body" className="text-muted-foreground">
            Configure the MCP tool catalog, scope mapping, and implementation notes for this server.
          </Text>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <Card className="p-6 space-y-4">
            <div className="space-y-2">
              <Text as="label" variant="small">Framework</Text>
              <Input value={framework} onChange={e => setFramework(e.target.value)} />
            </div>
            <div className="space-y-2">
              <Text as="label" variant="small">Template</Text>
              <Input value={template} onChange={e => setTemplate(e.target.value)} />
            </div>
            <div className="space-y-2">
              <Text as="label" variant="small">Instructions</Text>
              <Textarea value={instructions} onChange={e => setInstructions(e.target.value)} rows={8} />
            </div>
          </Card>

          <Card className="p-6 space-y-4">
            <div className="space-y-2">
              <Text as="label" variant="small">Required Scopes</Text>
              <Input
                value={scopeMappings}
                onChange={e => setScopeMappings(e.target.value)}
                placeholder="read:data, write:data"
              />
            </div>
            <div className="space-y-2">
              <Text as="label" variant="small">Tool Catalog JSON</Text>
              <Textarea
                value={toolCatalogJSON}
                onChange={e => setToolCatalogJSON(e.target.value)}
                rows={14}
                className="font-mono text-xs"
              />
            </div>
            <Text variant="caption" className="text-muted-foreground">
              Last edited by {data.lastEditedBy || 'nobody yet'}
              {data.lastEditedAt ? ` at ${new Date(data.lastEditedAt).toLocaleString()}` : ''}
            </Text>
            <Button onClick={handleSave} disabled={saving}>
              {saving ? 'Saving...' : 'Save Builder Config'}
            </Button>
            {message ? <Text variant="small" className="text-muted-foreground">{message}</Text> : null}
          </Card>
        </div>
      </div>
    </AppShell>
  )
}
