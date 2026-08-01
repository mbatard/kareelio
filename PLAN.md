# PLAN.md

## Objective

Improve observability and email-verification operations for Kareelio so that account registration, verification-email sending, resend actions, and admin user operations produce explicit, safe frontend/backend logs and actionable errors, and add an admin capability to force-resend a user creation/verification email.

## Constraints

- Protected `main`: feature branch + PR only; no direct pushes to `main`.
- Do not expose secrets, SMTP credentials, session IDs, verification tokens, password values, CSRF values, or raw request bodies in logs.
- Logs may include request ID, route/action, status, duration, user ID, target user ID, and normalized email only where already visible to the acting user/admin; avoid logging verification links in production.
- Preserve generic public registration responses where needed to avoid account enumeration, while logging internal failure reasons server-side.
- Keep changes incremental and reversible.
- Frontend logs should be useful during debugging but must not leak passwords, tokens, or private form contents.
- Admin resend endpoint must remain admin-only, protected by existing auth/CSRF middleware, and must not allow public token generation.
- SMTP changes must support the reported self-hosted unauthenticated SMTP path (`SMTP_HOST` + `SMTP_PORT=25` + `SMTP_FROM`) without requiring username/password.
- Kubernetes/Docker/config changes must include rollout impact, validation, and rollback.

## Current State

- Backend request logging emits method/path/status/duration/request ID in `backend/internal/middleware/middleware.go`, e.g. `POST /api/auth/register`, and now excludes `/api/healthz` and `/api/readyz` probe traffic.
- `AuthHandler.Register` now emits structured lifecycle/error logs and records mail-send success/failure context while keeping the public response generic on internal failures.
- `AuthHandler.ResendVerification` now emits structured lifecycle/error logs and records mail-send success/failure context while keeping the public response generic on internal failures.
- `mailer.SendVerificationEmail` now logs safe send start/success/error events and avoids leaking verification tokens or full links in logs.
- `mailer.send` now uses an explicit SMTP client flow with request timeouts, supports unauthenticated SMTP when username/password are empty, and returns phase-specific errors for connect/starttls/auth/sender/recipient/data/quit failures.
- `.env.example` documents SMTP defaults for the unauthenticated relay path as `SMTP_PORT=25`; `docker-compose.yml` defaults `SMTP_PORT` to `25`; production `deploy/k8s/configmap.yaml` now sets `SMTP_PORT: "25"` and `SMTP_FROM`, while `SMTP_HOST` comes from `kareelio-secret` if configured and `SMTP_USERNAME`/`SMTP_PASSWORD` can be left empty for unauthenticated relay.
- For a self-hosted relay on port 25, the backend must actually receive `SMTP_PORT=25`; frontend pod connectivity does not prove backend pod SMTP config or delivery path.
- Backend now emits an explicit SMTP startup/config summary log.
- Backend now emits a safe data-encryption startup/config summary log without exposing encryption key values.
- Frontend `frontend/src/services/api.ts` now has env-gated safe axios request/success/error logs that emit action names, HTTP status, and backend request IDs when available.
- `RegisterPage` now logs submit start/success/error without exposing passwords or tokens.
- Frontend Nginx now exposes dedicated `/healthz` and `/readyz` endpoints with access logs disabled for probe traffic.
- Admin user management exists in `AdminUsersPage` and `AdminUserEditPage`; admin can create/edit/delete users, reset passwords, and resend verification for unverified non-admin users from the list and edit pages.
- Backend now exposes `POST /api/users/{id}/resend-verification` for admin-only forced verification-email resend on unverified non-admin users.
- Admin-created users currently go through `UserHandler.Create`, which creates an active user but does not create or send an email-verification token.
- Public resend endpoint `/api/auth/resend-verification` exists for unverified users by email; admin resend is available for selected unverified non-admin users.
- Cluster diagnostics for image `sha-d79c4b3` confirmed backend/frontend pods were using the expected image digests, but `DATA_ENCRYPTION_KEY`, `DATA_ENCRYPTION_KEY_ID`, and `JOB_APPLICATIONS_REQUIRE_ENCRYPTED_READS` were missing from the live backend environment.
- `.env.example`, `docker-compose.yml`, and `deploy/k8s/secret.example.yaml` now document/pass job application encryption settings so operators can align runtime env with the manifests.
- Audit actions already include `email_verification_resent`; audit UI translations exist for that action.
- Existing pending local changes include previous DB-encryption work and opencode model config changes; keep this new plan separate and do not implement code during planning.

