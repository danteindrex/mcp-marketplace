import type { Metadata } from 'next'
import Link from 'next/link'
import { JsonLd } from '@/components/json-ld'
import { Text } from '@/components/retroui/Text'
import { Badge } from '@/components/ui/badge'
import { Card } from '@/components/ui/card'

const title = 'What Is Model Context Protocol'
const description =
  'Learn what Model Context Protocol means, why MCP matters for AI tool use, and how MCP Marketplace helps buyers discover and install MCP servers.'

export const metadata: Metadata = {
  title,
  description,
  alternates: { canonical: '/guides/what-is-mcp' },
  openGraph: { title, description, url: '/guides/what-is-mcp', type: 'article' },
}

export default function WhatIsMCPPage() {
  const jsonLd = {
    '@context': 'https://schema.org',
    '@type': 'FAQPage',
    mainEntity: [
      {
        '@type': 'Question',
        name: 'What is Model Context Protocol?',
        acceptedAnswer: {
          '@type': 'Answer',
          text: 'Model Context Protocol is a standard for connecting AI clients to tools and data sources through structured capability, auth, and transport contracts.',
        },
      },
    ],
  }

  return (
    <main className="min-h-screen bg-background px-4 py-16 sm:px-6 lg:px-8">
      <JsonLd data={jsonLd} />
      <div className="mx-auto flex max-w-4xl flex-col gap-8">
        <div className="space-y-4">
          <Badge variant="outline">MCP Explainer</Badge>
          <Text variant="h1" className="uppercase sm:text-5xl">{title}</Text>
          <Text variant="h6" className="max-w-3xl text-muted-foreground">
            MCP gives AI clients a consistent way to discover tools, negotiate permissions, and
            connect to remote or local capabilities.
          </Text>
        </div>
        <Card className="border-2 border-foreground p-6">
          <Text variant="h4" className="uppercase">Answer Capsule</Text>
          <Text variant="small" className="mt-3 leading-7 text-muted-foreground">
            Model Context Protocol is a standard for connecting AI applications to tools and data
            sources through predictable discovery, auth, and capability contracts.
          </Text>
        </Card>
        <Card className="border-2 border-foreground p-6">
          <Text variant="h4" className="uppercase">Why it matters</Text>
          <Text variant="small" className="mt-3 leading-7 text-muted-foreground">
            Without a shared protocol, every client-to-tool integration becomes custom work. MCP
            reduces that fragmentation so marketplaces can publish installable servers that work
            across multiple AI clients.
          </Text>
        </Card>
        <Card className="border-2 border-foreground p-6">
          <Text variant="h4" className="uppercase">Next Step</Text>
          <Text variant="small" className="mt-3 leading-7 text-muted-foreground">
            Read the <Link className="font-semibold text-foreground underline" href="/guides/mcp-server-installation">installation guide</Link>{' '}
            or browse live listings in the <Link className="font-semibold text-foreground underline" href="/marketplace">marketplace</Link>.
          </Text>
        </Card>
      </div>
    </main>
  )
}
