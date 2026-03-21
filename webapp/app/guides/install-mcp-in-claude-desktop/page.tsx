import type { Metadata } from 'next'
import Link from 'next/link'
import { JsonLd } from '@/components/json-ld'
import { Text } from '@/components/retroui/Text'
import { Badge } from '@/components/ui/badge'
import { Card } from '@/components/ui/card'
import { createBreadcrumbJsonLd, toAbsoluteUrl } from '@/lib/seo'

const title = 'How To Install MCP In Claude Desktop'
const description =
  'Learn how MCP Marketplace prepares a Claude Desktop-ready MCP install flow with metadata checks, scope review, and client-specific setup actions.'

export const metadata: Metadata = {
  title,
  description,
  alternates: { canonical: '/guides/install-mcp-in-claude-desktop' },
  openGraph: { title, description, url: '/guides/install-mcp-in-claude-desktop', type: 'article' },
}

export default function InstallMCPInClaudeDesktopGuidePage() {
  const jsonLd = [
    createBreadcrumbJsonLd([
      { name: 'Home', path: '/' },
      { name: 'Guides', path: '/guides' },
      { name: title, path: '/guides/install-mcp-in-claude-desktop' },
    ]),
    {
      '@context': 'https://schema.org',
      '@type': 'HowTo',
      name: title,
      description,
      url: toAbsoluteUrl('/guides/install-mcp-in-claude-desktop'),
      step: [
        { '@type': 'HowToStep', name: 'Open a marketplace listing', text: 'Choose the MCP server you want to connect.' },
        { '@type': 'HowToStep', name: 'Select Claude', text: 'Use the install flow to choose Claude Desktop as the target client.' },
        { '@type': 'HowToStep', name: 'Confirm readiness', text: 'Let the marketplace validate metadata, buyer hub state, and scopes.' },
        { '@type': 'HowToStep', name: 'Resolve payment if needed', text: 'If the listing is paid, settle payment before the final connection step.' },
        { '@type': 'HowToStep', name: 'Launch setup', text: 'Open the generated Claude Desktop install action.' },
      ],
    },
  ]

  return (
    <main className="min-h-screen bg-background px-4 py-16 sm:px-6 lg:px-8">
      <JsonLd data={jsonLd} />
      <div className="mx-auto flex max-w-4xl flex-col gap-8">
        <div className="space-y-4">
          <Badge variant="outline">Claude Desktop Guide</Badge>
          <Text variant="h1" className="uppercase sm:text-5xl">{title}</Text>
          <Text variant="h6" className="max-w-3xl text-muted-foreground">
            MCP Marketplace keeps Claude Desktop installation clear by validating the protocol
            pieces before it asks the buyer to finish the client setup step.
          </Text>
        </div>
        <Card className="border-2 border-foreground p-6">
          <Text variant="h4" className="uppercase">Answer Capsule</Text>
          <Text variant="small" className="mt-3 leading-7 text-muted-foreground">
            To install MCP in Claude Desktop through MCP Marketplace, open a listing, choose Claude
            in the install flow, confirm metadata and scopes, then launch the generated
            Claude-ready install action.
          </Text>
        </Card>
        <Card className="border-2 border-foreground p-6">
          <Text variant="h4" className="uppercase">Next Step</Text>
          <Text variant="small" className="mt-3 leading-7 text-muted-foreground">
            Browse the <Link className="font-semibold text-foreground underline" href="/marketplace">marketplace</Link>{' '}
            or read the <Link className="font-semibold text-foreground underline" href="/guides/mcp-server-installation">general install guide</Link>.
          </Text>
        </Card>
      </div>
    </main>
  )
}
