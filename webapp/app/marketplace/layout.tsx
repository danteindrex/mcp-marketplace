import type { Metadata } from 'next'
import { JsonLd } from '@/components/json-ld'
import { getSiteUrl } from '@/lib/site'

export const metadata: Metadata = {
  title: 'Marketplace',
  description: 'Discover and install verified MCP servers across Codex, Claude, Cursor, VS Code, and ChatGPT.',
  alternates: {
    canonical: '/marketplace',
  },
  openGraph: {
    title: 'MCP Marketplace',
    description: 'Discover and install verified MCP servers across leading AI clients.',
    url: '/marketplace',
    type: 'website',
  },
}

export default function MarketplaceLayout({ children }: { children: React.ReactNode }) {
  const siteUrl = getSiteUrl()
  const jsonLd = {
    '@context': 'https://schema.org',
    '@type': 'CollectionPage',
    name: 'MCP Marketplace',
    description:
      'Discover and install verified MCP servers across Codex, Claude, Cursor, VS Code, and ChatGPT.',
    url: `${siteUrl}/marketplace`,
  }

  return (
    <>
      <JsonLd data={jsonLd} />
      {children}
    </>
  )
}
