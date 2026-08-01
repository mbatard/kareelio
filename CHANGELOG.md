# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Safe backend startup diagnostics for job application encryption configuration.
- Kubernetes secret example entries for job application encryption keys.
- Admin users list action to resend verification email for unverified non-admin users.

### Changed

- Backend request logs now skip `/api/healthz` and `/api/readyz` probe traffic.
- Local Docker Compose examples now pass job application encryption flags to the backend.

## [1.2.0] - 2026-08-01

### Added

- Email verification operations for public registration and resend flows.
- Admin-only endpoint and UI action to force-resend verification emails for unverified non-admin users.
- Safe backend action logs for registration, public resend, admin resend, and mail delivery paths.
- Safe frontend API lifecycle logs gated by development mode or `VITE_DEBUG_LOGS=true`.
- SMTP startup/config summary and phase-specific SMTP send diagnostics.
- SMTP timeout configuration with `SMTP_TIMEOUT_SECONDS`.

### Changed

- SMTP examples and production configuration now default to unauthenticated relay port `25`.
- Public registration and resend responses remain generic where needed to avoid account enumeration while internal failures are logged server-side.

## [1.1.0] - 2026-08-01

### Added

- Encrypted storage support for sensitive job application fields.
- Job application encryption backfill with dry-run support.
- Encrypted-read enforcement flag for production rollout.
- Kubernetes configuration for requiring encrypted job application reads.

### Changed

- Job application reads can decrypt encrypted fields and optionally reject legacy plaintext rows until backfill is complete.

## [1.0.2] - 2026-07-28

### Fixed

- Aligned job application status badge colors.

## [1.0.1] - 2026-07-28

### Fixed

- Improved applications mobile layout.

## [1.0.0] - 2026-07-28

### Added

- Semantic Release workflow for automated tags, GitHub releases, and release notes.

### Changed

- Backend Go toolchain upgraded to Go 1.26.

## [0.1.0] - 2026-07-14

### Added

- Initial project structure
- Backend API with Go (chi router, PostgreSQL)
- Authentication with HttpOnly session cookies
- User management (admin + user roles)
- Job application tracking with full CRUD
- Light/dark/system theme toggle
- FR/EN/system language toggle
- User profile management
- Docker Compose local deployment
- Kubernetes deployment manifests
- Semantic Release configuration
- Healthcheck endpoints (`/api/healthz`, `/api/readyz`)
- Admin dashboard with KPIs, conversion funnel, and breakdowns (status, source, remote, priority)
- Admin user management with create/edit/delete, description field, email validation
- Admin audit log with pagination (actor, action, IP, timestamp, metadata)
- Welcome to the Jungle (WTTJ) as job source
- RFC-compliant email validation (backend + frontend mirror)
- Sortable columns in applications table (company, title, status, location, remote, priority, date)
- Last-date column showing the most recent milestone date per application
- Audit logging for login/logout, user CRUD, profile changes, password changes, job application CRUD
- Shared components: KpiCard, MapCard, statusColors
- Date localization using `toLocaleDateString()`
- Admin can reset user password from edit page (with confirmation)
- CSV export and import for job applications (append or replace mode)
- Transactional CSV import with validation and error reporting
- Audit logging for export and import events

## [0.2.0] - 2026-07-19

### Added

- Session cookie configuration from env (`SESSION_COOKIE_SECURE`, `SESSION_COOKIE_SAMESITE`, `SESSION_DURATION_HOURS`)
- Rate limiting on `/api/auth/login` (10 requests per minute per IP)
- CSRF protection middleware (Origin/Referer validation on state-changing methods)
- CSV formula injection protection (prefix `=`, `+`, `-`, `@` fields on export, strict validation on import)
- CSV import strict enum validation (status, remote, contract type, priority, source, contact type)
- CSV import row limit (1000 max)
- `ReplaceAll` transactional method for CSV import replace mode
- Session revocation on admin password reset (all sessions for the user are deleted)
- Kubernetes security contexts (`runAsNonRoot`, `readOnlyRootFilesystem`, `drop ALL` capabilities, `automountServiceAccountToken: false`)
- CiliumNetworkPolicy (default deny-all, explicit allow for Traefik→frontend→backend→postgres)
- Traefik IngressRoute with HSTS, rate limiting, security headers via Middleware CRDs
- Traefik Middleware manifests (`security-headers`, `rate-limit`)
- Kubernetes image pinning (version tags, no `latest`)
- Nginx security headers for local dev (CSP, HSTS, Permissions-Policy, dotfile blocking)
- GitHub Actions CI (Go tests + frontend build + npm audit)
- GitHub Actions Docker build + push to GHCR (`ghcr.io/mbatard/kareelio-backend`, `ghcr.io/mbatard/kareelio-frontend`)
- GitHub CodeQL analysis (Go + JavaScript/TypeScript)
- Dependabot configuration (Go modules, npm, Docker, GitHub Actions)

### Changed

- `router.New()` now accepts `*config.Config` for cookie/security configuration
- `UserHandler` now accepts `*repository.SessionRepository` for session revocation on password reset
- CSV import uses strict CSV parsing (`LazyQuotes` disabled)
- CSV import returns detailed row-level validation errors
- CSRF protection now fails closed (rejects requests without Origin or Referer)
- `UpdateUserRequest` no longer accepts `role` field (role escalation prevented)
- K8s images pinned to `ghcr.io/mbatard/kareelio-*` instead of `kareelio/*`
- Session cookie `Secure` is now always `true` (removed `SESSION_COOKIE_SECURE` env var)
- Logout cookie now also includes `Secure: true` and `SameSite: Strict`

### Removed

- `SessionCookieSecure` from `Config` struct (always `true`)
- `SESSION_COOKIE_SECURE` env var from `.env.example` and K8s ConfigMap
