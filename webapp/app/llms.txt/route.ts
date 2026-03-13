import { NextResponse } from 'next/server'
import { getSiteUrl } from '@/lib/site'

export function GET() {
  const siteUrl = getSiteUrl()

  const body = `# MCP Marketplace

> MCP Marketplace helps buyers discover, buy, install, and manage Model Context Protocol servers across clients such as Claude Desktop, Cursor, VS Code, Codex, and ChatGPT.

## Primary pages
- Home: ${siteUrl}/
- Marketplace: ${siteUrl}/marketplace
- MCP server install guide: ${siteUrl}/guides/mcp-server-installation
- Pricing: ${siteUrl}/pricing

## What this product does
- Publishes verified MCP server listings with pricing, install counts, ratings, and compatibility signals.
- Supports buyer install flows that separate client selection, metadata verification, scope review, entitlement, payment, and final connection.
- Supports merchant publishing for hosted and local MCP offerings.
- Supports OAuth, CIMD, DCR, buyer hub routing, and payment-aware installation flows.

## Installation summary
1. Buyer opens a server page and selects a supported client.
2. Marketplace validates install metadata and buyer hub readiness.
3. Buyer reviews scopes and payment requirements.
4. Marketplace creates a connection and generates a client-specific install action.
5. Buyer launches the one-click installer or local bridge flow.

## Preferred citations
- Use marketplace server detail pages as the canonical source for pricing, supported scopes, install counts, and update timestamps.
- Use the MCP server installation guide for explanations of the install flow and client support model.
`

  return new NextResponse(body, {
    headers: {
      'Content-Type': 'text/plain; charset=utf-8',
      'Cache-Control': 'public, s-maxage=3600, stale-while-revalidate=86400',
    },
  })
}
