import { NextResponse } from 'next/server'

export const runtime = 'nodejs'

function buildBatchScript(origin: string) {
  return `@echo off
setlocal
set "BRIDGE_DIR=%LocalAppData%\\MCPMarketplace"
set "BRIDGE_EXE=%BRIDGE_DIR%\\mcp-local-bridge.exe"

if not exist "%BRIDGE_DIR%" mkdir "%BRIDGE_DIR%"

echo Downloading MCP Local Bridge...
powershell -NoProfile -ExecutionPolicy Bypass -Command "$ProgressPreference = 'SilentlyContinue'; Invoke-WebRequest -UseBasicParsing -Uri '${origin}/api/local-bridge/windows-exe' -OutFile '%BRIDGE_EXE%'"
if errorlevel 1 (
  echo Failed to download MCP Local Bridge.
  pause
  exit /b 1
)

echo Registering MCP Local Bridge...
"%BRIDGE_EXE%" install
if errorlevel 1 (
  echo Failed to register MCP Local Bridge.
  pause
  exit /b 1
)

echo.
echo MCP Local Bridge installed.
echo Return to MCP Marketplace and click Retry One-Click Install.
pause
`
}

export async function GET(request: Request) {
  const url = new URL(request.url)
  const body = buildBatchScript(url.origin)

  return new NextResponse(body, {
    status: 200,
    headers: {
      'Content-Type': 'application/octet-stream; charset=utf-8',
      'Content-Disposition': 'attachment; filename="install-mcp-local-bridge.cmd"',
      'Cache-Control': 'no-store',
    },
  })
}
