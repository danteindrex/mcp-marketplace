# Buyer Install Journeys Skill

Use this skill when working on discovery, one-click install, client onboarding, hub flows, and post-install recovery UX.

## Why This Skill Exists

Research notes:
- the requirements center on one-click install using CIMD and DCR concepts for clients like Claude Desktop and VS Code
- MCP supports a wide ecosystem of hosts, so compatibility and guided UX matter more than client-specific shortcuts
- MCP remote flows depend on capability negotiation, auth metadata, and transport assumptions being understandable to the buyer flow

## What This Repo Needs

- a real install journey from listing -> auth/payment -> entitlement -> connection
- clear explanation of local versus cloud support
- install recovery when auth, scope, or payment blocks progress
- buyer hub and connection management that feels coherent across clients

## Decision Rules

- optimize for reducing dead-end install failures
- always show the next action when install is blocked
- treat scope prompts, payment prompts, and local setup as separate user states
- keep the buyer's mental model simple even if backend flows are complex

## Implementation Focus

- marketplace listing to install CTA
- DCR/CIMD-inspired setup and redirect flows
- hub endpoint projection and connection visibility
- local-agent enrollment and one-click local install bridge
- billing, entitlement, and scope-challenge recovery UX

## Done Looks Like

- a buyer can understand what they are buying and where it runs
- install succeeds or fails with a clear recovery path
- hub, connections, and billing state line up with the install outcome
