import type { Metadata } from 'next'
import Link from 'next/link'
import { JsonLd } from '@/components/json-ld'
import { Text } from '@/components/retroui/Text'
import { Badge } from '@/components/ui/badge'
import { Card } from '@/components/ui/card'
import { createBreadcrumbJsonLd, toAbsoluteUrl } from '@/lib/seo'

const title = 'MCP OAuth Explained'
const description =
  'Learn how OAuth fits into MCP installation flows and why secure auth setup should be separate from scopes, billing, and launch actions.'

export const metadata: Metadata = {
  title,
  description,
  alternates: { canonical: '/guides/mcp-oauth-explained' },
  openGraph: { title, description, url: '/guides/mcp-oauth-explained', type: 'article' },
}

export default function MCPOAuthExplainedPage() {
  const jsonLd = [
    createBreadcrumbJsonLd([
      { name: 'Home', path: '/' },
      { name: 'Guides', path: '/guides' },
      { name: title, path: '/guides/mcp-oauth-explained' },
    ]),
    {
      '@context': 'https://schema.org',
      '@type': 'WebPage',
      name: title,
      description,
      url: toAbsoluteUrl('/guides/mcp-oauth-explained'),
    },
  ]

  return (
    <main className="min-h-screen bg-background px-4 py-16 sm:px-6 lg:px-8">
      <JsonLd data={jsonLd} />
      <div className="mx-auto flex max-w-4xl flex-col gap-8">
        <div className="space-y-4">
          <Badge variant="outline">Auth Explainer</Badge>
          <Text variant="h1" className="uppercase sm:text-5xl">{title}</Text>
          <Text variant="h6" className="max-w-3xl text-muted-foreground">
            OAuth gives MCP servers a standard way to authorize access without exposing buyers to
            raw token plumbing during install.
          </Text>
        </div>
        <Card className="border-2 border-foreground p-6">
          <Text variant="h4" className="uppercase">Answer Capsule</Text>
          <Text variant="small" className="mt-3 leading-7 text-muted-foreground">
            OAuth in MCP install flows handles secure authorization and token exchange, while the
            marketplace keeps that work separate from permissions review, billing, and final client setup.
          </Text>
        </Card>
        <Card className="border-2 border-foreground p-6">
          <Text variant="small" className="leading-7 text-muted-foreground">
            Read the <Link className="font-semibold text-foreground underline" href="/guides/mcp-server-installation">install guide</Link>{' '}
            or the <Link className="font-semibold text-foreground underline" href="/guides/x402-pricing-explained">x402 pricing explainer</Link>.
          </Text>
        </Card>
      </div>
    </main>
  )
}