## Tasks

- [x] Add safe structured backend action/error logs for auth and mail paths.
  - Add explicit logs for `register.start`, `register.user_created`, `register.token_created`, `mail.verification_send_start`, `mail.verification_send_success`, `mail.verification_send_error`, `register.completed`, and analogous `resend_verification.*` events.
  - Include request ID and safe identifiers where available; never log password, SMTP password, verification token, or full verification URL.
  - Stop silently swallowing `SendVerificationEmail` errors: log them with context and keep public responses generic where account enumeration protection is required.
  - Add a startup/config summary log for SMTP with host present/absent, port, from present/absent, auth enabled/disabled, and TLS mode; never log SMTP credentials.
  - Add targeted Go tests for mailer/config/log helper behavior where practical.
  - Findings: backend now emits request-scoped auth/mail event logs, SMTP startup summary logs, and mailer send success/error logs without leaking tokens or credentials; verified with `go test ./...` and `go build ./...` in `backend/`.

- [x] Fix and harden SMTP sending diagnostics for self-hosted unauthenticated relay.
  - Verify support for `SMTP_HOST` + `SMTP_PORT=25` + `SMTP_FROM` without auth.
  - Add explicit connection/send phase errors so operators can distinguish DNS/connect, STARTTLS/TLS, sender, recipient, and DATA failures.
  - Consider adding `SMTP_TIMEOUT_SECONDS` with a safe default if the current `net/smtp.SendMail` path can hang or hide connection timing.
  - Validate locally with a disposable SMTP capture service, e.g. Mailpit/MailHog, without committing real SMTP secrets.
  - Findings: backend now uses an explicit SMTP client flow with request-timeout support, phase-specific errors (`smtp connect`, `smtp starttls`, `smtp sender`, `smtp recipient`, `smtp data`, etc.), and a local disposable SMTP capture test that confirms unauthenticated port-25 delivery works without leaking verification tokens. Verified with `go test ./...` and `go build ./...` in `backend/`.

- [x] Align local and production SMTP configuration examples.
  - Update `.env.example`, `docker-compose.yml`, `deploy/k8s/configmap.yaml`, and/or `deploy/k8s/secret.example.yaml` only as needed to make unauthenticated port-25 relay configuration clear.
  - Decide whether production `SMTP_PORT` should be changed from `587` to `25` based on the user-provided self-hosted relay setup.
  - For K8s changes, include rollout impact, dry-run validation, and rollback commands.
  - Findings: examples and production config now default to `SMTP_PORT=25`, with unauthenticated relay clearly documented in `.env.example`, `docker-compose.yml`, and `deploy/k8s/secret.example.yaml`. Verified with `docker compose config`; `kubectl` is not installed in this environment, so Kubernetes dry-run validation could not be executed here.

- [x] Add safe frontend action/error logs around API calls and registration UI.
  - Add opt-in or environment-gated frontend logging to `frontend/src/services/api.ts` so request start/success/error can be correlated with backend request IDs when available.
  - Log action names and HTTP status only; never log passwords, tokens, request bodies, cookies, or raw auth headers.
  - Add explicit register page logs for submit start/success/error, again without logging password/token.
  - Keep user-facing UI behavior unchanged unless a backend error should now be surfaced more clearly.
  - Findings: frontend now emits env-gated safe API lifecycle logs keyed by action names and request IDs, plus register submit lifecycle logs without leaking passwords/tokens. Verified with `npm run lint` and `npm run build` in `frontend/`.

