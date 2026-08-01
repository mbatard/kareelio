# PLAN.md

## Objective

Harden Kareelio database data security so that direct PostgreSQL reads, dumps, or backups do not expose user job application content in plaintext, while preserving the protected-main workflow and keeping changes incremental and reversible.

## Constraints

- Protected `main`: feature branch + PR only.
- No direct pushes to `main`.
- No secrets in code, logs, PRs, committed manifests, or workflow output.
- Keep production changes small, reversible, and validated before rollout.
- Treat database migrations as sensitive and keep them idempotent where possible because the current migration runner executes `.up.sql` files at startup.
- Do not run destructive Kubernetes or database operations against production without explicit approval and a verified backup.
- Application-layer encryption protects against direct DB/PVC/dump reads, but it does **not** protect against an attacker/operator who can read both the database and the application data-encryption key.
- K8s/Terraform/storage changes must include dry-run/plan, rollout impact, and rollback.

## Current State

- Stack: Go backend, React/TypeScript frontend, PostgreSQL, Docker Compose, Kubernetes manifests.
- PostgreSQL runs in-cluster as a `StatefulSet` with `postgres:16.4-alpine` and a PVC.
- `CiliumNetworkPolicy` already denies all by default and allows PostgreSQL ingress only from pods labeled `app: backend`.
- Backend connects with `POSTGRES_USER` / `POSTGRES_PASSWORD` and `sslmode=disable`.
- The same database user is currently used for application queries, migration execution, and admin seed logic.
- `DB_MIGRATE` defaults to `true`; production config currently has `DB_MIGRATE: "true"`.
- `job_applications` stores sensitive user data in plaintext: company, title, salary, location, benefits, announcement URL, contact notes, test notes, recruiter contact, free-form notes, offer amount.
- Admin UI intentionally shows only aggregate statistics, but direct DB access can still reveal all application data.
- Admin dashboard queries aggregate data from `job_applications`; those aggregates can continue to use non-sensitive plaintext fields such as status/source/remote/priority/dates.
- Kubernetes `Secret` example stores DB credentials and app secrets as normal K8s secrets; no external secret manager, SealedSecrets, SOPS, or KMS integration is currently configured in repo.
- There is no application-layer encryption for candidate/application data today.

## Tasks

- [x] Confirm threat model and data classification.
  - Decision for first implementation: protect against accidental DB dumps/direct SQL reads, compromised read-only DB credentials, and database backups/PVC snapshots.
  - Out of scope for this phase: malicious cluster admins or any attacker who has both the database and `DATA_ENCRYPTION_KEY`.
  - Encrypted fields: company, title, salary, location, benefits, announcement URL, contact notes, test notes, recruiter contact, free-form notes, offer amount.
  - Plaintext metadata: `owner_user_id`, `status`, `source`, `remote`, `priority`, boolean flags, dates, `created_at`, `updated_at`.

- [x] Design application-layer encryption format and key management.
  - Added `DATA_ENCRYPTION_KEY` / `DATA_ENCRYPTION_KEY_ID` to backend config loading so the values can come from environment/K8s secret.
  - Implemented AES-256-GCM envelope encryption in `backend/internal/encryption` with versioned format `v1:<key_id>:<base64 nonce+ciphertext>`.
  - Added unit tests for round-trip, corrupt ciphertext, wrong-key, invalid envelope, and empty-string handling.
  - Key material is expected as a base64-encoded 32-byte value (e.g. `openssl rand -base64 32`) stored outside the repo in production secret management/K8s secret.

- [x] Add an expand migration for encrypted columns.
  - Added nullable encrypted columns for the sensitive job application fields, including `company_enc`, `title_enc`, `salary_min_enc`, `salary_max_enc`, `salary_currency_enc`, `location_enc`, `benefits_enc`, `announcement_url_enc`, `test_notes_enc`, `offer_amount_enc`, `recruiter_contact_enc`, and `notes_enc`.
  - Kept existing plaintext columns intact for compatibility and rollback.
  - Avoided dropping or scrubbing plaintext in the first migration.
  - Validation: ran backend tests and a local PostgreSQL startup through `docker compose` with startup migrations; migration `008_add_encrypted_job_application_columns.up.sql` executed successfully.

