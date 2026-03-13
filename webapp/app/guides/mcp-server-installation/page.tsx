import type { Metadata } from 'next'
import Link from 'next/link'
import { JsonLd } from '@/components/json-ld'
import { Badge } from '@/components/ui/badge'
import { Card } from '@/components/ui/card'
import { getSiteUrl } from '@/lib/site'

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
  const siteUrl = getSiteUrl()
  const guideUrl = `${siteUrl}/guides/mcp-server-installation`
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
      <JsonLd data={faqData} />
      <JsonLd data={howToData} />
      <div className="mx-auto flex max-w-4xl flex-col gap-8">
        <div className="space-y-4">
          <Badge variant="outline">Install Guide</Badge>
          <h1 className="text-4xl font-black uppercase sm:text-5xl">{title}</h1>
          <p className="max-w-3xl text-lg text-muted-foreground">
            MCP Marketplace makes installation feel seamless by turning complex protocol work into
            a guided buyer journey with one decision at a time.
          </p>
        </div>

        <Card className="border-2 border-foreground p-6">
          <h2 className="text-2xl font-black uppercase">Answer Capsule</h2>
          <p className="mt-3 text-sm leading-7 text-muted-foreground">
            MCP Marketplace makes MCP server installation seamless by checking metadata, separating
            payment from permissions, and generating a client-specific one-click action after the
            buyer selects Codex, Claude Desktop, Cursor, VS Code, or ChatGPT.
          </p>
        </Card>

        <section className="space-y-4">
          <h2 className="text-2xl font-black uppercase">Why The Flow Feels Seamless</h2>
          <p className="text-sm leading-7 text-muted-foreground">
            Buyers should not have to reason about discovery documents, token endpoints, or payment
            headers before they can connect a tool. The marketplace handles those checks first and
            only asks the buyer for the next decision that matters.
          </p>
          <p className="text-sm leading-7 text-muted-foreground">
            This removes the usual failure pattern where a user clicks install, gets a protocol
            error, and has no idea whether the issue is authentication, payment, or client support.
          </p>
        </section>

        <section className="grid gap-4 md:grid-cols-2">
          <Card className="border-2 border-foreground p-6">
            <h2 className="text-xl font-black uppercase">The 5-Step Install Model</h2>
            <ol className="mt-4 space-y-3 text-sm text-muted-foreground">
              <li>1. Choose the client so the marketplace can tailor the launch action.</li>
              <li>2. Verify install metadata before asking for trust or consent.</li>
              <li>3. Review scopes so the buyer sees exactly what the server needs.</li>
              <li>4. Settle payment only when the server is not free.</li>
              <li>5. Launch a one-click action or local bridge flow to finish setup.</li>
            </ol>
          </Card>
          <Card className="border-2 border-foreground p-6">
            <h2 className="text-xl font-black uppercase">Recovery By Failure Type</h2>
            <ul className="mt-4 space-y-3 text-sm text-muted-foreground">
              <li>Metadata block: re-check CIMD, OAuth metadata, and JWKS readiness.</li>
              <li>Scope block: show the missing scopes before continuing.</li>
              <li>Payment block: fund a wallet or settle the x402 challenge, then retry.</li>
              <li>Client block: switch to a supported client without losing install context.</li>
            </ul>
          </Card>
        </section>

        <section className="space-y-4">
          <h2 className="text-2xl font-black uppercase">Canonical Questions</h2>
          <div className="space-y-4">
            {faqs.map(item => (
              <Card key={item.question} className="border-2 border-foreground p-6">
                <h3 className="text-lg font-black">{item.question}</h3>
                <p className="mt-2 text-sm leading-7 text-muted-foreground">{item.answer}</p>
              </Card>
            ))}
          </div>
        </section>

        <Card className="border-2 border-foreground p-6">
          <h2 className="text-2xl font-black uppercase">Next Step</h2>
          <p className="mt-3 text-sm leading-7 text-muted-foreground">
            Browse live listings in the <Link className="font-semibold text-foreground underline" href="/marketplace">marketplace</Link>{' '}
            or open a specific server page to see pricing, required scopes, install support, and
            the guided setup path.
          </p>
        </Card>
      </div>
    </main>
  )
}
