# PLAN.md

## Objective

Address the review feedback for PR #47 (`feat(admin): add registration notifications`) with small, production-safe follow-up changes: bound the async admin notification lookup, add missing invalid-admin-email test coverage, clean stale planning notes, and explicitly triage the remaining React Router audit finding.

## Constraints

- Protected `main`: feature branch + PR only; no direct pushes to `main`.
- Current feature branch is `fix/admin-notifications`; keep PR #47 as the integration path unless the user requests a separate branch/PR.
- Planning step modifies only `PLAN.md`; source code changes start with `/next`.
- Keep public registration non-blocking and generic; admin notification failures must not affect registration success.
- Do not expose secrets, SMTP credentials, passwords, session IDs, CSRF values, verification tokens, cookies, or raw request bodies in logs/tests/emails.
- Use existing SMTP env vars only; do not introduce new plaintext secrets.
- Web notification endpoints must remain admin-only behind existing auth + CSRF middleware.
- Prefer small, reversible, testable changes.
- No Docker/Kubernetes/Terraform changes are expected. If implementation unexpectedly touches infra, include dry-run/diff validation and rollback before applying.
- Do not perform a major React Router upgrade in this PR unless explicitly approved; the audit fix currently points to a breaking v7 upgrade.

## Current State

- Branch/worktree analyzed: `/tmp/opencode/kareelio-admin-notifications` on `fix/admin-notifications`, tracking `origin/fix/admin-notifications`.
- Current committed feature head is `c390b80 feat(admin): add registration notifications`; PR #47 is open against `main`.
- `origin/main` remains at `46cb28c` / tag `v1.3.0`.
- The branch was clean before this planning update; this plan update intentionally dirties only `PLAN.md`.
- Review feedback identified four items:
  - Medium: `backend/internal/handler/admin_notification_email.go` uses `context.Background()` for async admin lookup without a timeout, so stalled DB operations could accumulate goroutines.
  - Low: invalid non-local admin email skip is implemented but not directly tested; only `admin@kareelio.local` skip is covered.
  - Low: old `PLAN.md` current-state text contradicted completed frontend badge work.
  - Medium/residual risk: `frontend/package.json` / `package-lock.json` still use `react-router-dom`/`react-router` 6.30.4; `npm audit --audit-level=high` passes, but `npm audit` reports moderate React Router advisories whose available fix is a breaking v7 upgrade.
- Existing backend notification code is in `backend/internal/handler/admin_notification_email.go`, `backend/internal/handler/admin_notifications.go`, `backend/internal/repository/admin_notification.go`, and `backend/internal/repository/user.go`.
- Existing frontend badge code is in `frontend/src/components/Navbar.tsx`, with focused test coverage in `frontend/src/components/Navbar.test.tsx`.
- No Kubernetes/Docker manifests were changed for the feature; the DB migration remains additive (`admin_notification_state`).

## Tasks

- [x] Bound the async admin notification lookup.
  - Update `sendAdminNewRegistrationEmailAsync` / `sendAdminNewRegistrationEmail` so the background active-admin lookup uses a bounded context (recommended small constant timeout, e.g. 5 seconds) and always calls `cancel()`.
  - Preserve the current behavior: registration response remains non-blocking; timeout/admin lookup errors only produce safe logs and skip admin notification email.
  - Add or adjust a focused handler test so a fake admin lookup can assert that it receives a context with a deadline.
  - Do not change frontend code or dependency files in this step.
  - Findings: `sendAdminNewRegistrationEmailAsync` now creates a 5s timeout context, defers `cancel()` in the goroutine, and passes the bounded context into the active-admin lookup. Added a focused handler test that asserts the lookup receives a deadline and that the async path cancels the context after completion. Verified with `go test ./internal/handler ./internal/mailer ./internal/repository`, `go test ./...`, and `go build ./...`.

- [x] Add invalid admin email skip coverage.
  - Add a targeted test for an invalid, non-local active admin email (for example `not-an-email`) that confirms the mailer is not called and `reason=invalid_admin_email` is logged.
  - Confirm logs do not include verification tokens, passwords, session IDs, SMTP credentials, or raw request bodies.
  - Keep this as backend test coverage only unless implementation discovers a compile requirement.
  - Findings: added `TestSendAdminNewRegistrationEmailSkipsInvalidAdminEmail` in `backend/internal/handler/admin_notification_email_test.go`; it confirms the mailer is skipped and the safe invalid-email reason is logged. Verified with `go test ./internal/handler ./internal/mailer ./internal/repository`, `go test ./...`, and `go build ./...`.

