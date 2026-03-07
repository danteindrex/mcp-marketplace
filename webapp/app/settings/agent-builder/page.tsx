'use client'

import Link from 'next/link'
import { AppShell } from '@/components/app-shell'
import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { ExternalLink, Workflow } from 'lucide-react'

const n8nURL = process.env.NEXT_PUBLIC_N8N_URL || 'http://localhost:5678'

export default function AgentBuilderPage() {
  return (
    <AppShell role="buyer">
      <div className="p-6 space-y-6">
        <div>
          <h1 className="text-3xl font-bold mb-2">Agent Builder</h1>
          <p className="text-muted-foreground">
            Build and manage automations in n8n. The builder opens in a new tab.
          </p>
        </div>

        <Tabs defaultValue="launch" className="space-y-4">
          <TabsList>
            <TabsTrigger value="launch">Open Builder</TabsTrigger>
            <TabsTrigger value="details">Connection Details</TabsTrigger>
          </TabsList>

          <TabsContent value="launch">
            <Card className="p-6 space-y-4">
              <div className="flex items-start gap-3">
                <Workflow className="w-5 h-5 mt-0.5 text-primary" />
                <div>
                  <h2 className="text-lg font-semibold">Launch n8n</h2>
                  <p className="text-sm text-muted-foreground">
                    This opens the n8n builder in a separate browser tab. No iframe is used.
                  </p>
                </div>
              </div>
              <Button asChild>
                <Link href={n8nURL} target="_blank" rel="noopener noreferrer">
                  <ExternalLink className="w-4 h-4 mr-2" />
                  Open n8n In New Tab
                </Link>
              </Button>
            </Card>
          </TabsContent>

          <TabsContent value="details">
            <Card className="p-6 space-y-2">
              <h2 className="text-lg font-semibold">n8n URL</h2>
              <p className="text-sm text-muted-foreground">
                Configure with <code>NEXT_PUBLIC_N8N_URL</code> in webapp environment.
              </p>
              <p className="font-mono text-sm break-all">{n8nURL}</p>
            </Card>
          </TabsContent>
        </Tabs>
      </div>
    </AppShell>
  )
}
