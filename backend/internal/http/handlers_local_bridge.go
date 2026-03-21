package http

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const localBridgeBinDir = "/app/bin"

func (a *App) downloadLocalBridgeWindowsBinary(w http.ResponseWriter, r *http.Request) {
	a.serveLocalBridgeBinary(w, filepath.Join(localBridgeBinDir, "mcp-local-bridge.exe"), "mcp-local-bridge.exe")
}

func (a *App) downloadLocalBridgeMacOSBinary(w http.ResponseWriter, r *http.Request) {
	a.serveLocalBridgeBinary(w, filepath.Join(localBridgeBinDir, "mcp-local-bridge-macos"), "mcp-local-bridge-macos")
}

func (a *App) downloadLocalBridgeWindowsInstaller(w http.ResponseWriter, r *http.Request) {
	downloadURL := localBridgeBaseURL(a.cfg.BaseURL, r) + "/v1/local-bridge/windows-exe"
	body := `@echo off
setlocal
set "BRIDGE_DIR=%LocalAppData%\MCPMarketplace"
set "BRIDGE_EXE=%BRIDGE_DIR%\mcp-local-bridge.exe"

if not exist "%BRIDGE_DIR%" mkdir "%BRIDGE_DIR%"

echo Downloading MCP Local Bridge...
powershell -NoProfile -ExecutionPolicy Bypass -Command "$ProgressPreference = 'SilentlyContinue'; Invoke-WebRequest -UseBasicParsing -Uri '` + downloadURL + `' -OutFile '%BRIDGE_EXE%'"
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
	w.Header().Set("Content-Type", "application/octet-stream; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="install-mcp-local-bridge.cmd"`)
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(body))
}

func (a *App) downloadLocalBridgeMacOSInstaller(w http.ResponseWriter, r *http.Request) {
	downloadURL := localBridgeBaseURL(a.cfg.BaseURL, r) + "/v1/local-bridge/macos-bin"
	body := `#!/usr/bin/env bash
set -euo pipefail

BRIDGE_DIR="$HOME/.local/share/mcp-marketplace"
BRIDGE_BIN="$BRIDGE_DIR/mcp-local-bridge"

mkdir -p "$BRIDGE_DIR"

echo "Downloading MCP Local Bridge for macOS..."
curl -fsSL "` + downloadURL + `" -o "$BRIDGE_BIN"
chmod +x "$BRIDGE_BIN"

echo "Registering MCP Local Bridge..."
"$BRIDGE_BIN" install

echo
echo "MCP Local Bridge installed."
echo "Return to MCP Marketplace and click Retry One-Click Install."
`
	w.Header().Set("Content-Type", "application/octet-stream; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="install-mcp-local-bridge-macos.command"`)
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(body))
}

func (a *App) serveLocalBridgeBinary(w http.ResponseWriter, path string, filename string) {
	binary, err := os.ReadFile(path)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "local bridge binary unavailable"})
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(binary)
}

func localBridgeBaseURL(configuredBaseURL string, r *http.Request) string {
	if baseURL := strings.TrimRight(strings.TrimSpace(configuredBaseURL), "/"); baseURL != "" {
		return baseURL
	}
	return strings.TrimRight(aBaseURL(r), "/")
}
