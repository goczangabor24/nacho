# Restaurant Inventory Management Application — Project Specification

**Status:** Living project brief, not a claim that features already exist.  
**Audience:** Developer and AI coding assistant (OpenCode).  
**Primary goal:** Build a reliable, tablet-friendly restaurant inventory system that is useful in daily operations and serves as a hands-on learning project for Go, TypeScript, PostgreSQL, authentication, Docker, and AWS deployment.

## 1. Confirmed context and constraints

- The application is for a real restaurant operated by the developer's family; waiters will also use it.
- The main client device will be a **shared tablet at the restaurant**.
- **Every employee needs an individual account**; shared generic employee accounts are not acceptable.
- Inventory measurements/stock entries must be attributable to the authenticated employee.
- The application will **not be hosted on-premises**. The intended learning/deployment direction is **Docker and AWS**.
- The intended technology stack is **Go backend, PostgreSQL, TypeScript frontend**, and **Keycloak** for identity management (proposed architecture, not necessarily implemented).
- The user wants to learn implementation and deployment, not merely receive a large AI-generated codebase they cannot understand.
- Access is expected primarily from the restaurant tablet; access from elsewhere is not a current requirement.

## 2. Goals

1. Replace or simplify manual restaurant inventory recording with an intuitive digital workflow.
2. Allow staff to record stock counts or measurements quickly on a shared touch device.
3. Record **who** made each entry and **when**.
4. Separate permissions for owners/managers and waiters.
5. Make it possible to review inventory and export records (exact export format TBD).
6. Operate reliably enough for daily restaurant use, with backups and a recovery plan.
7. Provide a realistic learning path through API design, SQL, auth, containerization, cloud deployment, and maintenance.

## 3. Open product questions — do not invent answers

Clarify these before implementing domain-specific workflows:

- What exactly is inventoried (drinks, ingredients, bottles, kegs, other items)?
- What does a **measurement** mean for each product (whole units, partial bottles, milliliters, weight, percentage, etc.)?
- How are products organized (category, storage location, bar, cellar, shelf)?
- Is inventory captured per shift, daily, weekly, monthly, or on demand?
- Can a waiter edit their own submissions? Can managers correct past records? Must corrections retain history?
- Is the desired output a current-stock view, periodic inventory snapshots, consumption calculations, purchasing reports, CSV/Excel exports, or some combination?
- Are there existing spreadsheets or paper forms that should define the data model and tablet UI?
- What tablet OS/browser and screen size will be used?
- Approximate number of employees and concurrent sessions?
- Does the restaurant require offline entry during internet outages, or is an explicit unavailable-state acceptable for v1?
- Who will operate backups, user onboarding, and cloud billing?

**Do not silently choose business rules.** Ask concise, focused questions when a decision affects data correctness.

## 4. Suggested MVP scope (proposal, not finalized requirements)

### Core
- Employee sign-in/sign-out and a highly visible **Switch user** action.
- Product catalog with name, category, unit of measure, and active/inactive status.
- Touch-friendly inventory entry: select item, enter quantity/measurement, submit.
- Server-generated entry timestamp and server-derived authenticated user ID.
- Inventory history showing item, value, employee, and time.
- Admin interface for product management and reviewing entries.
- Role-aware access controls enforced in the backend.
- Basic export of inventory records (CSV is a reasonable first option, pending confirmation).

### Not necessarily MVP
- Advanced analytics, forecasts, purchasing automation, multi-location support, barcode scanning, notifications, and offline sync.
- Custom PIN-based login. Standard Keycloak login is the initial default; a short PIN alone should not be treated as strong authentication.

## 5. User roles and authorization

Proposed initial roles:

| Role | Proposed abilities |
| --- | --- |
| Admin / owner | Manage products, manage or disable employees (via Keycloak admin workflow), review all records, export, correct records with audit history |
| Waiter | Sign in, view relevant products, create inventory measurements, review their submissions as permitted |
| Manager (optional) | Review/correct entries and access reports without full identity administration |

These permissions are **proposals** and must be confirmed. Always enforce authorization in Go; hiding UI buttons is insufficient. Distinguish `401 Unauthorized` (not authenticated) from `403 Forbidden` (authenticated but lacking permission).

## 6. Architecture

```text
Shared restaurant tablet (TypeScript web app / PWA candidate)
              |
              | HTTPS
              v
      Caddy reverse proxy
         |          |
         v          v
      Go API     Keycloak (OIDC)
         |          |
         v          v
   App PostgreSQL  Keycloak database/schema
```

