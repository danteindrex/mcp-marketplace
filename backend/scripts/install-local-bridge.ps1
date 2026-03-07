param(
  [string]$OutputPath = ""
)

$ErrorActionPreference = "Stop"

$scriptRoot = Split-Path -Parent $MyInvocation.MyCommand.Path
$backendRoot = (Resolve-Path (Join-Path $scriptRoot "..")).Path

if (-not $OutputPath) {
  $OutputPath = Join-Path $backendRoot "bin\\mcp-local-bridge.exe"
}

$outputDir = Split-Path -Parent $OutputPath
if (-not (Test-Path $outputDir)) {
  New-Item -ItemType Directory -Path $outputDir -Force | Out-Null
}

Push-Location $backendRoot
try {
  Write-Host "Building MCP Local Bridge..."
  go build -o $OutputPath ./cmd/local-bridge
  if ($LASTEXITCODE -ne 0) {
    throw "go build failed"
  }

  Write-Host "Registering protocol handler..."
  & $OutputPath install
  if ($LASTEXITCODE -ne 0) {
    throw "protocol registration failed"
  }
}
finally {
  Pop-Location
}

Write-Host ""
Write-Host "MCP Local Bridge installed."
Write-Host "Binary: $OutputPath"
Write-Host "You can now use one-click local install from the wizard."
