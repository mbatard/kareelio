# AGENTS.md — Conventions du projet Kareelio

## Git Workflow

- **Jamais de push direct sur `main`** — toujours via feature branch + PR
- Branch naming : `feat/...`, `fix/...`, `chore/...`
- Squash merge sur main (historique propre)
- Supprimer la branch après merge

## Commit Convention (Conventional Commits)

- Format : `<type>(<scope>): <description>`
- Types : `feat`, `fix`, `chore`, `refactor`, `test`, `docs`, `ci`
- Scope : `auth`, `admin`, `profile`, `jobs`, `k8s`, `docker`, `migration`, `i18n`
- Pas de `.` à la fin du titre
- Exemples :
  - `feat(auth): add email verification`
  - `fix(migration): make 007 idempotent`
  - `chore(deps): bump vite to v7`

## Build & Test

- `make dev` — lance Docker Compose (Postgres + backend + frontend)
- `make test` — go test + vitest
- `make lint` — go vet + eslint
- `make build` — go build + vite build
- `make deploy VERSION=x.y.z` — déploie sur K8s
- **Avant de créer une PR** : `make lint && make test` doit passer

## Sécurité (règles critiques)

- Passwords : 12+ caractères, majuscule, minuscule, chiffre, spécial
- Cookies : Secure, HttpOnly, SameSite
- Secrets : jamais en dur, toujours via env vars ou K8s secrets
- CSRF : fail-closed sur POST/PUT/PATCH/DELETE sans Origin/Referer
- Dependabot : reviewer et merger les PRs régulièrement

## Releases

- semantic-release sur merge to main → bump version + CHANGELOG + Docker tags
- Tags : `v1.2.3`
- Images Docker : `ghcr.io/mbatard/kareelio-backend`, `ghcr.io/mbatard/kareelio-frontend`

## Déploiement

### Local
```bash
make dev          # Docker Compose up --build
make stop         # Docker Compose down
make dev-d        # Detached mode
make logs         # Follow logs
```

### Production (K8s)
```bash
# 1. Créer et merger la PR sur main
# 2. semantic-release crée le tag + push les images Docker
# 3. Déployer :
make deploy VERSION=x.y.z

# 4. Vérifier :
make deploy-status
make deploy-logs
```

## Structure du Projet

```
backend/          Go 1.22, chi router, pgx, bcrypt
  cmd/server/     Point d'entrée
  internal/       Handlers, middleware, models, repository, validation, config, mailer
  migrations/     SQL migrations (001-xxx.up.sql / .down.sql)
frontend/         React 18, TypeScript, Vite, Tailwind, i18next
  src/pages/      Pages React
  src/components/ Composants réutilisables
  src/i18n/       Traductions fr.json, en.json
  src/services/   API client (axios)
deploy/k8s/       Manifests Kubernetes
marketing/        Screenshots Playwright + seed SQL
```

## Environments

| Env | URL | Stack |
|---|---|---|
| Local | localhost:5173 | Docker Compose |
| Production | app.kareelio.fr | K8s (Traefik, CiliumNetworkPolicy) |
