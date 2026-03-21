'use client'

import { useState } from 'react'
import Link from 'next/link'
import { AppShell } from '@/components/app-shell'
import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group'
import { Text } from '@/components/retroui/text'

export default function MerchantOnboardingPage() {
  const [sourceType, setSourceType] = useState('docker')

  return (
    <AppShell role="merchant">
      <div className="p-6 max-w-4xl mx-auto space-y-6">
        <div>
          <Text variant="h3" className="mb-2">Merchant Onboarding</Text>
          <Text variant="body" className="text-muted-foreground">Publish an MCP server with OAuth, pricing, and deployment ready defaults.</Text>
        </div>

        <Card className="p-6 space-y-4">
          <Text variant="h5">1. Import Source</Text>
          <RadioGroup value={sourceType} onValueChange={setSourceType}>
            <div className="flex items-center space-x-3 p-3 border border-border rounded-lg"><RadioGroupItem value="docker" id="docker" /><Label htmlFor="docker">Docker Hub Image</Label></div>
            <div className="flex items-center space-x-3 p-3 border border-border rounded-lg"><RadioGroupItem value="custom" id="custom" /><Label htmlFor="custom">Custom Registry Image</Label></div>
          </RadioGroup>
          <div>
            <Label className="mb-2 block">Image Reference</Label>
            <Input placeholder={sourceType === 'docker' ? 'org/image:tag' : 'registry.example.com/org/image:tag'} />
          </div>
        </Card>

        <Card className="p-6 space-y-4">
          <Text variant="h5">2. Server Metadata</Text>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div><Label className="mb-2 block">Server Name</Label><Input placeholder="My MCP Server" /></div>
            <div><Label className="mb-2 block">Slug</Label><Input placeholder="my-mcp-server" /></div>
          </div>
          <div><Label className="mb-2 block">Canonical Resource URI</Label><Input placeholder="https://mcp.yourplatform.com/resource/my-server" /></div>
        </Card>

        <div className="flex gap-3">
          <Button variant="outline" asChild><Link href="/merchant/servers">Cancel</Link></Button>
          <Button asChild><Link href="/merchant/servers/new/import-docker">Continue to Import</Link></Button>
        </div>
      </div>
    </AppShell>
  )
}
