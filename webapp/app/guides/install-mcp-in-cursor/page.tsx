import type { Metadata } from 'next'
import Link from 'next/link'
import { JsonLd } from '@/components/json-ld'
import { Text } from '@/components/retroui/text'
import { Badge } from '@/components/ui/badge'
import { Card } from '@/components/ui/card'
import { createBreadcrumbJsonLd, toAbsoluteUrl } from '@/lib/seo'

const title = 'How To Install MCP In Cursor'
const description =
  'Learn how MCP Marketplace prepares a Cursor-ready MCP install flow with metadata checks, permissions review, and client-specific setup.'

export const metadata: Metadata = {
  title,
  description,
  alternates: { canonical: '/guides/install-mcp-in-cursor' },
  openGraph: { title, description, url: '/guides/install-mcp-in-cursor', type: 'article' },
}

export default function InstallMCPInCursorGuidePage() {
  const jsonLd = [
    createBreadcrumbJsonLd([
      { name: 'Home', path: '/' },
      { name: 'Guides', path: '/guides' },
      { name: title, path: '/guides/install-mcp-in-cursor' },
    ]),
    {
      '@context': 'https://schema.org',
      '@type': 'HowTo',
      name: title,
      description,
      url: toAbsoluteUrl('/guides/install-mcp-in-cursor'),
      step: [
        { '@type': 'HowToStep', name: 'Pick a listing', text: 'Choose the MCP server you want from the marketplace.' },
        { '@type': 'HowToStep', name: 'Select Cursor', text: 'Choose Cursor as the client target in the guided flow.' },
        { '@type': 'HowToStep', name: 'Review permissions', text: 'Confirm the required scopes and install readiness state.' },
        { '@type': 'HowToStep', name: 'Resolve billing when needed', text: 'Only paid servers trigger wallet or x402 settlement.' },
        { '@type': 'HowToStep', name: 'Launch Cursor setup', text: 'Use the generated action to finish setup in Cursor.' },
      ],
    },
  ]

  return (
    <main className="min-h-screen bg-background px-4 py-16 sm:px-6 lg:px-8">
      <JsonLd data={jsonLd} />
      <div className="mx-auto flex max-w-4xl flex-col gap-8">
        <div className="space-y-4">
          <Badge variant="outline">Cursor Guide</Badge>
          <Text variant="h1" className="uppercase sm:text-5xl">{title}</Text>
          <Text variant="h6" className="max-w-3xl text-muted-foreground">
            Cursor installation stays straightforward when the marketplace handles readiness checks
            before the buyer reaches the final setup step.
          </Text>
        </div>
        <Card className="border-2 border-foreground p-6">
          <Text variant="h4" className="uppercase">Answer Capsule</Text>
          <Text variant="small" className="mt-3 leading-7 text-muted-foreground">
            To install MCP in Cursor through MCP Marketplace, open a listing, choose Cursor,
            review scopes and readiness, settle payment if required, and launch the generated
            Cursor-specific install action.
          </Text>
        </Card>
        <Card className="border-2 border-foreground p-6">
          <Text variant="h4" className="uppercase">Next Step</Text>
          <Text variant="small" className="mt-3 leading-7 text-muted-foreground">
            Continue to the <Link className="font-semibold text-foreground underline" href="/marketplace">marketplace</Link>{' '}
            or compare with the <Link className="font-semibold text-foreground underline" href="/guides/install-mcp-in-vs-code">VS Code install path</Link>.
          </Text>
        </Card>
      </div>
    </main>
  )
}
