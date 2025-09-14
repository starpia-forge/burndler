Goal: From ADR-001..003 + CLAUDE.md, propose the INITIAL repo plan for a mono-repo.
Deliverables:
1) repo/FILES.md: directory layout and purpose per folder (backend/, frontend/, ops/, etc.)
2) backend/openapi/openapi.yaml: minimal initial endpoints for lint/compose/build (+ error envelope).
3) backend/docs/config.md: required env vars (DB, S3 & Local FS modes, JWT).
4) ops/adr/ADR-004-repo-workflow.md: Make targets, dev compose policy, test strategy.
Rules:
- Go+Gin+GORM(Postgres), React+Tailwind.
- Storage interface with S3 (default) and Local FS mode (dev/offline), switchable by env.
- `build:` forbidden anywhere; prefer image@sha256.
- RBAC roles: Developer(RW), Engineer(R).
Acceptance:
- FILES.md fits on one screen and references ADRs.
- OpenAPI passes basic validation.
- No secrets committed; add `.env.example`.