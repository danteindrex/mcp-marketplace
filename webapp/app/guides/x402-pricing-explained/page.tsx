import type { Metadata } from 'next'
import Link from 'next/link'
import { Text } from '@/components/retroui/Text'
import { Badge } from '@/components/ui/badge'
import { Card } from '@/components/ui/card'

const title = 'x402 Pricing Explained'
const description =
  'Learn how x402 pricing works for MCP servers and why MCP Marketplace separates payment from auth and install readiness.'

export const metadata: Metadata = {
  title,
  description,
  alternates: { canonical: '/guides/x402-pricing-explained' },
  openGraph: { title, description, url: '/guides/x402-pricing-explained', type: 'article' },
}

export default function X402PricingExplainedPage() {
  return (
    <main className="min-h-screen bg-background px-4 py-16 sm:px-6 lg:px-8">
      <div className="mx-auto flex max-w-4xl flex-col gap-8">
        <div className="space-y-4">
          <Badge variant="outline">Pricing Explainer</Badge>
          <Text variant="h1" className="uppercase sm:text-5xl">{title}</Text>
          <Text variant="h6" className="max-w-3xl text-muted-foreground">
            x402 pricing gives MCP servers a payment-aware flow, and MCP Marketplace keeps that
            step explicit so buyers know when billing is the only blocker.
          </Text>
        </div>
        <Card className="border-2 border-foreground p-6">
          <Text variant="h4" className="uppercase">Answer Capsule</Text>
          <Text variant="small" className="mt-3 leading-7 text-muted-foreground">
            x402 pricing adds a payment challenge to the install or tool-use path, and MCP Marketplace
            handles that challenge as a distinct billing step instead of mixing it with auth or scope errors.
          </Text>
        </Card>
        <Card className="border-2 border-foreground p-6">
          <Text variant="small" className="leading-7 text-muted-foreground">
            Compare with <Link className="font-semibold text-foreground underline" href="/pricing">pricing</Link>{' '}
            or return to the <Link className="font-semibold text-foreground underline" href="/guides/mcp-server-installation">install guide</Link>.
          </Text>
        </Card>
      </div>
    </main>
  )
}
