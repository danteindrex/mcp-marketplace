import type { Metadata } from 'next'
import Link from 'next/link'
import { JsonLd } from '@/components/json-ld'
import { Badge } from '@/components/ui/badge'
import { Card } from '@/components/ui/card'
import { getSiteUrl } from '@/lib/site'

const title = 'How To Install MCP In ChatGPT'
const description =
  'Learn how MCP Marketplace prepares a ChatGPT-ready remote MCP setup flow with metadata checks, permissions review, and guided connection steps.'

export const metadata: Metadata = {
  title,
  description,
  alternates: { canonical: '/guides/install-mcp-in-chatgpt' },
  openGraph: { title, description, url: '/guides/install-mcp-in-chatgpt', type: 'article' },
}

export default function InstallMCPInChatGPTGuidePage() {
  const jsonLd = {
    '@context': 'https://schema.org',
    '@type': 'HowTo',
    name: title,
    description,
    url: `${getSiteUrl()}/guides/install-mcp-in-chatgpt`,
    step: [
      { '@type': 'HowToStep', name: 'Open a server page', text: 'Choose the MCP server to connect through the marketplace.' },
      { '@type': 'HowToStep', name: 'Select ChatGPT', text: 'Choose ChatGPT as the target client in the install flow.' },
      { '@type': 'HowToStep', name: 'Validate metadata and scopes', text: 'The platform confirms discovery, auth readiness, and permission scope requirements.' },
      { '@type': 'HowToStep', name: 'Complete payment only when needed', text: 'Paid listings handle billing before the connection completes.' },
      { '@type': 'HowToStep', name: 'Finish the remote connection', text: 'Use the generated ChatGPT setup action to complete the install.' },
    ],
  }

  return (
    <main className="min-h-screen bg-background px-4 py-16 sm:px-6 lg:px-8">
      <JsonLd data={jsonLd} />
      <div className="mx-auto flex max-w-4xl flex-col gap-8">
        <div className="space-y-4">
          <Badge variant="outline">ChatGPT Guide</Badge>
          <h1 className="text-4xl font-black uppercase sm:text-5xl">{title}</h1>
          <p className="max-w-3xl text-lg text-muted-foreground">
            The ChatGPT path works best when metadata, auth, and billing are handled before the
            final remote connection step.
          </p>
        </div>
        <Card className="border-2 border-foreground p-6">
          <h2 className="text-2xl font-black uppercase">Answer Capsule</h2>
          <p className="mt-3 text-sm leading-7 text-muted-foreground">
            To install MCP in ChatGPT through MCP Marketplace, open a listing, choose ChatGPT,
            verify readiness and permissions, settle payment if required, and finish the generated
            remote connection flow.
          </p>
        </Card>
        <Card className="border-2 border-foreground p-6">
          <h2 className="text-2xl font-black uppercase">Next Step</h2>
          <p className="mt-3 text-sm leading-7 text-muted-foreground">
            Start from the <Link className="font-semibold text-foreground underline" href="/marketplace">marketplace</Link>{' '}
            or review the <Link className="font-semibold text-foreground underline" href="/guides/what-is-mcp">MCP explainer</Link>.
          </p>
        </Card>
      </div>
    </main>
  )
}
