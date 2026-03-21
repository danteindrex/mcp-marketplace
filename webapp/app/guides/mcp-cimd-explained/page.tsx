import type { Metadata } from 'next'
import Link from 'next/link'
import { JsonLd } from '@/components/json-ld'
import { Text } from '@/components/retroui/text'
import { Badge } from '@/components/ui/badge'
import { Card } from '@/components/ui/card'
import { createBreadcrumbJsonLd, toAbsoluteUrl } from '@/lib/seo'

const title = 'MCP CIMD Explained'
const description =
  'Learn what CIMD means in MCP installation flows and why client metadata discovery helps marketplaces create smoother one-click setup.'

export const metadata: Metadata = {
  title,
  description,
  alternates: { canonical: '/guides/mcp-cimd-explained' },
  openGraph: { title, description, url: '/guides/mcp-cimd-explained', type: 'article' },
}

export default function MCPCIMDExplainedPage() {
  const jsonLd = [
    createBreadcrumbJsonLd([
      { name: 'Home', path: '/' },
      { name: 'Guides', path: '/guides' },
      { name: title, path: '/guides/mcp-cimd-explained' },
    ]),
    {
      '@context': 'https://schema.org',
      '@type': 'WebPage',
      name: title,
      description,
      url: toAbsoluteUrl('/guides/mcp-cimd-explained'),
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
            CIMD helps the marketplace understand client-specific metadata requirements before a
            buyer reaches the final install action.
          </Text>
        </div>
        <Card className="border-2 border-foreground p-6">
          <Text variant="h4" className="uppercase">Answer Capsule</Text>
          <Text variant="small" className="mt-3 leading-7 text-muted-foreground">
            In MCP install flows, CIMD lets the platform discover the metadata needed to prepare a
            client-specific connection path instead of asking the buyer to supply that information manually.
          </Text>
        </Card>
        <Card className="border-2 border-foreground p-6">
          <Text variant="h4" className="uppercase">Why buyers benefit</Text>
          <Text variant="small" className="mt-3 leading-7 text-muted-foreground">
            The marketplace can check readiness earlier, separate metadata failures from auth or
            billing failures, and keep the install experience predictable.
          </Text>
        </Card>
        <Card className="border-2 border-foreground p-6">
          <Text variant="small" className="leading-7 text-muted-foreground">
            Continue with the <Link className="font-semibold text-foreground underline" href="/guides/mcp-dcr-explained">DCR explainer</Link>{' '}
            or the <Link className="font-semibold text-foreground underline" href="/guides/mcp-server-installation">install guide</Link>.
          </Text>
        </Card>
      </div>
    </main>
  )
}
