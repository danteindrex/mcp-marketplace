'use client'

import { useState, useEffect } from 'react'
import Link from 'next/link'
import { ArrowLeft, AlertCircle, Copy } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { AppShell } from '@/components/app-shell'
import { fetchServerAuth } from '@/lib/api-client'
import { toast } from 'sonner'

export default function AuthConfigPage({ params }: { params: { serverId: string } }) {
  const [data, setData] = useState<any>(null)

  useEffect(() => {
    fetchServerAuth(params.serverId).then(setData)
  }, [params.serverId])

  if (!data) return <AppShell role="merchant"><div className="p-6">Loading...</div></AppShell>

  return (
    <AppShell role="merchant">
      <div className="max-w-4xl mx-auto p-6 space-y-6">
        <Link href="/merchant/servers" className="flex items-center gap-2 text-muted-foreground hover:text-foreground transition-colors w-fit"><ArrowLeft className="w-4 h-4" />Back to Servers</Link>
        <div><h1 className="text-3xl font-bold mb-2">Authentication Configuration</h1><p className="text-muted-foreground">Set up OAuth2 and scope requirements for your server</p></div>

        <Card className="p-8 space-y-4">
          <h2 className="text-xl font-bold">OAuth Configuration</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div><p className="text-xs text-muted-foreground mb-1">PKCE Required</p><p className="font-medium">{data.oauth.pkceRequired ? 'Yes' : 'No'}</p></div>
            <div><p className="text-xs text-muted-foreground mb-1">Resource Indicators</p><p className="font-medium">{data.oauth.resourceIndicatorRequired ? 'Required' : 'Optional'}</p></div>
          </div>
          <div><p className="text-xs text-muted-foreground mb-1">Canonical Resource URI</p><p className="font-mono text-sm break-all">{data.oauth.canonicalResourceUri}</p></div>
          <div><p className="text-xs text-muted-foreground mb-2">Registration Modes</p><div className="flex flex-wrap gap-2">{data.oauth.registrationModes.map((mode: string) => <span key={mode} className="text-xs bg-muted px-2 py-1 rounded border border-border">{mode}</span>)}</div></div>
          <Button variant="outline" onClick={() => { navigator.clipboard.writeText(data.oauth.canonicalResourceUri); toast.success('Canonical URI copied') }}><Copy className="w-4 h-4 mr-2" />Copy Canonical URI</Button>
        </Card>

        <Card className="bg-blue-500/10 border-blue-200 dark:border-blue-800 p-6 space-y-3">
          <h3 className="font-semibold text-sm flex items-center gap-2"><AlertCircle className="w-4 h-4" />OAuth Registration Modes</h3>
          <ul className="text-sm space-y-2 text-foreground/80"><li><strong>Pre-Registered:</strong> Clients are pre-registered in your auth server</li><li><strong>CIMD:</strong> Client metadata discovery flow</li><li><strong>DCR:</strong> Dynamic registration for public clients</li></ul>
        </Card>
      </div>
    </AppShell>
  )
}