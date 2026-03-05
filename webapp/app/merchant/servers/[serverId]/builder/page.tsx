'use client'

import { useEffect, useState } from 'react'
import Link from 'next/link'
import { ArrowLeft } from 'lucide-react'
import { AppShell } from '@/components/app-shell'
import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { fetchServerBuilder } from '@/lib/api-client'

export default function BuilderPage({ params }: { params: { serverId: string } }) {
  const [data, setData] = useState<any>(null)

  useEffect(() => {
    fetchServerBuilder(params.serverId).then(setData)
  }, [params.serverId])

  if (!data) return <AppShell role="merchant"><div className="p-6">Loading...</div></AppShell>

  return (
    <AppShell role="merchant">
      <div className="p-6 max-w-6xl mx-auto space-y-6">
        <Link href="/merchant/servers" className="flex items-center gap-2 text-muted-foreground hover:text-foreground transition-colors w-fit"><ArrowLeft className="w-4 h-4" />Back to Servers</Link>
        <div><h1 className="text-3xl font-bold mb-2">Agent Builder</h1><p className="text-muted-foreground">Framework: {data.framework}</p></div>

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <Card className="p-6 space-y-4">
            <h2 className="text-lg font-semibold">Tool Catalog</h2>
            {(data.toolCatalog || []).map((tool: any) => (
              <div key={tool.name} className="p-3 border border-border rounded-lg">
                <p className="font-medium">{tool.name}</p>
                <p className="text-xs text-muted-foreground mt-1">Input: {JSON.stringify(tool.inputSchema)}</p>
                <p className="text-xs text-muted-foreground">Output: {JSON.stringify(tool.outputSchema)}</p>
              </div>
            ))}
          </Card>

          <Card className="p-6 space-y-4">
            <h2 className="text-lg font-semibold">Scope Mapping</h2>
            <div className="flex flex-wrap gap-2">{(data.scopeMappings || []).map((scope: string) => <span key={scope} className="text-xs bg-muted px-2 py-1 rounded border border-border">{scope}</span>)}</div>
            <Button>Save Builder Config</Button>
          </Card>
        </div>
      </div>
    </AppShell>
  )
}