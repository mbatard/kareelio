# Kareelio

Job application tracker — manage your job search in one place.

## Stack

- **Backend**: Go 1.22, chi router, PostgreSQL 16, bcrypt, HttpOnly session cookies
- **Frontend**: React 18, TypeScript, Vite, TailwindCSS, i18next
- **Database**: PostgreSQL 16
- **Local deployment**: Docker Compose
- **Production deployment**: Kubernetes (manifests in `deploy/k8s/`)

## Features

- Track job applications through the full pipeline (draft → applied → interview → test → offer → decision)
- Per-user data isolation (multi-tenant by user ID, no data leakage)
- Admin role with full user management (create, disable, delete users)
- Admin dashboard with KPIs, conversion funnel, and breakdowns
- Admin audit log with actor, action, IP, timestamp, and metadata
- Admin account is protected (cannot be deleted, deactivated, or role-changed)
- Admin can reset user passwords from the edit page
- Light/dark/system theme toggle
- FR/EN/system language toggle
- Profile management (name, email, password, language, theme)
- Email validation (RFC-compliant, backend + frontend)
- Application table with sortable columns (company, title, status, location, remote, priority, date)
- Last-date column showing the most recent milestone date per application
- Application filtering by status and text search
- CSV export and import (append or replace mode) with transactional safety
- About page with version info
- Semantic Release with Conventional Commits
- Health and readiness endpoints

## Job Application Fields

| Field | Description |
|-------|-------------|
| Company | Company name |
| Title | Job title |
| Status | Draft, Applied, Responded, Interview, Test, Offer, Rejected, Withdrawn |
| Salary | Min/max, currency |
| Contract type | CDI, CDD, Freelance, Internship, Apprentice, Other |
| Location | City / area |
| Remote | On site, Hybrid, Full remote |
| Benefits | Perks and benefits |
| Announcement URL | Link to the job posting |
| Applied at | Application date |
| Response received | Yes/No + date |
| First contact | Date + Video/Phone/In Person |
| Test | Yes/No + date + notes |
| Offer | Yes/No + date + amount |
| Priority | Low, Medium, High |
| Source | LinkedIn, Welcome to the Jungle (WTTJ), Indeed, Referral, Agency, Website, Other |
| Recruiter contact | Contact details |
| Notes | Free-form notes |

## Quick Start (Docker Compose)

```bash
# Clone and configure
cp .env.example .env
# Edit .env with your values (at minimum, change passwords)

# Start everything
make dev

# Or manually
docker compose up --build
```

- Frontend: http://localhost:5173
- Backend API: http://localhost:8080
- PostgreSQL: localhost:5432

Default admin credentials:
- Email: `admin@kareelio.local`
- Password: `admin`

> **Change the default password immediately in production.**

## Development

### Prerequisites

- Go 1.22+
- Node.js 20+
- PostgreSQL 16+ (or use Docker)

### Backend

```bash
cd backend
go mod download
go run ./cmd/server
```

### Frontend

```bash
cd frontend
npm install
npm run dev
```

### Tests

```bash
make test            # Run all tests
make test-backend    # Backend only
make test-frontend   # Frontend only
```

### Lint

```bash
make lint
```

## API Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | /api/healthz | No | Health check |
| GET | /api/readyz | No | Readiness check |
| POST | /api/auth/login | No | Login |
| POST | /api/auth/logout | Yes | Logout |
| GET | /api/auth/me | Yes | Current user |
| GET | /api/profile | Yes | Get profile |
| PUT | /api/profile | Yes | Update profile |
| PUT | /api/profile/password | Yes | Change password |
| GET | /api/job-applications | Yes | List applications |
| POST | /api/job-applications | Yes | Create application |
| GET | /api/job-applications/:id | Yes | Get application |
| PUT | /api/job-applications/:id | Yes | Update application |
| DELETE | /api/job-applications/:id | Yes | Delete application |
| GET | /api/job-applications/export | Yes | Export applications as CSV |
| POST | /api/job-applications/import | Yes | Import applications from CSV |
| GET | /api/about | Yes | Version info |
| GET | /api/users | Admin | List users |
| POST | /api/users | Admin | Create user |
| GET | /api/users/:id | Admin | Get user |
| PUT | /api/users/:id | Admin | Update user |
| DELETE | /api/users/:id | Admin | Delete user |
| PUT | /api/users/:id/password | Admin | Reset user password |
| GET | /api/admin/dashboard | Admin | Dashboard stats |
| GET | /api/admin/audit | Admin | Audit log |

## Kubernetes Deployment

Manifests are in `deploy/k8s/`. Security-hardened with `runAsNonRoot`, `readOnlyRootFilesystem`, `drop ALL` capabilities, `automountServiceAccountToken: false`, CiliumNetworkPolicy (deny-all + explicit allow), and Traefik ingress with HSTS + rate limiting via Middleware CRDs.