**Initial deployment proposal:** One AWS EC2 instance running Docker Compose. This minimizes infrastructure complexity for learning. For production, choose a server size based on measured Keycloak/DB memory requirements; one instance is a single point of failure and needs a tested recovery plan.

- **Go:** REST API, validation, business rules, audit metadata, authorization, DB access.
- **PostgreSQL:** Product and inventory data; migrations under version control.
- **TypeScript:** Responsive, touch-first frontend optimized for a shared tablet. Exact framework TBD; do not assume React/Vue without discussing trade-offs.
- **Keycloak:** Individual identities, credentials, login, logout, roles, optional MFA.
- **Caddy:** HTTPS termination, reverse proxy, frontend asset serving.
- **Docker Compose:** Local integration environment and first cloud deployment.
- **AWS EC2:** Initial non-local host; future managed services only if justified.

Keycloak may use its own PostgreSQL database (prefer separate database and credentials from the app). Do not expose PostgreSQL publicly.

## 7. Authentication and shared-tablet behavior

- Use OpenID Connect Authorization Code Flow with PKCE for a browser-based frontend.
- The frontend authenticates with Keycloak; API calls carry access tokens where appropriate.
- The Go API validates access tokens: cryptographic signature against trusted JWKS, issuer, expiration, audience, and any relevant authorized-party/role claims.
- **Do not confuse ID tokens with API access tokens.**
- Never trust a client-provided `recorded_by` field. Derive it from the verified token's stable subject (`sub`).
- Provide an obvious **Log out / Switch user** action, clearing local application state and ending the Keycloak session as appropriate.
- Avoid exposing the previous employee's data after switching users; reset cached views and browser history-sensitive state where feasible.
- Use short, practical idle timeouts for a shared device, balanced against usability; exact values TBD.
- Never put client secrets in browser-delivered code. Do not store access tokens in persistent browser storage without a carefully justified design.
- Protect administrative operations with stronger controls than basic waiter operations.
- Keycloak handles authentication; Go is the final authority for application permissions.

## 8. Proposed data model — illustrative, not final

Possible entities:

- `products`: id, name, category_id, unit, active, created_at, updated_at.
- `categories`: id, name.
- `inventory_sessions` (optional): id, started_at, closed_at, status, created_by.
- `inventory_entries`: id, product_id, session_id (nullable if not needed), measured_quantity (NUMERIC), unit, recorded_by_subject, recorded_at, note.
- `inventory_entry_corrections` or append-only events: preserve who corrected what, when, and why rather than silently overwriting history.

Use UUIDs or another consistent identifier strategy. Prefer PostgreSQL `NUMERIC` for fractional stock measurements where decimal accuracy matters. Store timestamps as `TIMESTAMPTZ` in UTC and render in the restaurant's local timezone. Confirm measurement units and snapshot semantics **before writing migrations**.

**Audit principle:** A record should be traceable to the Keycloak user ID, and ideally preserve a display-name snapshot or resolve names for reporting. Do not assume usernames never change.

## 9. API design principles

Possible endpoints (examples only; adjust to the real workflow):

```text
GET    /healthz
GET    /api/me
GET    /api/products
POST   /api/products                 [admin]
PATCH  /api/products/{id}           [admin]
POST   /api/inventory-entries       [waiter/admin]
GET    /api/inventory-entries       [authorized roles]
GET    /api/inventory-entries/export [admin/manager]
```

- Prefer explicit request/response types and consistent JSON error responses.
- Validate input server-side, including units, ranges, and required fields.
- Use request-scoped contexts and database timeouts.
- Keep handlers thin; put domain rules in services when complexity justifies it.
- Use parameterized SQL and migrations. `database/sql` with `github.com/jackc/pgx/v5/stdlib` is a suitable starting point; `sql.DB` already pools connections.
- If the existing code uses sqlc, preserve it rather than adding a competing query abstraction.
- Add pagination/filtering for history once necessary.
- Protect state-changing requests against replay/duplicate taps where appropriate.

## 10. Frontend UX principles

- Tablet-first, large tap targets, clear typography, simple navigation.
- Minimize steps to record measurements; avoid unnecessary text entry.
- Make the current signed-in employee clearly visible at all times.
- Always show clear save confirmation and recoverable validation errors.
- Avoid accidental duplicate submissions and destructive actions.
- Provide loading, network-error, and expired-session states.
- Optimize for staff who may be busy and unfamiliar with technical interfaces.
- Consider installable PWA behavior for a home-screen icon, but do not promise offline functionality without designing synchronization and conflict handling.

## 11. Configuration and secrets

