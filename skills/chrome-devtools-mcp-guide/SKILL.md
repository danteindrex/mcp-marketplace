---
name: chrome-devtools-mcp-guide
description: Configure, connect, and operate the official Chrome DevTools MCP server from Codex. Use when Codex needs to install or update chrome-devtools-mcp, choose between launching Chrome or attaching to an existing session, troubleshoot connection failures, or drive browser inspection tasks such as navigation, DOM snapshots, script execution, network inspection, screenshots, and performance traces.
---

# Chrome DevTools MCP Guide

## Overview

Use this skill to work with the official `chrome-devtools-mcp` server in Codex on Windows or other local environments. Prefer it when the task is browser automation or debugging through the live Chrome session rather than generic web search.

## Workflow

1. Confirm the server mode before taking actions.
2. Attach to the correct Chrome session.
3. Verify connectivity with a minimal live check.
4. Perform the requested browser task with the narrowest tool set needed.
5. Report concrete results and any environment limits.

## Confirm Server Mode

- Check whether Codex is configured to launch Chrome or reuse an existing instance.
- Prefer `--autoConnect` when the user wants to reuse an already running local Chrome session and Chrome 144+ is installed.
- Prefer `--browserUrl http://127.0.0.1:9222` when the user has a dedicated debug-port Chrome instance or `--autoConnect` is unreliable.
- Keep the Windows Codex config in the form:

```toml
[mcp_servers.chrome-devtools]
command = "cmd"
args = ["/c", "npx", "-y", "chrome-devtools-mcp@latest", "--autoConnect"]
env = { SystemRoot="C:\\Windows", PROGRAMFILES="C:\\Program Files" }
startup_timeout_ms = 20000
```

- Read [connection-modes.md](references/connection-modes.md) when deciding which flags to use or when troubleshooting attachment failures.

## Attach And Verify

1. List pages first to confirm the server can attach.
2. If attach fails, identify whether Chrome is not running, remote debugging is not enabled, or the server is targeting the wrong connection mode.
3. Once attached, run a minimal verification:
   - list pages
   - take a snapshot on the selected page
   - evaluate a small script that returns `document.title` and `location.href`
4. Treat successful page listing plus one live script evaluation as the baseline proof that the MCP server works.

## Use The Tools

- Use `list_pages` to see the live session and detect whether Codex attached to the expected browser.
- Use `select_page` before interacting with a specific tab.
- Use `navigate_page` only when changing location is part of the requested task.
- Prefer `take_snapshot` over screenshots for most inspection tasks because it is lighter and exposes structured page content.
- Use `evaluate_script` for DOM state, local storage, JS variables, and targeted extraction.
- Use `list_network_requests` and `get_network_request` for API debugging.
- Use `take_screenshot` only when visual evidence is necessary.
- Use performance tools only when the user asks for performance or Core Web Vitals analysis.

## Working Rules

- Assume the attached Chrome session may contain sensitive tabs and accounts.
- Do not navigate away from the selected tab unless the user asked for it or a verification step requires it.
- Prefer opening a new tab for isolated testing when the current tab matters.
- State clearly whether the server attached to an existing Chrome session or launched its own instance.
- When a connection fails, report the exact symptom and the next corrective action instead of retrying blindly.

## Common Requests

- "Install the official Chrome MCP server in Codex."
- "Make Codex use my already running Chrome."
- "Test whether Chrome DevTools MCP is working."
- "Inspect this page and extract data from the DOM."
- "Check failing network requests in my current browser tab."
- "Run a performance trace on this page."