- [x] Add backend admin endpoint to force resend a verification/user-creation email.
  - Add an admin-only route such as `POST /api/users/{id}/resend-verification` under existing authenticated/admin/CSRF-protected routes.
  - For non-admin target users only, delete existing verification tokens for the target user, create a new token, send the verification email, and audit `email_verification_resent` with target user metadata.
  - Decide behavior for already verified users: recommended first behavior is return `409 Conflict` or `400 Bad Request` and log a safe reason rather than sending a new token.
  - Return actionable admin-facing errors on mail send failure, while avoiding token exposure.
  - Add targeted Go tests for success, already-verified user, missing user, and mail send error if handler dependencies can be tested with fakes; otherwise document test limitations.
  - Findings: backend now exposes an admin-only resend endpoint that blocks already-verified/admin targets, regenerates verification tokens for unverified users, sends the mail, and audits `email_verification_resent`. Verified with `go test ./...` and `go build ./...` in `backend/`.

- [x] Add frontend admin UI for force resend.
  - Extend `userApi` with `resendVerification(id)`.
  - Add a button in `AdminUsersPage` and/or `AdminUserEditPage` for unverified non-admin users.
  - Show loading/success/error states and refresh user data after success.
  - Add i18n keys in both `fr.json` and `en.json`.
  - Findings: frontend admin edit page now exposes an unverified-user resend action with loading/success/error states, refreshes the user record after success, and uses new i18n keys in both locales. Verified with `npm run lint` and `npm run build` in `frontend/`.

- [x] Validate end-to-end mail flow in local/dev and production-safe rollout.
  - Local: run backend against a disposable SMTP capture service and verify logs show send start/success without sensitive token values.
  - Backend: `cd backend && go test ./... && go build ./...`.
  - Frontend: `cd frontend && npm run lint && npm run build`.
  - Docker/K8s if manifests change: `docker compose config`; `kubectl apply --dry-run=server -f deploy/k8s/`; `kubectl diff -f deploy/k8s/` where available.
  - Production rollout: deploy backend/frontend images, verify `/api/auth/register`, public resend, and admin force resend against server-mail logs.
  - Findings: local end-to-end validation succeeded with a disposable SMTP capture server and live backend against local PostgreSQL. Verified three mail flows (`/api/auth/register`, `/api/auth/resend-verification`, and admin `POST /api/users/{id}/resend-verification`) produced SMTP deliveries and request-scoped logs without leaking verification links in backend logs. `docker compose config`, `go test ./...`, `go build ./...`, `npm run lint`, and `npm run build` all passed; `kubectl` is not installed in this environment, so server-side Kubernetes dry-run/diff could not be executed here.

- [x] Mask frontend Nginx live/ready probe logs.
  - Add dedicated `/healthz` and `/readyz` endpoints in `frontend/nginx.conf` with `access_log off` so liveness/readiness probes do not spam access logs.
  - Update the frontend Kubernetes probes to use the new probe endpoints instead of `/`.
  - Findings: frontend probe traffic now returns `204` from Nginx without access-log spam, and the Kubernetes frontend probes now target `/healthz` and `/readyz`. Verified with `git diff --check`, `docker build -t kareelio-frontend-probe-test ./frontend`, and `docker compose up -d --build postgres backend frontend` followed by curl checks against `/healthz` and `/readyz`. `kubectl` is not installed in this environment, so server-side dry-run/diff could not be executed here.

- [x] Add runtime diagnostics and operator fixes for missing backend behavior.
  - Add safe backend startup diagnostics for data-encryption configuration and job application encryption flags, without logging key values or key IDs.
  - Exclude backend `/api/healthz` and `/api/readyz` from request logs so registration/admin logs are not drowned by probe traffic.
  - Document `DATA_ENCRYPTION_KEY_ID` and `DATA_ENCRYPTION_KEY` in Kubernetes secret examples, local env examples, and README operator guidance.
  - Pass job application encryption env vars through Docker Compose for local parity.
  - Add a resend verification action directly in `AdminUsersPage` for unverified non-admin users, reusing the existing admin endpoint/i18n keys.
  - Findings: cluster outputs showed the deployed `sha-d79c4b3` images were correct, but live backend env was missing encryption keys and `JOB_APPLICATIONS_REQUIRE_ENCRYPTED_READS`; backend now logs those booleans at startup, examples document the required secret values, backend health probes no longer pollute request logs, and the resend action is visible in the admin user list. Verified with `go test ./...`, `go build ./...`, `npm ci`, `npm run lint`, `npm run build`, and `docker compose config`. `kubectl` is not installed in this environment, so server-side Kubernetes dry-run/diff could not be executed here.

