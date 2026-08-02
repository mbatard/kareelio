# PLAN.md

## Objective

Notify the administrator when a new user registers: send an admin email notification to the valid active admin account email, and show an in-app admin notification badge whose count increments with newly registered users until acknowledged.

## Constraints

- Protected `main`: feature branch + PR only; no direct pushes to `main`.
- Planning only modified `PLAN.md`; implementation must start from a clean worktree/branch based on `origin/main`.
- Do not expose secrets, SMTP credentials, passwords, session IDs, CSRF values, verification tokens, or raw request bodies in logs or emails.
- Public registration responses must remain generic/non-blocking enough to avoid account enumeration and poor UX.
- Admin notification email must be skipped if the active admin email is invalid or local/dev-only, especially `@kareelio.local`.
- Do not introduce plaintext secrets; use existing SMTP env vars/Kubernetes secrets only.
- Web notification endpoints must be admin-only behind existing auth + CSRF middleware.
- Prefer small, reversible, testable steps.
- Any database migration must be additive/idempotent where practical, with rollout and rollback documented.
- Kubernetes/Docker config changes are not expected; if implementation discovers they are needed, include dry-run/diff validation and rollback before applying.

## Current State

- `origin/main` is currently at `46cb28c` / tag `v1.3.0`; PR #46 for localized verification emails has been merged.
- The primary worktree `/Users/mikael/kareelio` is still on an old dirty branch (`fix/applications-status-colors`) with many pending changes. Implementation should not use it directly; use a clean feature branch/worktree from `origin/main`.
- Backend public registration lives in `backend/internal/handler/auth.go` and logs/audits `model.AuditActionUserRegistered` after creating the user and verification token.
- Public verification email delivery is asynchronous for registration/public resend; admin forced resend remains synchronous.
- `mailer.Mailer` currently supports localized verification emails via `SendVerificationEmail(requestID, to, token, language)` and safe SMTP logs.
- Admin account is stored as a normal `users` row with `role='admin'`; `database.SeedAdmin` creates the default admin from `DEFAULT_ADMIN_EMAIL`, whose fallback is `admin@kareelio.local`.
- `UserRepository` can list users and get users by ID/email, and now has a helper to fetch the active admin contact.
- Admin notification email delivery now exists in the feature worktree, but there is still no frontend badge.
- `admin_notification_state` storage, repository helpers, and the admin notification API now exist in the feature worktree: per-admin ack state is persisted, counts/ack can be driven from audit events, and admin-only GET/POST routes expose the notification summary and acknowledgement.
- Admin routes are grouped under `RequireRole(model.RoleAdmin)` in `backend/internal/router/router.go`; examples: `/api/admin/dashboard`, `/api/admin/audit`, and `/api/users`.
- Existing audit events include `user_registered`, so unread counts can be derived from audit events if a per-admin acknowledgement marker is stored.
- Frontend admin navigation is in `frontend/src/components/Navbar.tsx`; the Admin Users link now shows an unread-count badge for admin users.
- Frontend admin API helpers are in `frontend/src/services/api.ts`; notification summary/ack helpers now exist.
- `AdminUsersPage` already loads users and is a natural place to acknowledge new-user notifications when the admin views the user list.
- No Docker/Kubernetes manifest change appears necessary for this feature if existing SMTP config is reused.

## Tasks

- [x] Add backend notification state storage and repository helpers.
  - Add an additive migration, e.g. `009_create_admin_notification_state.up.sql`, for per-admin notification acknowledgement state.
  - Recommended table: `admin_notification_state(admin_user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE, user_registrations_seen_at TIMESTAMPTZ NOT NULL DEFAULT 'epoch', updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW())`.
  - Add repository methods to count unacknowledged public registration audit events (`action='user_registered'`, `target_type='user'`) created after the admin's `seen_at`, and to acknowledge current registrations for the acting admin.
  - Keep the migration additive and reversible via a matching `.down.sql` dropping only the new table.
  - Add targeted repository tests where practical; otherwise cover counting/ack behavior via handler tests with fakes.
  - Findings: added `admin_notification_state` with per-admin `user_registrations_seen_at`, implemented repository helpers to count public `user_registered` audits since the last ack and upsert the ack timestamp, and validated the migration on a disposable local PostgreSQL via Docker Compose. Verified with `go test ./internal/repository`, `go test ./...`, and `go build ./...`.