```bash
cd deploy/k8s
cp secret.example.yaml secret.yaml
# Edit secret.yaml with production values

kubectl apply -f namespace.yaml
kubectl apply -f configmap.yaml
kubectl apply -f secret.yaml
kubectl apply -f networkpolicy.yaml
kubectl apply -f traefik-middlewares.yaml
kubectl apply -f postgres-statefulset.yaml
kubectl apply -f postgres-service.yaml
kubectl apply -f backend-deployment.yaml
kubectl apply -f backend-service.yaml
kubectl apply -f frontend-deployment.yaml
kubectl apply -f frontend-service.yaml
kubectl apply -f ingress.yaml
```

## Project Structure

```
kareelio/
├── backend/                    # Go API
│   ├── cmd/server/             # Entry point
│   ├── internal/
│   │   ├── config/             # Environment configuration
│   │   ├── database/           # PostgreSQL connection + migrations
│   │   ├── handler/            # HTTP handlers
│   │   ├── middleware/          # Auth, CORS, security, logging, audit
│   │   ├── model/              # Data types
│   │   ├── repository/         # Database queries
│   │   ├── router/             # Route definitions
│   │   └── validation/         # Email and input validation
│   ├── migrations/             # SQL migrations
│   ├── test/                   # Tests
│   ├── Dockerfile
│   └── go.mod
├── frontend/                   # React + TypeScript
│   ├── src/
│   │   ├── components/         # Shared components (KpiCard, MapCard, Navbar)
│   │   ├── contexts/           # Auth, Theme, Language
│   │   ├── constants/          # Shared constants (statusColors)
│   │   ├── i18n/               # Translations (FR, EN)
│   │   ├── pages/              # Page components
│   │   ├── services/           # API client
│   │   ├── types/              # TypeScript types
│   │   └── utils/              # Utility functions (email validation)
│   ├── Dockerfile
│   └── package.json
├── deploy/k8s/                 # Kubernetes manifests
├── docker-compose.yml          # Local development
├── Makefile                    # Common commands
├── package.json                # Semantic Release
└── .releaserc.json
```

## Security

### OWASP Top 10 Coverage

| Risk | Mitigation |
|------|------------|
| Broken Access Control | Data isolation by `owner_user_id`, RBAC, admin protections |
| Cryptographic Failures | bcrypt passwords, HttpOnly/SameSite session cookies |
| Injection | Parameterized queries (pgx), input validation |
| Insecure Design | CSRF protection, rate limiting, security headers |
| Security Misconfiguration | K8s security contexts, CiliumNetworkPolicy, Traefik hardening |
| Vulnerable Components | Pinned base images, no `latest` tags in K8s |
| Auth Failures | Rate limiting on login (10 req/min per IP), session timeout |
| Data Integrity Failures | CSV formula injection protection, strict import validation |
| Logging Failures | Comprehensive audit log with actor, IP, action, metadata |
| SSRF | Origin/Referer validation on state-changing requests |

### Specific Measures

- **Passwords**: bcrypt with default cost
- **Sessions**: HttpOnly, SameSite=Lax, configurable Secure flag, configurable duration
- **Rate limiting**: 10 requests/min per IP on `/api/auth/login`
- **CSRF**: Origin/Referer check on POST/PUT/PATCH/DELETE
- **CSV import**: Formula injection protection (prefix `=`, `+`, `-`, `@`), strict enum validation, 1000-row limit
- **CSV export**: All fields sanitized against formula injection
- **Password reset** (admin): revokes all sessions for the target user
- **Security headers**: CSP, HSTS, X-Content-Type-Options, X-Frame-Options, Referrer-Policy, Permissions-Policy
- **K8s**: `runAsNonRoot`, `readOnlyRootFilesystem`, `drop ALL` capabilities, `automountServiceAccountToken: false`
- **K8s CiliumNetworkPolicy**: default deny-all, explicit allow for Traefik→frontend→backend→postgres
- **K8s Ingress**: Traefik with HSTS, rate limiting, security headers via Middleware CRDs
- **Image pinning**: specific version tags (no `latest` in K8s manifests)
- **Panic recovery**: middleware catches panics, returns 500
- **Request timeouts**: 30s timeout per request
- **Data isolation**: all job application queries filter by `owner_user_id` (never just `id`)
- **Admin protections**: cannot be deleted, deactivated, or role-changed; email locked

## Release

This project uses [Semantic Release](https://github.com/semantic-release/semantic-release).

Commits must follow [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: add job application filtering
fix: resolve data leakage between users
chore(release): 1.0.0
```

## Development Note

Kareelio was built with AI-assisted development using [OpenCode](https://opencode.ai).

The codebase has been reviewed, tested, and curated by the project maintainer.
AI assistance was used as a development accelerator, but the project direction,
feature decisions, validation, and final responsibility remain with the maintainer.

## License

[GNU Affero General Public License v3.0](LICENSE) - Copyright (c) 2026 Mikael Batard
