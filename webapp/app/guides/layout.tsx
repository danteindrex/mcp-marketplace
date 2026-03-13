import type { Metadata } from 'next'

export const metadata: Metadata = {
  title: 'Guides',
  description: 'Canonical MCP Marketplace guides for MCP installation, compatibility, auth, and pricing.',
}

export default function GuidesLayout({ children }: { children: React.ReactNode }) {
  return children
}