- [x] Update backend repository/model logic for dual-read and encrypted writes.
  - On read: prefer encrypted columns when present; fall back to legacy plaintext for rows not yet backfilled.
  - On create/update/import: write encrypted columns while keeping plaintext populated for compatibility during the rollout.
  - CSV export now returns decrypted plaintext to the authenticated owner via the repository read path.
  - Added repository tests covering encrypted write payloads, encrypted read preference, plaintext fallback, and update-field merging.

- [x] Add a controlled backfill path.
  - Added a startup-safe one-shot job controlled by `JOB_APPLICATIONS_BACKFILL=true` and `JOB_APPLICATIONS_BACKFILL_DRY_RUN=true|false`.
  - The backfill is resumable/idempotent: it targets only rows where encrypted fields are still missing and preserves any already populated `_enc` values.
  - Dry-run/count mode logs only candidate counts; runtime logs never print plaintext or ciphertext values.
  - Validation: ran backend tests/build and executed the backfill against a local PostgreSQL copy with a legacy plaintext row; dry-run reported 1 candidate row, the real run updated 1 row, and the candidate count dropped to 0.

- [x] Switch to encrypted-read enforcement.
  - Added `JOB_APPLICATIONS_REQUIRE_ENCRYPTED_READS=true` for production config and a startup check that fails fast if enforcement is enabled without `DATA_ENCRYPTION_KEY`.
  - Legacy plaintext rows now fail closed when enforcement is enabled; compatibility mode still exists for temporary migration use.
  - Legacy reads in compatibility mode increment a counter and log a warning with the row ID for operator visibility.

- [ ] Scrub or remove plaintext sensitive columns in a contract migration.
  - Only after a successful encrypted rollout and verified backup, replace legacy plaintext sensitive columns with neutral placeholders, nullable defaults, or drop them if code no longer needs them.
  - Preserve non-sensitive aggregate columns required for admin dashboard and filtering.
  - This is the irreversible/highest-risk step and requires explicit production approval.
  - Blocked: no explicit production approval/verified backup has been provided yet; keep plaintext columns until approval is granted.

- [ ] Split database roles and reduce DB privileges.
  - Introduce separate roles/secrets: app runtime role (`kareelio_app`) and migration/admin role (`kareelio_migrator`).
  - App role should have only required DML privileges, no schema ownership, no broad DDL.
  - Migration role handles schema changes in controlled jobs/operations.
  - Change production `DB_MIGRATE` default/path so backend pods do not routinely run DDL with runtime credentials.
  - Add privilege verification SQL/tests, e.g. app role cannot `ALTER TABLE`, cannot access unrelated schemas, can only read/write expected tables.

- [ ] Consider PostgreSQL Row Level Security as defense-in-depth.
  - Evaluate RLS on `job_applications` keyed by `owner_user_id` and a transaction/session setting for authenticated user ID.
  - RLS helps reduce blast radius of SQL injection or application bugs, but it does not solve direct superuser/owner access or plaintext dumps by itself.
  - Implement only after role split is clear.

- [ ] Harden Kubernetes and operational access.
  - Confirm PostgreSQL Service is cluster-internal only and not exposed by Ingress/LoadBalancer.
  - Keep/validate NetworkPolicy default deny and backend-only PostgreSQL ingress.
  - Review Kubernetes RBAC outside the app manifests: restrict who can read `kareelio-secret`, exec into backend/postgres pods, port-forward PostgreSQL, or read PVC snapshots.
  - Document production secret handling using External Secrets, SealedSecrets, SOPS, or equivalent; do not commit real secrets.
  - Confirm storage encryption at rest at the infrastructure/storage-class level.

