import type { Metadata } from 'next'
import Link from 'next/link'
import { JsonLd } from '@/components/json-ld'
import { Text } from '@/components/retroui/Text'
import { Badge } from '@/components/ui/badge'
import { Card } from '@/components/ui/card'
import { getSiteUrl } from '@/lib/site'

const title = 'How To Install MCP In Codex'
const description =
  'Learn how MCP Marketplace prepares a Codex-ready MCP install flow with metadata checks, scope review, and client-specific launch actions.'

export const metadata: Metadata = {
  title,
  description,
  alternates: {
    canonical: '/guides/install-mcp-in-codex',
  },
  openGraph: {
    title,
    description,
    url: '/guides/install-mcp-in-codex',
    type: 'article',
  },
}

export default function InstallMCPInCodexGuidePage() {
  const siteUrl = getSiteUrl()
  const jsonLd = {
    '@context': 'https://schema.org',
    '@type': 'HowTo',
    name: title,
    description,
    url: `${siteUrl}/guides/install-mcp-in-codex`,
    step: [
      { '@type': 'HowToStep', name: 'Open a listing', text: 'Choose an MCP server from the marketplace and open its detail page.' },
      { '@type': 'HowToStep', name: 'Select Codex', text: 'Use the install flow to choose Codex so the marketplace prepares the Codex-specific action.' },
      { '@type': 'HowToStep', name: 'Verify metadata', text: 'The marketplace validates install metadata and buyer hub readiness before continuing.' },
      { '@type': 'HowToStep', name: 'Approve scopes and payment', text: 'Review permissions and settle payment only if the server is not free.' },
      { '@type': 'HowToStep', name: 'Launch the install action', text: 'Run the generated Codex install action or local bridge step to finish setup.' },
    ],
  }

  return (
    <main className="min-h-screen bg-background px-4 py-16 sm:px-6 lg:px-8">
      <JsonLd data={jsonLd} />
      <div className="mx-auto flex max-w-4xl flex-col gap-8">
        <div className="space-y-4">
          <Badge variant="outline">Codex Guide</Badge>
          <Text variant="h1" className="uppercase sm:text-5xl">{title}</Text>
          <Text variant="h6" className="max-w-3xl text-muted-foreground">
            MCP Marketplace makes Codex installation feel predictable by preparing the install
            action after metadata, scopes, and payment requirements are resolved.
          </Text>
        </div>

        <Card className="border-2 border-foreground p-6">
          <Text variant="h4" className="uppercase">Answer Capsule</Text>
          <Text variant="small" className="mt-3 leading-7 text-muted-foreground">
            To install MCP in Codex through MCP Marketplace, open a listing, choose Codex in the
            guided install flow, let the platform validate metadata and scopes, then launch the
            generated Codex-specific install action.
          </Text>
        </Card>

        <Card className="border-2 border-foreground p-6">
          <Text variant="h4" className="uppercase">Why this stays seamless</Text>
          <Text variant="small" className="mt-3 leading-7 text-muted-foreground">
            Codex buyers should not have to manually assemble connection metadata, auth state, or
            payment context. The marketplace stages those checks first so the last step can stay
            focused on the actual Codex install command or launch action.
          </Text>
        </Card>

        <Card className="border-2 border-foreground p-6">
          <Text variant="h4" className="uppercase">Next Step</Text>
          <Text variant="small" className="mt-3 leading-7 text-muted-foreground">
            Browse the <Link className="font-semibold text-foreground underline" href="/marketplace">marketplace</Link>{' '}
            or read the general <Link className="font-semibold text-foreground underline" href="/guides/mcp-server-installation">MCP install guide</Link>.
          </Text>
        </Card>
      </div>
    </main>
  )
}
