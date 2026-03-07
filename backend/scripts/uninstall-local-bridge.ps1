param(
  [string]$BinaryPath = ""
)

$ErrorActionPreference = "Stop"

$scriptRoot = Split-Path -Parent $MyInvocation.MyCommand.Path
$backendRoot = (Resolve-Path (Join-Path $scriptRoot "..")).Path

if (-not $BinaryPath) {
  $BinaryPath = Join-Path $backendRoot "bin\\mcp-local-bridge.exe"
}

if (Test-Path $BinaryPath) {
  & $BinaryPath uninstall
  if ($LASTEXITCODE -ne 0) {
    throw "local bridge uninstall failed"
  }
} else {
  Write-Host "Binary not found at $BinaryPath"
  Write-Host "Removing protocol handler directly..."
  reg delete HKCU\\Software\\Classes\\mcp-marketplace /f | Out-Null
}

Write-Host "MCP Local Bridge uninstalled."
