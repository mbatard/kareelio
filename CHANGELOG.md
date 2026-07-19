# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
- Traefik ingress with HSTS, rate limiting, security headers via Middleware CRDs
- Traefik Middleware manifests (`security-headers`, `rate-limit`)
- Kubernetes image pinning (version tags, no `latest`)
- Nginx security headers for local dev (CSP, HSTS, Permissions-Policy, dotfile blocking)

### Changed

- `router.New()` now accepts `*config.Config` for cookie/security configuration
- `UserHandler` now accepts `*repository.SessionRepository` for session revocation on password reset
- CSV import uses strict CSV parsing (`LazyQuotes` disabled)
- CSV import returns detailed row-level validation errors
