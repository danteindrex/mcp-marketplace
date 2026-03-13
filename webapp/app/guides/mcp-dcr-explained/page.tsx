import type { Metadata } from 'next'
import Link from 'next/link'
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
          <h1 className="text-4xl font-black uppercase sm:text-5xl">{title}</h1>
          <p className="max-w-3xl text-lg text-muted-foreground">
            DCR gives public clients a safer way to register and connect without pushing registration
            complexity onto the buyer.
          </p>
        </div>
        <Card className="border-2 border-foreground p-6">
          <h2 className="text-2xl font-black uppercase">Answer Capsule</h2>
          <p className="mt-3 text-sm leading-7 text-muted-foreground">
            In MCP install flows, DCR allows the platform to create or negotiate client registration
            details dynamically so buyers can complete setup without manual client provisioning.
          </p>
        </Card>
        <Card className="border-2 border-foreground p-6">
          <p className="text-sm leading-7 text-muted-foreground">
            Pair this with the <Link className="font-semibold text-foreground underline" href="/guides/mcp-oauth-explained">OAuth explainer</Link>{' '}
            or the <Link className="font-semibold text-foreground underline" href="/guides/mcp-cimd-explained">CIMD explainer</Link>.
          </p>
        </Card>
      </div>
    </main>
  )
}