- [ ] Add backup, restore, and key-rotation procedures.
  - Backups after encryption must include ciphertext and must be useless without the data-encryption key.
  - Store encryption keys separately from DB backups.
  - Document restore test: restore DB + key into staging and verify app can decrypt.
  - Plan key rotation with `DATA_ENCRYPTION_KEY_ID`, multiple accepted keys, and re-encryption job.

- [ ] Document security limitations and operator guidance.
  - Document that application-layer encryption protects DB-only access, not compromise of both DB and application secrets.
  - Document emergency access procedure, audit expectations, and incident response for suspected DB dump/key exposure.

## Validation

- Backend targeted crypto tests: `cd backend && go test ./internal/crypto/...` or package-specific equivalent once created.
- Backend full tests: `cd backend && go test ./...`.
- Backend build: `cd backend && go build ./...`.
- Frontend when API shape changes: `cd frontend && npm run lint`, `cd frontend && npm run build`.
- Migration checks:
  - Run against an empty local DB.
  - Run against a local/staging copy with existing plaintext rows.
  - Run twice to confirm idempotence where applicable.
  - Verify backfill does not log sensitive values.
- Database privilege checks:
  - App role can perform normal app CRUD.
  - App role cannot run DDL such as `ALTER TABLE` or read secrets outside its required tables.
  - Migration role can run migrations only in controlled context.
- Kubernetes checks:
  - `kubectl apply --dry-run=server -f deploy/k8s/` before apply.
  - `kubectl diff -f deploy/k8s/` where available.
  - Verify NetworkPolicy still allows frontend→backend and backend→postgres only as intended.
- Production rollout checks:
  - Take/verify backup before encryption/backfill/scrub steps.
  - Deploy expand migration + dual-read/write first.
  - Backfill with dry-run/count, then execute.
  - Verify representative users can list/edit/export applications.
  - Only then consider plaintext scrub/drop.

## Risks

- Losing `DATA_ENCRYPTION_KEY` makes encrypted application data unrecoverable.
- A bad encryption rollout can make existing applications unreadable if fallback/backfill is wrong.
- Direct DB operators with access to both DB and application secrets can still decrypt data.
- Encrypting searchable fields can reduce DB-side search/filter capability; current search is frontend-side after list, but future DB search may need blind indexes or dedicated search design.
- Scrubbing/dropping plaintext columns is destructive and requires explicit approval and verified backups.
- Splitting DB roles changes migration/deployment behavior and can cause startup failures if credentials or privileges are wrong.
- RLS can break queries if session context is missing or migrations/admin queries are not adapted.

## Rollback

- Before plaintext scrub/drop:
  - Revert application PR to plaintext/fallback behavior.
  - Keep plaintext columns populated during the initial dual-write phase.
  - Disable backfill/encryption enforcement flag if introduced.
- After encrypted columns are added:
  - Safe rollback is possible while legacy plaintext columns remain intact.
- After plaintext scrub/drop:
  - Rollback requires restoring a verified DB backup plus the correct encryption key and/or reverting to an earlier schema backup.
  - Previous application image alone is not sufficient.
- Kubernetes/role changes:
  - Reapply previous manifests/secrets if role split breaks connectivity.
  - Keep old DB credentials until the new app/migrator role rollout is verified, then rotate/remove.

## Notes / Decisions

- Recommended first implementation is application-layer envelope encryption in the Go backend, not PostgreSQL `pgcrypto`, because DB-side encryption still exposes plaintext to direct SQL users who can call decrypt functions or access keys if stored in DB.
- Keep admin dashboard aggregates based on non-sensitive metadata so admin UI remains useful without decrypting user content.
- Do not commit actual encryption keys, DB passwords, or production secret material.
- Use feature branch + PR for every increment.
- First `/next` task for `platform-build`: implement backend encryption primitives and config only, with tests, without changing database schema yet.
