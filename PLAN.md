# PLAN.md

## Objective

Track the current incremental development objective for Kareelio.

## Constraints

- Protected `main`: feature branch + PR only.
- No direct pushes to `main`.
- No secrets in code, logs, PRs, or committed files.
- Prefer small, reversible changes.
- Validate each completed step before moving to the next.
- For Kubernetes, Helm, Terraform, Docker, and GitHub Actions changes, include rollout impact and rollback.

## Current State

- Project stack: Go backend, React/TypeScript frontend, PostgreSQL, Docker Compose, Kubernetes manifests.
- Production target: `app.kareelio.fr` on Kubernetes.
- Current workflow: protected `main`, feature branch, PR, required checks.

## Tasks

- [ ] Define the next objective with `/plan <objective>`.

## Validation

- Backend: `cd backend && go test ./...`, `cd backend && go build ./...`.
- Frontend: `cd frontend && npm run lint`, `cd frontend && npm run build`, `cd frontend && npm audit --audit-level=high`.
- Docker: build or compose checks when Dockerfiles or runtime packaging change.
- Kubernetes/Helm: dry-run/template/diff before apply.
- Terraform: `terraform fmt`, `terraform validate`, `terraform plan` before apply.

## Risks

- Dependency updates can introduce breaking changes.
- CI/CD or GitHub Actions changes can weaken PR isolation if permissions or events are wrong.
- K8s/Terraform changes can affect production availability if rollback is not documented.

## Rollback

- Code: revert PR or commit.
- Docker/K8s: redeploy previous image tag with `make deploy VERSION=<previous>`.
- Terraform: use the previous reviewed plan/state-backed rollback strategy.

## Notes / Decisions

- Use `/plan` for non-trivial work.
- Use `/next` for one incremental implementation step.
- Use `/review` before commit/PR for risky changes.