- [x] Triage the React Router audit finding safely.
  - Run `cd frontend && npm audit --audit-level=high` and `cd frontend && npm audit` to confirm severity and affected versions.
  - If only moderate React Router advisories remain and the available fix is still a breaking v7 upgrade, document the risk/decision in `PLAN.md` and PR notes rather than bundling a major router migration into this feature PR.
  - If the audit result changes to high/critical or a non-breaking patched v6 is available, stop and update the plan before implementation.
  - Findings: both `npm audit --audit-level=high` and `npm audit` still report two moderate React Router advisories for `react-router`/`react-router-dom` 6.30.4. The only suggested fix is `react-router-dom@7.18.2` via `npm audit fix --force`, which is a breaking major upgrade. Decision: defer the router upgrade to a separate dependency/security PR instead of bundling it into this review follow-up.

- [x] Final validation and PR hygiene.
  - Run backend targeted validation and full build/tests after backend changes.
  - Run frontend audit validation if the dependency advisory is triaged/documented.
  - Update `PLAN.md` findings for each completed review item.
  - Update PR #47 description or comment with the review follow-up summary if requested by the user; do not push/create additional commits unless explicitly requested by `/commit` or PR update workflow.
  - Findings: reran `go test ./internal/handler ./internal/mailer ./internal/repository`, `go test ./...`, `go build ./...`, `npm audit --audit-level=high`, and `npm audit`. Backend tests/build passed, and the frontend audit still reports the same two moderate `react-router` advisories with only a breaking `react-router-dom@7.18.2` fix available, matching the documented decision to defer the dependency upgrade to a separate PR. PR note/comment update was not performed because it was not requested in this turn.

## Validation

- Backend validation for review fixes:
  - `cd backend && go test ./internal/handler ./internal/mailer ./internal/repository`.
  - `cd backend && go test ./...`.
  - `cd backend && go build ./...`.
- Frontend/dependency advisory validation:
  - `cd frontend && npm ci` if dependencies are not installed in the worktree.
  - `cd frontend && npm audit --audit-level=high`.
  - `cd frontend && npm audit` to capture/document remaining moderate React Router advisory details.
  - If frontend source or package files are changed unexpectedly: `cd frontend && npm run lint`, `cd frontend && npx vitest run src/components/Navbar.test.tsx`, and `cd frontend && npm run build`.
- Security/log validation:
  - Confirm admin lookup timeout/invalid-email skip logs use safe reason fields only.
  - Confirm no token/password/session/SMTP secret/raw-body leakage is introduced.
- Infra validation if manifests unexpectedly change:
  - `kubectl apply --dry-run=server -f deploy/k8s/`.
  - `kubectl diff -f deploy/k8s/`.
  - Docker/Kubernetes changes should include rollout impact and rollback before implementation proceeds.

## Risks

- Too-short admin lookup timeout could skip admin emails during transient DB pressure; registration must remain successful and logs should make the skip diagnosable.
- Too-long timeout could still accumulate background goroutines under DB stalls; keep the bound modest.
- Additional logging/tests could accidentally include email addresses or sensitive values; keep skip/error logs minimal and avoid raw payloads.
- Bundling a React Router v7 migration into this feature PR could introduce routing/auth regressions; prefer a dedicated dependency/security PR unless explicitly approved.
- Leaving moderate dependency advisories unresolved creates residual risk; document the decision and follow-up clearly.

## Rollback

- Backend review-fix rollback: revert the follow-up commit or restore the previous `admin_notification_email.go` behavior; the original feature remains usable.
- Test-only rollback: revert the added test changes if they are flaky or incorrectly specified.
- React Router advisory handling rollback:
  - If only documentation is changed, revert the plan/PR note.
  - If a dependency upgrade is explicitly approved and later causes issues, revert `frontend/package.json` and `frontend/package-lock.json` and redeploy the previous frontend image.
- No DB migration rollback is expected for these review fixes. The existing feature rollback remains: previous app images can ignore `admin_notification_state`, or the down migration can drop only that table if notification state is disposable.
- Kubernetes rollback if manifests unexpectedly change: reapply previous manifests and redeploy with `make deploy VERSION=<previous>`.

## Notes / Decisions

- Decision: address backend reliability/test coverage in small backend-only steps first.
- Decision: treat the React Router finding as dependency/security follow-up unless the user explicitly approves a breaking v7 migration in this PR.
- Decision: keep admin notification email delivery asynchronous and best-effort.
- Assumption: the review findings listed above are the current scope of “améliorations suite au retour de la review”.
- Assumption: PR #47 remains the target PR for these follow-up changes.
- Unresolved: whether the React Router moderate advisories should be accepted temporarily with a tracked follow-up, or fixed immediately in a separate dependency PR.
- No further `/next` task remains for `platform-build`; the review follow-up plan is complete.
