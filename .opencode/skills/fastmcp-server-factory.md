# FastMCP Server Factory Skill

Use this skill when designing Python-based MCP server authoring, templates, generated scaffolds, or hosted server standards.

## Why This Skill Exists

Research notes:
- FastMCP positions itself as a standard framework for building MCP servers, clients, and interactive apps.
- It emphasizes automatic schema generation, transport handling, and production-friendly MCP patterns.
- It explicitly treats servers, clients, and interactive apps as separate pillars.

## What This Repo Needs

- a clean path for sellers to create or onboard MCP servers rapidly
- opinionated templates for common tool, resource, and prompt patterns
- hosted-server assumptions that still preserve MCP correctness

## Decision Rules

- optimize for quick authoring without hiding protocol responsibilities
- prefer generated scaffolds that are easy to inspect and extend
- keep tool schemas, descriptions, and prompts high quality because marketplace discovery depends on them
- if interactive app support is added, keep it optional and capability-driven

## Implementation Focus

- Python/FastMCP starter templates
- seller onboarding flows for server creation
- packaging and deployment expectations for hosted MCP servers
- schema quality, prompt quality, and server metadata quality

## Done Looks Like

- sellers can go from idea to valid MCP server quickly
- generated servers are standards-aligned and deployable
- server metadata is good enough for marketplace listing and client use
