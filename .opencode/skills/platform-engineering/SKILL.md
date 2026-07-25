---
name: platform-engineering
description: Use for Platform Engineer/SRE tasks involving Go, TypeScript, Docker, Kubernetes, Helm, Terraform, GitHub Actions, CI/CD, production deployment, reliability, or security hardening.
---

# Platform Engineering / SRE Skill

Use this skill for Kareelio changes touching backend Go, frontend TypeScript, Docker, Kubernetes, Helm, Terraform, GitHub Actions, CI/CD, dependency security, production deployments, or reliability.

## Operating Model

- Use `PLAN.md` for non-trivial changes.
- Prefer small, reversible changes.
- Validate each step before moving on.
- Never bypass protected `main`; use feature branch + PR.
- Never expose secrets in files, logs, workflow output, PR comments, or chat.

## Go Backend

- Run targeted tests when possible, then `go test ./...` for broader validation.
- Preserve context propagation and error wrapping.
- Treat database nullable columns carefully with pgx scans.
- Keep migrations idempotent unless a migration tracking system exists.
- Do not weaken auth, CSRF, cookies, or audit logging.

## TypeScript / React

- Run `npm run lint`, `npm run build`, and `npm audit --audit-level=high` when frontend dependencies or code change.
- Keep i18n updates in both `fr.json` and `en.json`.
- Treat routing and redirects as security-sensitive.

## Docker

- Avoid `latest` tags in deploy manifests.
- Prefer small, patched base images.
- Keep runtime containers non-root when possible.

## Kubernetes / Helm

- Validate with `kubectl diff`, `kubectl apply --dry-run=server`, `helm lint`, or `helm template` when applicable.
- Include rollout and rollback commands.
- Preserve probes, resource requests/limits, securityContext, and NetworkPolicy.
- Do not apply destructive commands without explicit user approval.

## Terraform

- Always run `terraform fmt`, `terraform validate`, and `terraform plan` before apply.
- Never run `terraform destroy` without explicit approval.
- Keep secrets in provider secret stores, not `.tfvars` committed files.

## GitHub Actions Security

- Avoid `pull_request_target` unless explicitly reviewed.
- PR workflows must not access secrets.
- Use minimum permissions per workflow/job.
- Publishing packages/images/releases should happen only on trusted events such as push to `main` or tags.
- Consider pinning third-party actions by full SHA for high-security workflows.
