# MCP Protocol Integration Skill

Use this skill when working on MCP protocol behavior, client compatibility, transports, or remote/local server lifecycle.

## Why This Skill Exists

Research notes:
- MCP defines hosts, clients, and servers as distinct roles and uses JSON-RPC 2.0 for lifecycle, tools, resources, prompts, and notifications.
- MCP supports both stdio and streamable HTTP transports.
- Remote MCP servers commonly rely on HTTP transport with authentication, while local servers often use stdio.
- FastMCP positions servers, clients, and interactive apps as first-class MCP building blocks.

## What This Repo Needs

- remote MCP hub behavior that works across clients like Claude, Codex, VS Code, and Cursor
- correct transport assumptions for local versus hosted servers
- support for tools, resources, prompts, and notifications where useful
- safe handling of interactive app or connector capabilities when supported

## Decision Rules

- treat MCP as a protocol product, not just an internal API
- preserve clear distinction between host, client, and server behavior
- prefer standards-aligned JSON-RPC and capability negotiation over ad hoc payloads
- for hosted servers, assume many clients may connect concurrently
- for local-bridge flows, assume stdio-like local execution constraints and explicit device registration

## Implementation Focus

- MCP hub routes and discovery endpoints
- client compatibility assumptions in buyer install flows
- remote server transport, SSE, and connection lifecycle behavior
- tool/resource/prompt exposure decisions
- capability negotiation and notification support

## Done Looks Like

- protocol behavior matches MCP concepts cleanly
- local and remote flows are not conflated
- install and client connection behavior is explainable in MCP terms
- new work improves compatibility instead of relying on one client-specific hack
