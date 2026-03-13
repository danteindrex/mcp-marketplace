import type { Metadata } from 'next'
import Link from 'next/link'
import { Badge } from '@/components/ui/badge'
import { Card } from '@/components/ui/card'

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
  return (
    <main className="min-h-screen bg-background px-4 py-16 sm:px-6 lg:px-8">
      <div className="mx-auto flex max-w-4xl flex-col gap-8">
        <div className="space-y-4">
          <Badge variant="outline">Auth Explainer</Badge>
          <h1 className="text-4xl font-black uppercase sm:text-5xl">{title}</h1>
          <p className="max-w-3xl text-lg text-muted-foreground">
            OAuth gives MCP servers a standard way to authorize access without exposing buyers to
            raw token plumbing during install.
          </p>
        </div>
        <Card className="border-2 border-foreground p-6">
          <h2 className="text-2xl font-black uppercase">Answer Capsule</h2>
          <p className="mt-3 text-sm leading-7 text-muted-foreground">
            OAuth in MCP install flows handles secure authorization and token exchange, while the
            marketplace keeps that work separate from permissions review, billing, and final client setup.
          </p>
        </Card>
        <Card className="border-2 border-foreground p-6">
          <p className="text-sm leading-7 text-muted-foreground">
            Read the <Link className="font-semibold text-foreground underline" href="/guides/mcp-server-installation">install guide</Link>{' '}
            or the <Link className="font-semibold text-foreground underline" href="/guides/x402-pricing-explained">x402 pricing explainer</Link>.
          </p>
        </Card>
      </div>
    </main>
  )
}
