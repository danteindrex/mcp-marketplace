import type { Metadata } from 'next'
import Link from 'next/link'
import { JsonLd } from '@/components/json-ld'
import { Text } from '@/components/retroui/Text'
import { Badge } from '@/components/ui/badge'
import { Card } from '@/components/ui/card'
import { createBreadcrumbJsonLd, toAbsoluteUrl } from '@/lib/seo'

const title = 'How To Install MCP In VS Code'
const description =
  'Learn how MCP Marketplace prepares a VS Code-ready MCP install flow with metadata verification, scope review, and client-specific setup.'

export const metadata: Metadata = {
  title,
  description,
  alternates: { canonical: '/guides/install-mcp-in-vs-code' },
  openGraph: { title, description, url: '/guides/install-mcp-in-vs-code', type: 'article' },
}

export default function InstallMCPInVSCodeGuidePage() {
  const jsonLd = [
    createBreadcrumbJsonLd([
      { name: 'Home', path: '/' },
      { name: 'Guides', path: '/guides' },
      { name: title, path: '/guides/install-mcp-in-vs-code' },
    ]),
    {
      '@context': 'https://schema.org',
      '@type': 'HowTo',
      name: title,
      description,
      url: toAbsoluteUrl('/guides/install-mcp-in-vs-code'),
      step: [
        { '@type': 'HowToStep', name: 'Choose a listing', text: 'Open the server you want to use from the marketplace.' },
        { '@type': 'HowToStep', name: 'Select VS Code', text: 'Pick VS Code so the marketplace prepares the right install action.' },
        { '@type': 'HowToStep', name: 'Check readiness', text: 'Let the platform validate metadata, scopes, and buyer hub state.' },
        { '@type': 'HowToStep', name: 'Settle payment if required', text: 'Paid servers expose payment as a separate install step.' },
        { '@type': 'HowToStep', name: 'Open VS Code setup', text: 'Launch the generated VS Code action to finish the connection.' },
      ],
    },
  ]

  return (
    <main className="min-h-screen bg-background px-4 py-16 sm:px-6 lg:px-8">
      <JsonLd data={jsonLd} />
      <div className="mx-auto flex max-w-4xl flex-col gap-8">
        <div className="space-y-4">
          <Badge variant="outline">VS Code Guide</Badge>
          <Text variant="h1" className="uppercase sm:text-5xl">{title}</Text>
          <Text variant="h6" className="max-w-3xl text-muted-foreground">
            MCP Marketplace makes the VS Code path easier to trust by validating discovery and auth
            details before the buyer reaches the final install step.
          </Text>
        </div>
        <Card className="border-2 border-foreground p-6">
          <Text variant="h4" className="uppercase">Answer Capsule</Text>
          <Text variant="small" className="mt-3 leading-7 text-muted-foreground">
            To install MCP in VS Code through MCP Marketplace, open a listing, select VS Code in
            the guided flow, confirm readiness and scopes, then launch the generated VS Code setup action.
          </Text>
        </Card>
        <Card className="border-2 border-foreground p-6">
          <Text variant="h4" className="uppercase">Next Step</Text>
          <Text variant="small" className="mt-3 leading-7 text-muted-foreground">
            Open the <Link className="font-semibold text-foreground underline" href="/marketplace">marketplace</Link>{' '}
            or compare with the <Link className="font-semibold text-foreground underline" href="/guides/install-mcp-in-cursor">Cursor install path</Link>.
          </Text>
        </Card>
      </div>
    </main>
  )
}