## Validation

- Backend targeted tests:
  - `cd backend && go test ./internal/mailer/...` if mailer tests are added.
  - `cd backend && go test ./internal/handler/...` if handler tests are added.
  - `cd backend && go test ./...`.
  - `cd backend && go build ./...`.
- Frontend validation:
  - `cd frontend && npm run lint`.
  - `cd frontend && npm run build`.
- SMTP validation:
  - Use a local disposable SMTP capture service; do not use or commit real production SMTP credentials.
  - Verify server logs include SMTP config summary, send start, success/error, and request ID/action context.
  - Verify logs do not include verification tokens, full verification URLs, passwords, cookies, or request bodies.
- Admin resend validation:
  - As admin, resend for an unverified non-admin user and verify audit event + SMTP capture delivery.
  - Verify resend is unavailable/fails safely for admin users and already verified users.
  - Verify unauthenticated and non-admin users cannot call the endpoint.
- Kubernetes/config validation if manifests change:
  - `kubectl apply --dry-run=server -f deploy/k8s/` before apply.
  - `kubectl diff -f deploy/k8s/` where available.
  - Rollout check: `make deploy-status`, `make deploy-logs` after deployment.

## Risks

- Logging too much can leak emails, verification links, tokens, passwords, or other sensitive data.
- Public registration must avoid account enumeration even while backend logs become more explicit.
- Surfacing mail failures directly from public registration could expose operational state; prefer server logs and generic public response.
- SMTP behavior differs by server: port 25 unauthenticated relay, STARTTLS on 587, implicit TLS on 465, sender policies, SPF/relay ACLs, and HELO requirements may fail differently.
- Admin resend can be abused for email spam if not restricted/rate-limited/audited.
- Frontend console logs may be visible to end users; keep them opt-in or non-sensitive.
- K8s SMTP config changes can break verification email delivery if host/port/from are wrong.

## Rollback

- Backend logging/mailer changes: revert the PR or redeploy the previous backend image tag.
- Frontend logging/admin UI changes: revert the PR or redeploy the previous frontend image tag.
- Admin resend endpoint: disable/hide frontend button by reverting frontend only, or revert backend endpoint if server behavior is problematic.
- K8s config changes:
  - Reapply previous ConfigMap/Secret values.
  - Redeploy previous image/version with `make deploy VERSION=<previous>` if needed.
  - Keep previous SMTP settings documented before changing production config.
- If logs accidentally expose sensitive data, immediately stop affected pods/roll back, rotate exposed tokens/secrets as appropriate, and purge log retention where possible.

## Notes / Decisions

- Assumption: the desired mail flow is verification email on public signup and admin-forced resend for unverified users.
- Assumption: the self-hosted SMTP relay should work without SMTP username/password when traffic originates from allowed pod/node IPs.
- Resolved: production should use `SMTP_PORT=25` in `deploy/k8s/configmap.yaml` for the self-hosted unauthenticated relay path.
- Unresolved: confirm whether admin-created users should also receive a verification email automatically, or whether only the new admin force-resend action is required.
- Unresolved: choose frontend logging gate (`import.meta.env.DEV`, `VITE_DEBUG_LOGS=true`, or always log safe action/status lines).
- First `/next` task for `platform-build`: add safe structured backend logs for registration, public resend, and mailer send start/success/error paths, including SMTP config summary, without changing SMTP behavior or adding the admin resend endpoint yet.
