# Docker Marketplace Ops Skill

Use this skill when handling Docker Hub import, image validation, container deployment assumptions, or self-hosting product flows.

## Why This Skill Exists

Research notes:
- the requirements make Docker Hub import and hosted/self-hosted images a primary merchant workflow
- container packaging quality determines whether a server can be deployed remotely or offered for local self-hosting

## What This Repo Needs

- merchant import of existing Docker images
- validation that imported images can support the marketplace deployment contract
- a product distinction between hosted cloud execution and local self-hosted licensing

## Decision Rules

- treat Docker image ingestion as a trust and operability boundary
- validate image metadata and runtime expectations before publishability
- separate hosted readiness from self-hosting eligibility
- do not publish images that cannot meet discovery, auth, and deployment requirements

## Implementation Focus

- Docker Hub import UX and backend validation
- image metadata capture
- deployment compatibility checks
- self-hosting surcharge and local install product rules
- server readiness for required well-known auth/resource endpoints

## Done Looks Like

- merchants understand whether an image is importable, deployable, and publishable
- hosted and local offerings are clearly distinguished
- image validation failures are specific and actionable