- [x] Add backend admin notification API.
  - Add admin-only routes, e.g. `GET /api/admin/notifications` and `POST /api/admin/notifications/user-registrations/ack`.
  - Response shape recommendation: `{ "new_user_registrations": number }`.
  - Use the authenticated admin user from context for per-admin state.
  - Require auth/admin/CSRF through the existing admin route group.
  - Add safe logs for notification count/ack actions without listing all user emails.
  - Add targeted handler tests for admin count, ack, unauthenticated/non-admin denial where feasible.
  - Findings: added `GET /api/admin/notifications` returning `{ "new_user_registrations": number }`, `POST /api/admin/notifications/user-registrations/ack` returning 204, safe log events for count/ack, handler tests for success/error/unauthorized cases, and middleware coverage confirming `RequireRole(model.RoleAdmin)` blocks non-admin access.

- [x] Send admin email notification after public registration.
  - Add a mailer method such as `SendAdminNewRegistrationEmail(requestID, adminEmail, registeredUserEmail, registeredDisplayName, language)`.
  - Fetch the active admin user email from the database at registration time or via a small admin-email repository method.
  - Validate admin email with existing email validation and skip if invalid or dev/local-only (`@kareelio.local`, optionally any `.local` domain).
  - Email content must not include passwords, verification tokens, session IDs, or raw request payloads; include only safe fields such as registered user email/display name and a link to `/admin/users`.
  - Send asynchronously so public registration latency stays low; log success/skip/error with request ID and safe reason.
  - Add tests for valid admin email send, invalid/local admin email skip, SMTP error logging, and no token/password leakage.
  - Findings: added `UserRepository.GetActiveAdmin`, `Mailer.SendAdminNewRegistrationEmail`, and an async registration hook that looks up the active admin, skips invalid/local admin addresses, and logs safe queue/skip/error/completion events. Added tests covering valid delivery, local admin skip, SMTP errors, and no token/password leakage in both mail content and logs. Verified with `go test ./internal/mailer ./internal/handler ./internal/repository`, `go test ./...`, and `go build ./...`.

- [x] Add frontend admin notification badge.
  - Extend `frontend/src/types/index.ts` and `adminApi` in `frontend/src/services/api.ts` for notification summary and ack endpoints.
  - Update `Navbar` so admin users see a badge on/near the Admin Users navigation item when `new_user_registrations > 0`.
  - Fetch the summary when an admin session is active and refresh periodically with a modest interval (recommended 60s) rather than websockets for this first iteration.
  - Acknowledge/reset the badge when the admin opens `/admin/users` or clicks the Admin Users link, then refresh the count.
  - Add loading/error handling that fails quietly; notification failures must not block navigation.
  - Add i18n keys in both `fr.json` and `en.json` if labels/tooltips/accessibility text are introduced.
  - Findings: added `AdminNotificationSummary`, `adminApi.notificationSummary()` and `adminApi.acknowledgeUserRegistrations()`, updated `Navbar` to poll every 60s, display an admin-only badge, and acknowledge/reset on `/admin/users`, and added accessible badge labels in both locales. Verified with `npm ci`, `npm run lint`, `npm run build`, and `npm audit --audit-level=high` (audit still reports pre-existing moderate `react-router` advisories only).

- [x] Validate end-to-end behavior locally.
  - Use a disposable SMTP capture service; do not use real production SMTP credentials.
  - Register a new user with a valid active admin email and verify: user verification email sent, admin notification email sent, badge increments for logged-in admin.
  - Verify no admin notification email is sent when active admin email is `admin@kareelio.local` or otherwise invalid/local-only; logs should show a safe skip reason.
  - Verify badge resets only after admin acknowledgement and remains admin-only.
  - Verify public registration response remains fast and generic.
  - Findings: validated against a disposable local PostgreSQL + custom SMTP capture server with a local backend process. Confirmed first registration sent both verification and admin notification emails, duplicate registration returned the generic fallback response, unread count progressed 1 → 2 → 0 after acknowledgement, admin notification email was skipped after switching the active admin to `admin@kareelio.local`, backend logs recorded the safe skip reason, and a regular user received 403 from the admin notifications API. Frontend badge behavior was also covered with a focused `Navbar` component test, and `npm run lint`, `npx vitest run src/components/Navbar.test.tsx`, `npm run build`, and `npm audit --audit-level=high` were run in the frontend worktree.

