import type { Metadata } from 'next'
import Link from 'next/link'
import { JsonLd } from '@/components/json-ld'
import { Badge } from '@/components/ui/badge'
import { Card } from '@/components/ui/card'

const title = 'Hosted Vs Local MCP Servers'
const description =
  'Understand the difference between hosted and local MCP servers, when each model fits best, and how MCP Marketplace supports both install paths.'

export const metadata: Metadata = {
  title,
  description,
  alternates: {
    canonical: '/guides/hosted-vs-local-mcp-servers',
  },
  openGraph: {
    title,
    description,
    url: '/guides/hosted-vs-local-mcp-servers',
    type: 'article',
  },
}

export default function HostedVsLocalMCPServersPage() {
  const jsonLd = {
    '@context': 'https://schema.org',
    '@type': 'FAQPage',
    mainEntity: [
      {
        '@type': 'Question',
        name: 'What is a hosted MCP server?',
        acceptedAnswer: {
          '@type': 'Answer',
          text: 'A hosted MCP server runs on managed infrastructure and buyers connect to it remotely through the marketplace install flow.',
        },
      },
      {
        '@type': 'Question',
        name: 'What is a local MCP server?',
        acceptedAnswer: {
          '@type': 'Answer',
          text: 'A local MCP server runs on the buyer machine or local environment and may require a local bridge or machine-specific install action.',
        },
      },
    ],
  }

  return (
    <main className="min-h-screen bg-background px-4 py-16 sm:px-6 lg:px-8">
      <JsonLd data={jsonLd} />
      <div className="mx-auto flex max-w-4xl flex-col gap-8">
        <div className="space-y-4">
          <Badge variant="outline">Compatibility Guide</Badge>
          <h1 className="text-4xl font-black uppercase sm:text-5xl">{title}</h1>
          <p className="max-w-3xl text-lg text-muted-foreground">
            MCP Marketplace supports both hosted and local MCP listings, but buyers need the model
            explained in plain language before they click install.
          </p>
        </div>

        <Card className="border-2 border-foreground p-6">
          <h2 className="text-2xl font-black uppercase">Answer Capsule</h2>
          <p className="mt-3 text-sm leading-7 text-muted-foreground">
            Hosted MCP servers run remotely and usually install faster, while local MCP servers run
            on the buyer machine and may require a local bridge or client-specific machine setup.
          </p>
        </Card>

        <section className="grid gap-4 md:grid-cols-2">
          <Card className="border-2 border-foreground p-6">
            <h2 className="text-xl font-black uppercase">Hosted MCP</h2>
            <p className="mt-3 text-sm leading-7 text-muted-foreground">
              Best when the buyer wants fast onboarding, managed uptime, and fewer machine-specific
              setup steps. The marketplace can often complete the flow through remote auth and a
              client launch action.
            </p>
          </Card>
          <Card className="border-2 border-foreground p-6">
            <h2 className="text-xl font-black uppercase">Local MCP</h2>
            <p className="mt-3 text-sm leading-7 text-muted-foreground">
              Best when the buyer needs local execution, device-bound tools, or private network
              access. The marketplace keeps the flow clear by separating the local bridge step from
              metadata, scopes, and payment.
            </p>
          </Card>
        </section>

        <Card className="border-2 border-foreground p-6">
          <h2 className="text-2xl font-black uppercase">Next Step</h2>
          <p className="mt-3 text-sm leading-7 text-muted-foreground">
            Open the <Link className="font-semibold text-foreground underline" href="/marketplace">marketplace</Link>{' '}
            to compare live listings or read the <Link className="font-semibold text-foreground underline" href="/guides/mcp-server-installation">install guide</Link>{' '}
            for the end-to-end setup flow.
          </p>
        </Card>
      </div>
    </main>
  )
}
