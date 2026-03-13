import type { Metadata } from 'next'
import Link from 'next/link'
import { Badge } from '@/components/ui/badge'
import { Card } from '@/components/ui/card'
import { Text } from '@/components/retroui'

export const metadata: Metadata = {
  title: 'MCP Guides',
  description: 'Explore MCP Marketplace guides for installation, compatibility, auth, and pricing.',
  alternates: {
    canonical: '/guides',
  },
}

const guideGroups = [
  {
    title: 'Install Guides',
    links: [
      { href: '/guides/mcp-server-installation', label: 'General MCP Server Installation' },
      { href: '/guides/install-mcp-in-codex', label: 'Install MCP in Codex' },
      { href: '/guides/install-mcp-in-claude-desktop', label: 'Install MCP in Claude Desktop' },
      { href: '/guides/install-mcp-in-cursor', label: 'Install MCP in Cursor' },
      { href: '/guides/install-mcp-in-vs-code', label: 'Install MCP in VS Code' },
      { href: '/guides/install-mcp-in-chatgpt', label: 'Install MCP in ChatGPT' },
    ],
  },
  {
    title: 'Explainers',
    links: [
      { href: '/guides/what-is-mcp', label: 'What Is MCP?' },
      { href: '/guides/mcp-cimd-explained', label: 'CIMD Explained' },
      { href: '/guides/mcp-dcr-explained', label: 'DCR Explained' },
      { href: '/guides/mcp-oauth-explained', label: 'OAuth Explained' },
      { href: '/guides/x402-pricing-explained', label: 'x402 Pricing Explained' },
      { href: '/guides/hosted-vs-local-mcp-servers', label: 'Hosted vs Local MCP Servers' },
    ],
  },
]

export default function GuidesIndexPage() {
  return (
    <main className="min-h-screen bg-background px-4 py-16 sm:px-6 lg:px-8">
      <div className="mx-auto flex max-w-5xl flex-col gap-8">
        <div className="space-y-4">
          <Badge variant="outline">Guide Library</Badge>
          <Text variant="h1" className="uppercase sm:text-5xl">MCP Guides</Text>
          <Text variant="h6" className="max-w-3xl text-muted-foreground">
            Canonical MCP Marketplace guides for installation, compatibility, auth, pricing, and
            client-specific setup paths.
          </Text>
        </div>

        {guideGroups.map(group => (
          <section key={group.title} className="space-y-4">
            <Text variant="h4" className="uppercase">{group.title}</Text>
            <div className="grid gap-4 md:grid-cols-2">
              {group.links.map(link => (
                <Link key={link.href} href={link.href}>
                  <Card className="border-2 border-foreground p-6 transition-all hover:translate-x-[-2px] hover:translate-y-[-2px] hover:shadow-[6px_6px_0px_hsl(var(--shadow-color))]">
                    <Text variant="h6">{link.label}</Text>
                  </Card>
                </Link>
              ))}
            </div>
          </section>
        ))}
      </div>
    </main>
  )
}
