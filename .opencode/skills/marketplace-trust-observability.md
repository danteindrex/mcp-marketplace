# Marketplace Trust And Observability Skill

Use this skill when working on listing verification, ratings, audits, telemetry, operational dashboards, or marketplace governance.

## Why This Skill Exists

Research notes:
- the requirements call for a verified registry, ratings, real-time audit logs, telemetry, error-rate visibility, and latency monitoring
- n8n and MCP ecosystems both highlight dynamic operations and runtime visibility as important for production reliability

## What This Repo Needs

- trust signals for buyers choosing servers
- observability for merchants and admins managing deployments and payment flows
- auditability for security and billing-sensitive operations

## Decision Rules

- trust claims should be tied to real verification or operational facts
- audit logs should focus on meaningful mutations and security-relevant actions
- observability should connect errors to user impact, not just internal metrics
- merchant and admin telemetry should help diagnose deploy, auth, and payment issues quickly

## Implementation Focus

- verified listing indicators and future rating support
- audit logs for auth, deploy, publish, settlement, and entitlement events
- latency/error dashboards and API health visibility
- buyer-safe trust surfaces and merchant/admin ops visibility

## Done Looks Like

- buyers have better signals about what is safe and reliable
- merchants can diagnose deployment and monetization issues faster
- admins can review platform risk, payment readiness, and security events
