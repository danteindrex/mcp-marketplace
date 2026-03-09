---
name: mcp-gap-remediation-planner
description: Convert an audit gap list into a practical implementation backlog for marketplace platforms and API products. Use when you already have implemented/partial/missing findings and need phased delivery, dependency mapping, owners, acceptance criteria, and rollout risk controls.
---

# Mcp Gap Remediation Planner

## Workflow

1. Ingest audit findings and deduplicate by root cause.
2. Score each gap using risk, user impact, compliance/security exposure, and implementation effort.
3. Build phases:
   - Phase 1: critical security/protocol correctness
   - Phase 2: product-completeness gaps
   - Phase 3: quality, telemetry, and ecosystem expansion
4. Define per item:
   - owner role (`backend`, `webapp`, `infra`, `security`, `qa`)
   - dependencies
   - definition of done
   - test coverage required
5. Add rollout guardrails (feature flags, migration plan, fallback behavior).

## Planning Rules

- Prioritize protocol/security defects over UX polish.
- Link every backlog item back to one or more requirements.
- Avoid vague tasks; include deliverable endpoints/pages/tests.
- Include verification commands or test suites per item.

## Output Format

- Section 1: `Prioritized backlog (P0/P1/P2)`
- Section 2: `Dependency graph and sequencing`
- Section 3: `Acceptance criteria and tests`
- Section 4: `Risks and mitigations`

Use [remediation-priority-model.md](references/remediation-priority-model.md) for scoring.
