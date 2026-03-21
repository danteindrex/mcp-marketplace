import type { Metadata } from 'next'
import Link from 'next/link'
import { JsonLd } from '@/components/json-ld'
import { Text } from '@/components/retroui/text'
import { Badge } from '@/components/ui/badge'
import { Card } from '@/components/ui/card'
import { createBreadcrumbJsonLd, toAbsoluteUrl } from '@/lib/seo'

const title = 'How To Install an MCP Server'
const description =
  'Learn how MCP Marketplace makes MCP server installation feel seamless across Codex, Claude Desktop, Cursor, VS Code, and ChatGPT.'

export const metadata: Metadata = {
  title,
  description,
  alternates: {
    canonical: '/guides/mcp-server-installation',
  },
  openGraph: {
    title,
    description,
    url: '/guides/mcp-server-installation',
    type: 'article',
  },
}

const faqs = [
  {
    question: 'How does MCP Marketplace make installation feel seamless?',
    answer:
      'The marketplace breaks install into five clear states: choose a client, verify metadata, approve scopes, settle payment if needed, and launch the correct install action for that client.',
  },
  {
    question: 'Do buyers need to understand CIMD or DCR to install a server?',
    answer:
      'No. Buyers see a guided install flow while the marketplace handles metadata checks, OAuth discovery, scope prompts, and client-specific launch actions in the background.',
  },
  {
    question: 'What happens if payment or metadata blocks the install?',
    answer:
      'The install flow surfaces the blocking step separately so buyers know whether they need to fund a wallet, settle an x402 challenge, or re-check metadata before retrying.',
  },
]

export default function MCPServerInstallationGuidePage() {
  const guideUrl = toAbsoluteUrl('/guides/mcp-server-installation')
  const breadcrumbData = createBreadcrumbJsonLd([
    { name: 'Home', path: '/' },
    { name: 'Guides', path: '/guides' },
    { name: title, path: '/guides/mcp-server-installation' },
  ])
  const faqData = {
    '@context': 'https://schema.org',
    '@type': 'FAQPage',
    mainEntity: faqs.map(item => ({
      '@type': 'Question',
      name: item.question,
      acceptedAnswer: {
        '@type': 'Answer',
        text: item.answer,
      },
    })),
  }
  const howToData = {
    '@context': 'https://schema.org',
    '@type': 'HowTo',
    name: title,
    description,
    step: [
      {
        '@type': 'HowToStep',
        name: 'Choose a client',
        text: 'Pick Codex, Claude Desktop, Cursor, VS Code, or ChatGPT so the marketplace can prepare the right install action.',
      },
      {
        '@type': 'HowToStep',
        name: 'Verify install metadata',
        text: 'The marketplace checks CIMD, OAuth metadata, JWKS, and buyer hub readiness before asking the user to continue.',
      },
      {
        '@type': 'HowToStep',
        name: 'Approve scopes and entitlement',
        text: 'The buyer reviews required scopes and sees whether the server is free, auto-entitled, or payment-gated.',
      },
      {
        '@type': 'HowToStep',
        name: 'Complete payment only when needed',
        text: 'Paid installs separate wallet funding or x402 settlement from the rest of setup so the user understands the exact blocker.',
      },
      {
        '@type': 'HowToStep',
        name: 'Launch the client-specific action',
        text: 'The marketplace creates a connection and opens the correct launch URL or local bridge flow for the selected client.',
      },
    ],
    url: guideUrl,
  }

  return (
    <main className="min-h-screen bg-background px-4 py-16 sm:px-6 lg:px-8">
      <JsonLd data={breadcrumbData} />
      <JsonLd data={faqData} />
      <JsonLd data={howToData} />
      <div className="mx-auto flex max-w-4xl flex-col gap-8">
        <div className="space-y-4">
          <Badge variant="outline">Install Guide</Badge>
          <Text variant="h1" className="uppercase sm:text-5xl">{title}</Text>
          <Text variant="h6" className="max-w-3xl text-muted-foreground">
            MCP Marketplace makes installation feel seamless by turning complex protocol work into
            a guided buyer journey with one decision at a time.
          </Text>
        </div>

        <Card className="border-2 border-foreground p-6">
          <Text variant="h4" className="uppercase">Answer Capsule</Text>
          <Text variant="small" className="mt-3 leading-7 text-muted-foreground">
            MCP Marketplace makes MCP server installation seamless by checking metadata, separating
            payment from permissions, and generating a client-specific one-click action after the
            buyer selects Codex, Claude Desktop, Cursor, VS Code, or ChatGPT.
          </Text>
        </Card>

        <section className="space-y-4">
          <Text variant="h4" className="uppercase">Why The Flow Feels Seamless</Text>
          <Text variant="small" className="leading-7 text-muted-foreground">
            Buyers should not have to reason about discovery documents, token endpoints, or payment
            headers before they can connect a tool. The marketplace handles those checks first and
            only asks the buyer for the next decision that matters.
          </Text>
          <Text variant="small" className="leading-7 text-muted-foreground">
            This removes the usual failure pattern where a user clicks install, gets a protocol
            error, and has no idea whether the issue is authentication, payment, or client support.
          </Text>
        </section>

        <section className="grid gap-4 md:grid-cols-2">
          <Card className="border-2 border-foreground p-6">
            <Text variant="h5" className="uppercase">The 5-Step Install Model</Text>
            <ol className="mt-4 space-y-3">
              <Text as="li" variant="small" className="text-muted-foreground">1. Choose the client so the marketplace can tailor the launch action.</Text>
              <Text as="li" variant="small" className="text-muted-foreground">2. Verify install metadata before asking for trust or consent.</Text>
              <Text as="li" variant="small" className="text-muted-foreground">3. Review scopes so the buyer sees exactly what the server needs.</Text>
              <Text as="li" variant="small" className="text-muted-foreground">4. Settle payment only when the server is not free.</Text>
              <Text as="li" variant="small" className="text-muted-foreground">5. Launch a one-click action or local bridge flow to finish setup.</Text>
            </ol>
          </Card>
          <Card className="border-2 border-foreground p-6">
            <Text variant="h5" className="uppercase">Recovery By Failure Type</Text>
            <ul className="mt-4 space-y-3">
              <Text as="li" variant="small" className="text-muted-foreground">Metadata block: re-check CIMD, OAuth metadata, and JWKS readiness.</Text>
              <Text as="li" variant="small" className="text-muted-foreground">Scope block: show the missing scopes before continuing.</Text>
              <Text as="li" variant="small" className="text-muted-foreground">Payment block: fund a wallet or settle the x402 challenge, then retry.</Text>
              <Text as="li" variant="small" className="text-muted-foreground">Client block: switch to a supported client without losing install context.</Text>
            </ul>
          </Card>
        </section>

        <section className="space-y-4">
          <Text variant="h4" className="uppercase">Canonical Questions</Text>
          <div className="space-y-4">
            {faqs.map(item => (
              <Card key={item.question} className="border-2 border-foreground p-6">
                <Text variant="h6">{item.question}</Text>
                <Text variant="small" className="mt-2 leading-7 text-muted-foreground">{item.answer}</Text>
              </Card>
            ))}
          </div>
        </section>

        <Card className="border-2 border-foreground p-6">
          <Text variant="h4" className="uppercase">Next Step</Text>
          <Text variant="small" className="mt-3 leading-7 text-muted-foreground">
            Browse live listings in the <Link className="font-semibold text-foreground underline" href="/marketplace">marketplace</Link>{' '}
            or open a specific server page to see pricing, required scopes, install support, and
            the guided setup path.
          </Text>
        </Card>
      </div>
    </main>
  )
}
