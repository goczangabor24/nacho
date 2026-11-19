# OpenCode project instructions — Restaurant Inventory

Read `PROJECT_SPEC.md` at the start of work on this repository. It is the living product and architecture brief. If it conflicts with the actual repository, inspect the code and ask before making assumptions. Do not treat proposed features as already implemented.

## Confirmed project context
- Restaurant inventory app for family owners and waiters.
- Shared restaurant tablet, **separate employee accounts**, auditable inventory measurements.
- Go API, PostgreSQL, TypeScript frontend; Keycloak planned for authentication.
- Docker Compose + AWS EC2 is the initial deployment direction; not on-premises.
- Developer wants to learn, understand, and maintain the system.

## Working rules
1. Inspect relevant code and `go.mod`/frontend manifest before proposing changes.
2. Keep solutions idiomatic, simple, incremental, and testable. Explain unfamiliar choices.
3. Ask before making domain assumptions (measurement units, inventory cycle, edits/corrections, exports) or large architectural changes.
4. Keep Keycloak authentication separate from Go API authorization; verify API **access tokens** properly and derive audit identity from the verified token, never a request body.
5. Design for a shared tablet: visible current user, explicit logout/switch user, clear expired-session behavior, no cross-user state leakage.
6. Preserve migration history and auditability; avoid silently overwriting inventory history.
7. Use parameterized SQL, validation, safe error handling, and tests. Prefer existing project libraries.
8. Keep secrets out of git; never print or log tokens/passwords.
9. Never deploy, spend AWS money, change production infrastructure, or modify real inventory data without explicit approval.
10. At the end of a change, summarize edited files, commands/tests run, and any remaining decisions.

## Current technology direction
- Go `database/sql` + `github.com/jackc/pgx/v5/stdlib` is a suitable DB approach; `sql.DB` already manages a connection pool.
- `github.com/joho/godotenv` may load local `.env` files; production should use injected environment variables.
- TypeScript UI, PostgreSQL, Keycloak OIDC, Docker Compose, Caddy, AWS EC2.
- These are preferences, not instructions to overwrite a working implementation.

## First action in a new session
Read `PROJECT_SPEC.md`, inspect the repo, then ask which feature or issue to work on. If the task is already specified, proceed with that task instead of asking again.
