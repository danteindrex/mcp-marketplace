# Chrome DevTools MCP Connection Modes

Use this reference when choosing how Codex should connect to Chrome.

## Preferred Modes

### Reuse An Existing Local Chrome

Use `--autoConnect`.

- Best when the user wants Codex to attach to their current Chrome session.
- Requires Chrome 144+.
- The running Chrome instance must have remote debugging enabled from `chrome://inspect/#remote-debugging`.
- Typical Codex config:

```toml
[mcp_servers.chrome-devtools]
command = "cmd"
args = ["/c", "npx", "-y", "chrome-devtools-mcp@latest", "--autoConnect"]
env = { SystemRoot="C:\\Windows", PROGRAMFILES="C:\\Program Files" }
startup_timeout_ms = 20000
```

### Connect To A Dedicated Debug Instance

Use `--browserUrl http://127.0.0.1:9222`.

- Best when the user starts Chrome explicitly with a debug port.
- More explicit than `--autoConnect`.
- Useful when multiple Chrome profiles or sessions are open.

Example:

```toml
[mcp_servers.chrome-devtools]
command = "cmd"
args = ["/c", "npx", "-y", "chrome-devtools-mcp@latest", "--browserUrl", "http://127.0.0.1:9222"]
env = { SystemRoot="C:\\Windows", PROGRAMFILES="C:\\Program Files" }
startup_timeout_ms = 20000
```

## Quick Verification Flow

1. `list_pages`
2. `take_snapshot`
3. `evaluate_script` with:

```js
() => ({ title: document.title, url: location.href, readyState: document.readyState })
```

If all three work, the server is attached and usable.

## Common Failures

### "Could not find DevToolsActivePort"

Meaning:
- Chrome is running but the current session is not exposing remote debugging for `--autoConnect`.

Fix:
- Open `chrome://inspect/#remote-debugging`
- Enable remote debugging
- Accept any Chrome permission prompt
- Retry `list_pages`

### No pages or wrong pages

Meaning:
- Codex attached to a different Chrome instance or profile than expected.

Fix:
- Use `select_page` if the expected tab is present.
- If not, switch to `--browserUrl` and target the intended debug instance directly.

### MCP config parses but attach still fails

Meaning:
- The server command is installed, but the browser-side requirements are not satisfied.

Fix:
- Verify the active flags in `codex mcp get chrome-devtools`
- Verify Chrome version
- Verify remote debugging is enabled or the debug port is reachable
