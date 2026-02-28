# AGENTS

This file defines standards for AI agents and human contributors working in this repository.

## Mission

Deliver reliable, testable changes to Mission Control without leaking secrets or breaking execution workflows.

## Tech Stack

- Backend: Go + Echo + SQLite + sqlc
- Frontend: Next.js + TypeScript + Tailwind + Zustand
- Runtime: single binary serving API, WebSocket, and embedded UI

## Core Principles

- Prefer modular, reusable functions over one-off implementations
- Favor memory-efficient approaches and avoid unnecessary allocations
- Add meaningful logging for state transitions, failures, retries, and external calls
- Keep changes minimal, localized, and easy to review
- Preserve backward compatibility for API responses when possible

## Security Rules

- Never commit secrets (`.env`, tokens, private keys, certs)
- Never hardcode credentials in source
- Use `.env.example` placeholders only
- Avoid logging raw tokens, secret values, or full credentialed URLs
- Treat local DBs and artifacts as non-source files

## Backend Standards (Go)

### Project layout

- API handlers: `internal/api/handlers/`
- Route composition: `internal/api/server.go`
- Store layer: `internal/store/`
- SQL definitions: `internal/db/queries/*.sql`
- Migrations: `internal/db/migrations/`

### Coding expectations

- Keep handler logic thin; move data operations into store/query layers
- Return explicit HTTP status codes and structured JSON payloads
- Handle nullable DB fields carefully (`sql.NullString`, `sql.NullTime`, etc.)
- Use context-aware store calls (`ctx := c.Request().Context()`)
- Keep functions focused and composable

### Logging expectations

- Log important lifecycle transitions: task start/stop/retry/complete/fail
- Log integration failures with OpenClaw and include actionable context
- Log errors with operation and entity IDs
- Use warning level for recoverable faults and error level for failed operations

## Frontend Standards (TypeScript/React)

- Keep page-level orchestration in `ui/src/app/*`
- Keep reusable UI in `ui/src/components/*`
- Centralize API access in `ui/src/services/api.ts`
- Keep shared state in Zustand stores under `ui/src/stores/*`
- Prefer typed interfaces from `ui/src/types/index.ts`
- Keep components small and composable

## API and Data Contract Rules

- Add/modify endpoint behavior in handlers and route registration together
- If schema changes, create a migration and update SQL query files
- Regenerate sqlc code after query/schema updates
- Ensure task status transitions remain valid and documented
- Keep WebSocket event payloads consistent with frontend expectations

## Database and Migration Rules

- One migration per logical schema change
- Include both `.up.sql` and `.down.sql`
- Never edit historical migrations after they are shared
- Validate query changes against generated sqlc types

## Testing and Validation

Before finalizing changes:

1. Run relevant unit/integration tests (`make test` or scoped commands)
2. Run linters (`make lint`)
3. Verify build succeeds (`make build`) when touching cross-layer code
4. For UI-affecting changes, verify key pages and WebSocket-driven updates

## Performance and Reliability

- Avoid repeated full-table reads in hot paths
- Prefer bounded queries with limit/order where practical
- Be explicit with retry and timeout behavior
- Keep queue/watchdog changes deterministic and observable

## Documentation Rules

- Update `README.md` when setup, run, or deployment steps change
- Update `docs/API.md` when API contracts change
- Update `docs/ARCHITECTURE.md` for major structural changes
- Update `docs/AGENT_INTEGRATION.md` when workflow/protocol behavior changes
- Before opening a PR, summarize any temporary AI-generated QA/report docs into canonical docs and delete the temporary files
- Temporary docs include patterns like `docs/qa-*.md`, branch reports, and ad-hoc analysis notes
- Consolidation targets:
  - Runtime/protocol/process findings -> `docs/AGENT_INTEGRATION.md`
  - System/component behavior findings -> `docs/ARCHITECTURE.md`
  - Setup/run/developer workflow findings -> `README.md`

## Commit and PR Guidance

- Use clear, scoped commit messages (what changed and why)
- Keep unrelated edits out of the same commit
- Include rollout/risk notes for infrastructure and workflow changes
- Include test evidence in PR descriptions

## Execution Safety Checklist for Agents

Before editing:

- Read affected files fully
- Check for existing patterns and conventions
- Confirm no secrets are introduced

After editing:

- Re-run checks
- Re-scan for accidental sensitive values
- Ensure docs remain aligned with behavior
