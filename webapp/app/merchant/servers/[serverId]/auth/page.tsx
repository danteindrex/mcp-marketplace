'use client'

import { useState, useEffect } from 'react'
import { use } from 'react'
import Link from 'next/link'
import { ArrowLeft, AlertCircle, Copy } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { AppShell } from '@/components/app-shell'
import { Text } from '@/components/retroui/Text'
import { fetchServerAuth } from '@/lib/api-client'
import { toast } from 'sonner'

export default function AuthConfigPage({ params }: { params: Promise<{ serverId: string }> }) {
  const { serverId } = use(params)
  const [data, setData] = useState<any>(null)

  useEffect(() => {
    fetchServerAuth(serverId).then(setData)
  }, [serverId])

  if (!data) return <AppShell role="merchant"><div className="p-6">Loading...</div></AppShell>

  return (
    <AppShell role="merchant">
      <div className="max-w-4xl mx-auto p-6 space-y-6">
        <Link href="/merchant/servers" className="flex items-center gap-2 text-muted-foreground hover:text-foreground transition-colors w-fit"><ArrowLeft className="w-4 h-4" />Back to Servers</Link>
        <div><Text variant="h3" className="mb-2">Authentication Configuration</Text><Text variant="body" className="text-muted-foreground">Set up OAuth2 and scope requirements for your server</Text></div>

        <Card className="p-8 space-y-4">
          <Text variant="h5">OAuth Configuration</Text>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div><Text variant="caption" className="mb-1 text-muted-foreground">PKCE Required</Text><Text variant="body">{data.oauth.pkceRequired ? 'Yes' : 'No'}</Text></div>
            <div><Text variant="caption" className="mb-1 text-muted-foreground">Resource Indicators</Text><Text variant="body">{data.oauth.resourceIndicatorRequired ? 'Required' : 'Optional'}</Text></div>
          </div>
          <div><Text variant="caption" className="mb-1 text-muted-foreground">Canonical Resource URI</Text><Text variant="small" className="font-mono break-all">{data.oauth.canonicalResourceUri}</Text></div>
          <div><Text variant="caption" className="mb-2 text-muted-foreground">Registration Modes</Text><div className="flex flex-wrap gap-2">{data.oauth.registrationModes.map((mode: string) => <span key={mode} className="text-xs bg-muted px-2 py-1 rounded border border-border">{mode}</span>)}</div></div>
          <Button variant="outline" onClick={() => { navigator.clipboard.writeText(data.oauth.canonicalResourceUri); toast.success('Canonical URI copied') }}><Copy className="w-4 h-4 mr-2" />Copy Canonical URI</Button>
        </Card>

        <Card className="bg-blue-500/10 border-blue-200 dark:border-blue-800 p-6 space-y-3">
          <Text variant="small" className="flex items-center gap-2"><AlertCircle className="w-4 h-4" />OAuth Registration Modes</Text>
          <ul className="text-sm space-y-2 text-foreground/80"><li><strong>Pre-Registered:</strong> Clients are pre-registered in your auth server</li><li><strong>CIMD:</strong> Client metadata discovery flow</li><li><strong>DCR:</strong> Dynamic registration for public clients</li></ul>
        </Card>
      </div>
    </AppShell>
  )
}
