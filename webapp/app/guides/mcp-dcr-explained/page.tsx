import type { Metadata } from 'next'
import Link from 'next/link'
import { Text } from '@/components/retroui/Text'
import { Badge } from '@/components/ui/badge'
import { Card } from '@/components/ui/card'

const title = 'MCP DCR Explained'
const description =
  'Learn what DCR means in MCP installation flows and why dynamic client registration matters for public client setup.'

export const metadata: Metadata = {
  title,
  description,
  alternates: { canonical: '/guides/mcp-dcr-explained' },
  openGraph: { title, description, url: '/guides/mcp-dcr-explained', type: 'article' },
}

export default function MCPDCRExplainedPage() {
  return (
    <main className="min-h-screen bg-background px-4 py-16 sm:px-6 lg:px-8">
      <div className="mx-auto flex max-w-4xl flex-col gap-8">
        <div className="space-y-4">
          <Badge variant="outline">Auth Explainer</Badge>
          <Text variant="h1" className="uppercase sm:text-5xl">{title}</Text>
          <Text variant="h6" className="max-w-3xl text-muted-foreground">
            DCR gives public clients a safer way to register and connect without pushing registration
            complexity onto the buyer.
          </Text>
        </div>
        <Card className="border-2 border-foreground p-6">
          <Text variant="h4" className="uppercase">Answer Capsule</Text>
          <Text variant="small" className="mt-3 leading-7 text-muted-foreground">
            In MCP install flows, DCR allows the platform to create or negotiate client registration
            details dynamically so buyers can complete setup without manual client provisioning.
          </Text>
        </Card>
        <Card className="border-2 border-foreground p-6">
          <Text variant="small" className="leading-7 text-muted-foreground">
            Pair this with the <Link className="font-semibold text-foreground underline" href="/guides/mcp-oauth-explained">OAuth explainer</Link>{' '}
            or the <Link className="font-semibold text-foreground underline" href="/guides/mcp-cimd-explained">CIMD explainer</Link>.
          </Text>
        </Card>
      </div>
    </main>
  )
}
