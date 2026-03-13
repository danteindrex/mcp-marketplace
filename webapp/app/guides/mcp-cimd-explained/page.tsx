import type { Metadata } from 'next'
import Link from 'next/link'
import { Badge } from '@/components/ui/badge'
import { Card } from '@/components/ui/card'

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
  return (
    <main className="min-h-screen bg-background px-4 py-16 sm:px-6 lg:px-8">
      <div className="mx-auto flex max-w-4xl flex-col gap-8">
        <div className="space-y-4">
          <Badge variant="outline">Auth Explainer</Badge>
          <h1 className="text-4xl font-black uppercase sm:text-5xl">{title}</h1>
          <p className="max-w-3xl text-lg text-muted-foreground">
            CIMD helps the marketplace understand client-specific metadata requirements before a
            buyer reaches the final install action.
          </p>
        </div>
        <Card className="border-2 border-foreground p-6">
          <h2 className="text-2xl font-black uppercase">Answer Capsule</h2>
          <p className="mt-3 text-sm leading-7 text-muted-foreground">
            In MCP install flows, CIMD lets the platform discover the metadata needed to prepare a
            client-specific connection path instead of asking the buyer to supply that information manually.
          </p>
        </Card>
        <Card className="border-2 border-foreground p-6">
          <h2 className="text-2xl font-black uppercase">Why buyers benefit</h2>
          <p className="mt-3 text-sm leading-7 text-muted-foreground">
            The marketplace can check readiness earlier, separate metadata failures from auth or
            billing failures, and keep the install experience predictable.
          </p>
        </Card>
        <Card className="border-2 border-foreground p-6">
          <p className="text-sm leading-7 text-muted-foreground">
            Continue with the <Link className="font-semibold text-foreground underline" href="/guides/mcp-dcr-explained">DCR explainer</Link>{' '}
            or the <Link className="font-semibold text-foreground underline" href="/guides/mcp-server-installation">install guide</Link>.
          </p>
        </Card>
      </div>
    </main>
  )
}
