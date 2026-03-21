import { NextResponse } from 'next/server'

export const runtime = 'nodejs'

function buildMacScript(origin: string) {
  return `#!/usr/bin/env bash
set -euo pipefail

BRIDGE_DIR="$HOME/.local/share/mcp-marketplace"
BRIDGE_BIN="$BRIDGE_DIR/mcp-local-bridge"

mkdir -p "$BRIDGE_DIR"

echo "Downloading MCP Local Bridge for macOS..."
curl -fsSL "${origin}/api/local-bridge/macos-bin" -o "$BRIDGE_BIN"
chmod +x "$BRIDGE_BIN"

echo "Registering MCP Local Bridge..."
"$BRIDGE_BIN" install

echo
echo "MCP Local Bridge installed."
echo "Return to MCP Marketplace and click Retry One-Click Install."
`
}

export async function GET(request: Request) {
  const url = new URL(request.url)
  const body = buildMacScript(url.origin)

  return new NextResponse(body, {
    status: 200,
    headers: {
      'Content-Type': 'application/octet-stream; charset=utf-8',
      'Content-Disposition': 'attachment; filename="install-mcp-local-bridge-macos.command"',
      'Cache-Control': 'no-store',
    },
  })
}
