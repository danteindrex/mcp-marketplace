---
name: mcp-requirements-auditor
description: Build a requirements-to-implementation audit for software projects. Use when a repository contains a requirements document (for example requirements.md, PRD, checklist, compliance spec) and you need a severity-ranked completeness report with concrete file evidence and missing-item gaps.
---

# Mcp Requirements Auditor

## Workflow

1. Read the requirements source and normalize it into atomic requirements.
2. Group requirements into domains (backend, frontend, infra, security, monetization, testing).
3. Spawn explorer sub-agents per domain for parallel evidence collection.
4. Require each finding to include:
   - status (`implemented`, `partial`, `missing`)
   - severity (`critical`, `high`, `medium`, `low`)
   - at least one code/doc file reference
5. Merge duplicate findings and keep the strongest evidence.
6. Produce a final gap report ordered by severity.

## Audit Rules

- Do not claim implementation without file evidence.
- Mark as `partial` when UI exists but backend behavior is missing (or vice versa).
- Mark as `missing` when no meaningful implementation exists.
- Prefer primary implementation files over planning docs as evidence.
- Call out testing gaps separately from feature gaps.

## Output Format

- Section 1: `Critical/High gaps`
- Section 2: `Medium/Low gaps`
- Section 3: `Coverage summary by requirement domain`
- Section 4: `Top next actions`

Use the traceability template in [requirements-traceability-template.md](references/requirements-traceability-template.md).