- [x] Prepare production rollout and rollback notes.
  - No Kubernetes manifest changes are expected; if none are made, validate application builds/tests only plus standard deployment checks.
  - If manifests/config change unexpectedly: run `kubectl apply --dry-run=server -f deploy/k8s/` and `kubectl diff -f deploy/k8s/` before rollout.
  - For the DB migration, confirm it is additive and safe under rolling deployment: older pods should ignore the new table; newer pods require it after migration.
  - Findings: no Kubernetes/Docker manifest changes were needed for this feature. Rollout can use the normal backend/frontend image deployment after the additive DB migration; rollback is the previous backend/frontend image tag, and if notification state must be removed the matching down migration only drops `admin_notification_state`.

## Validation

- Backend targeted validation:
  - `cd backend && go test ./internal/mailer ./internal/handler ./internal/repository`.
  - `cd backend && go test ./...`.
  - `cd backend && go build ./...`.
- Frontend validation:
  - `cd frontend && npm ci` if dependencies are not installed in the worktree.
  - `cd frontend && npm run lint`.
  - `cd frontend && npm run build`.
  - `cd frontend && npm audit --audit-level=high`.
- Migration validation:
  - Ensure new `.up.sql` and `.down.sql` are idempotent/reversible where practical.
  - Run local backend migrations against disposable PostgreSQL if available.
  - Verify rollback down migration only removes notification state, not users/audit events.
- SMTP validation:
  - Use a local SMTP capture service to verify both verification emails and admin registration notification emails.
  - Confirm admin notification emails are skipped for `@kareelio.local` and invalid admin emails.
  - Confirm logs do not leak passwords, SMTP credentials, verification tokens, session IDs, cookies, or raw request bodies.
- Admin web validation:
  - As admin, create/register multiple users and verify badge count increments to 1, 2, etc.
  - Visit `/admin/users` or click the badge/link and verify ack resets the count.
  - Verify normal users cannot access notification APIs.
- Kubernetes/production validation if manifests change:
  - `kubectl apply --dry-run=server -f deploy/k8s/`.
  - `kubectl diff -f deploy/k8s/`.
  - After deployment: `make deploy-status` and `make deploy-logs`.

## Risks

- Email notifications can leak personal data if too verbose; keep content minimal and safe.
- Admin notification email delivery can fail due to SMTP issues; this must not block public registration.
- Sending to default/dev admin addresses like `admin@kareelio.local` would create noise or bounces; validate and skip.
- Badge counts can become confusing if based only on total users; use explicit acknowledgement state for newly registered users.
- Multi-admin behavior needs clear semantics; per-admin acknowledgement is recommended to avoid one admin clearing another admin's badge.
- Polling too frequently can add unnecessary backend load; use a modest interval.
- Database migration bugs could affect startup if migrations run automatically; keep migration additive and simple.
- If acknowledgement uses audit timestamps, clock/database precision edge cases should not double-count around the ack boundary.

## Rollback

- Application rollback: revert the PR or redeploy the previous backend/frontend image tag.
- Frontend-only issue: hide/remove the badge and ack calls by reverting frontend changes; backend notification endpoints/table can remain unused temporarily.
- Admin email issue: revert mailer/registration notification changes or disable by making admin email invalid/local-only as an emergency operational workaround, while keeping user verification email intact.
- Database rollback:
  - If no production data in notification state needs preservation, run the matching down migration to drop only `admin_notification_state`.
  - Otherwise leave the additive table in place and roll back application images; old images should ignore the table.
- Kubernetes rollback if manifests change:
  - Reapply previous ConfigMap/Secret/manifests.
  - Redeploy previous version with `make deploy VERSION=<previous>`.

## Notes / Decisions

- Decision: treat the badge count as unacknowledged **public user registrations**, not total users and not admin-created users.
- Decision: store notification acknowledgement per admin user.
- Decision: acknowledge the badge when the admin opens/clicks Admin Users, unless implementation discovers a better explicit UI action is needed.
- Decision: use existing SMTP configuration; do not add new secrets for this feature.
- Decision: send admin notification email asynchronously and skip invalid/local admin emails.
- Assumption: there is only one active admin in normal production, but the implementation should support multiple active admins safely.
- Assumption: the active admin row email is the target notification email; `DEFAULT_ADMIN_EMAIL` is only a seed value and should not be the runtime source if the admin profile has changed.
- Unresolved: exact French/English wording of the admin notification email subject/body; start with concise bilingual-safe text or use admin stored language if available.
- Unresolved: whether acknowledgement should happen automatically on `/admin/users` load or require an explicit click; recommended first behavior is automatic ack on `/admin/users` because it is simple and reversible.
- Next `/next` task for `platform-build`: none; the feature work is complete and only documentation/PR hygiene remains.