- Use environment variables for configuration (`DATABASE_URL`, server port, Keycloak issuer/audience, etc.).
- For local development, `github.com/joho/godotenv` is acceptable; `os.Getenv` only reads variables already present in the process environment.
- Validate required settings at startup and return informative errors.
- Do not commit `.env`, passwords, access tokens, private keys, or production credentials.
- Commit a sanitized `.env.example` documenting required variables.
- In production, inject configuration through the deployment environment/secrets mechanism rather than depending on a checked-in `.env` file.

## 12. Local development and deployment phases

### Phase A — Domain and backend
1. Inspect the current repository and identify what already exists.
2. Confirm inventory units, workflows, and roles.
3. Implement migrations and PostgreSQL connectivity.
4. Implement product and inventory endpoints with tests.

### Phase B — Identity
1. Add Keycloak to local Docker Compose.
2. Create a development realm and public frontend client; configure an API audience.
3. Integrate OIDC in the frontend and access-token verification in Go.
4. Test employee switching, token expiry, invalid tokens, and role enforcement.

### Phase C — Tablet UI
1. Implement login and main inventory flow.
2. Add history, corrections (if approved), and export.
3. Test on the actual tablet and with nontechnical users.

### Phase D — Containers and cloud
1. Build a multistage Go Dockerfile and frontend production build.
2. Add health checks, volumes, and least-privilege container settings.
3. Deploy with Docker Compose to AWS EC2.
4. Configure DNS, HTTPS through Caddy, security groups, and a restricted access strategy (e.g. VPN; do not assume a fixed restaurant IP).
5. Set up automated off-instance encrypted backups and test restoration.
6. Add monitoring, alerting, updates, and a rollback procedure.

### Phase E — Production readiness
- Verify backups, account disablement, tablet logout, HTTPS, permissions, error handling, and recovery from network/server outages.
- Document operating procedures and expected AWS costs.
- Decide explicitly whether offline entry is required before daily reliance.

## 13. Testing expectations

- Go unit tests for validation and business rules.
- Integration tests for DB operations against a test PostgreSQL instance.
- Auth tests for missing/expired/invalid tokens, wrong audience, and insufficient roles.
- Multi-user tests ensuring no user can impersonate another via request fields.
- Frontend tests for basic entry flow, loading/errors, and user switching.
- Manual tablet usability testing with real restaurant workflows.
- Restore a backup into a clean environment at least once before production.

## 14. Security and reliability baseline

- HTTPS only in production; use secure Keycloak issuer URLs and correct reverse-proxy configuration.
- Restrict access to the application where feasible (VPN/device-based access is worth considering); still require individual user authentication.
- Never expose database ports or privileged admin endpoints publicly.
- Use least-privilege DB users and separate development/production secrets.
- Keep Docker images, Keycloak, Go dependencies, and the OS patched.
- Back up both **application data and Keycloak identity/configuration data**.
- Avoid logging credentials, full bearer tokens, or sensitive customer/staff data.
- Plan for internet downtime: either display a clear unavailable state or explicitly design offline synchronization later.
- Treat production deployment as more than `docker compose up`: verify persistence, monitoring, restoration, and access controls.

## 15. How OpenCode should collaborate

The developer is learning. Prefer **small, understandable, reviewable changes** over broad autonomous rewrites.

- Start by inspecting actual files, dependencies, and existing architecture; do not assume this specification reflects implemented code.
- Explain the *why* behind architectural choices and introduce unfamiliar concepts briefly.
- Reuse existing conventions and libraries when sensible.
- Present a short plan before significant changes; ask before changing frameworks, authentication strategy, schema semantics, or deployment architecture.
- When implementing, modify only relevant files, show what changed, and describe how to run/test it.
- Do not introduce elaborate abstractions, unnecessary dependencies, microservices, or Kubernetes.
- Never invent restaurant business requirements; distinguish confirmed facts from recommendations.
- Prefer Go's standard library where practical, but use established libraries for hard problems (OIDC, PostgreSQL driver, `.env` parsing).
- Flag security-sensitive changes and include appropriate tests.
- Do not deploy to AWS, create billable resources, or modify production data without explicit approval.
- Preserve the developer's ability to learn: answer conceptual questions directly and avoid dumping full solutions when the developer requests guidance only.

## 16. Definition of done for an initial release

A first release is ready only when restaurant staff can individually sign in on the shared tablet, reliably enter and review agreed inventory measurements, and have entries accurately attributed to their accounts; admins can manage the relevant catalog and review records; permissions are enforced server-side; the application is available over HTTPS on the intended infrastructure; and backup/recovery and shared-device logout have been tested.

---

**Document maintenance:** Update this file whenever a business rule or major architectural decision is confirmed. Mark proposals as confirmed only after the developer approves them.
