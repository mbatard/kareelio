# PLAN.md

## Objective

Résoudre les problèmes identifiés lors de la review du durcissement des conteneurs Kareelio avant PR/rollout : aligner la NetworkPolicy frontend avec le port runtime 8080, rétablir une validation `npm audit --audit-level=high` verte, réduire les permissions GitHub Actions sur les PR, et remettre le plan/branch de hardening en cohérence avec `origin/main`.

## Constraints

- Protected `main`: feature branch + PR only; no direct pushes to `main`.
- Pendant cette étape de planification, ne modifier que `PLAN.md`.
- Le worktree principal `/Users/mikael/kareelio` est sale sur `fix/applications-status-colors`; ne pas y implémenter les corrections fonctionnelles.
- Implémenter depuis un worktree/branch propre basé sur `origin/main`, ou rafraîchir `/tmp/opencode/kareelio-runtime-harden-baseline` avant toute correction applicative.
- Garder les corrections petites, réversibles, et validées individuellement.
- Ne pas introduire ni exposer de secrets; GitHub Actions PR ne doit pas accéder aux secrets.
- GitHub Actions: ne pas utiliser `pull_request_target`; permissions minimales; publication d’images uniquement sur événements de confiance (`push` main/tags).
- Kubernetes: préserver `runAsNonRoot`, `allowPrivilegeEscalation: false`, `capabilities.drop: [ALL]`, `seccompProfile: RuntimeDefault`, resource requests/limits, probes et CiliumNetworkPolicy; aucun `kubectl apply` destructif ni rollout production sans validation explicite.
- Frontend dependencies: éviter une migration majeure React Router dans la même étape sauf si nécessaire pour passer les checks; documenter toute vulnérabilité moderate résiduelle et sa décision.

## Current State

- Repository analysé: `/Users/mikael/kareelio`.
- Worktree principal actuel: branche `fix/applications-status-colors`, en avance de 1 commit, avec de nombreuses modifications non liées; il sert à porter ce `PLAN.md`, pas à implémenter les corrections de hardening.
- Worktree de hardening existant: `/tmp/opencode/kareelio-runtime-harden-baseline`, branche `chore/harden-runtime-containers`, tracking `origin/main`, maintenant rafraîchi sur `origin/main` avec le diff de hardening réappliqué.
- Le worktree de hardening contient les changements de durcissement déjà revus: backend `scratch`, frontend `scratch` + `frontend/server.go`, compose hardening, frontend Deployment/Service sur port 8080, workflow Docker avec check no-shell.
- Problèmes confirmés par review/analyse:
  - `deploy/k8s/frontend-deployment.yaml` expose le frontend sur `containerPort: 8080` et `deploy/k8s/frontend-service.yaml` garde `port: 80` avec `targetPort: 8080`.
  - `deploy/k8s/networkpolicy.yaml` autorise encore Traefik vers les pods frontend uniquement sur le port `80`; avec Cilium, cela risque de bloquer le trafic Traefik -> frontend après rollout.
  - `npm audit --audit-level=high` échoue dans `frontend/` à cause de `brace-expansion` high severity; il signale aussi des advisories moderate sur `react-router` / `react-router-dom`.
  - `.github/workflows/ci.yml` exécute `npm audit --audit-level=high`, donc la PR échouera tant que la vulnérabilité high reste présente.
  - `.github/workflows/docker.yml` donne actuellement `packages: write` au niveau workflow, y compris pour l’événement `pull_request`; les étapes de push/login sont conditionnées, mais les permissions ne sont pas minimales sur PR.
  - Le `PLAN.md` dans `/tmp/opencode/kareelio-runtime-harden-baseline` décrit encore un objectif SMTP/email-verification et ne correspond pas au travail de hardening; le plan source de vérité est ce fichier.
- Validations déjà connues du hardening précédent:
  - Images backend/frontend sans `/bin/sh` confirmées localement.
  - `make lint`, `make build`, `go test ./...`, `go build ./...`, `docker compose config`, et smoke tests Compose ont passé.
  - `make test` reste bloqué dans cet environnement par l’absence de compilateur C requis pour `go test -race` (`aarch64-linux-gnu-gcc`).
  - `kubectl` n’est pas installé ici; les validations server-side Kubernetes devront être exécutées dans un environnement cluster-capable.

## Tasks

- [x] Rafraîchir le worktree/branch de hardening et synchroniser le plan.
  - Dans `/tmp/opencode/kareelio-runtime-harden-baseline`, vérifier `git status -sb`, sauvegarder/inspecter le diff courant, puis rafraîchir proprement contre le dernier `origin/main` sans perdre les changements de hardening.
  - Si le rebase est risqué à cause du retard de 6 commits, créer un nouveau worktree/branch propre depuis `origin/main` et y réappliquer uniquement les changements de hardening nécessaires.
  - Copier ou mettre à jour `PLAN.md` dans le branch de hardening pour qu’il corresponde à ce plan de remediation.
  - Validation: `git status -sb`, `git diff --check`, et revue que le branch contient le bon `PLAN.md` et aucune modification non voulue.
  - Findings: le worktree de hardening a été fast-forwardé sur le dernier `origin/main`, le diff de hardening a été stashed puis réappliqué, et `PLAN.md` a été resynchronisé avec ce plan de remediation. Vérifié avec `git status -sb` et `git diff --check`; aucun changement applicatif supplémentaire n’a été introduit pendant le rafraîchissement.

- [x] Corriger la CiliumNetworkPolicy frontend pour le port runtime 8080.
  - Modifier seulement `deploy/k8s/networkpolicy.yaml` pour autoriser Traefik vers les pods frontend sur le port `8080` au lieu de `80`.
  - Ne pas changer `frontend-service.yaml` `port: 80`; le Service peut rester exposé sur 80 tant que `targetPort: 8080` pointe vers le pod.
  - Vérifier que `frontend-deployment.yaml`, `frontend-service.yaml`, `ingress.yaml`, et `networkpolicy.yaml` restent cohérents: Traefik route vers le Service port 80, Service target vers pod 8080, NetworkPolicy autorise pod 8080.
  - Validation: `git diff --check`; `kubectl apply --dry-run=server -f deploy/k8s/` et `kubectl diff -f deploy/k8s/` dans un environnement où `kubectl` est disponible. Si `kubectl` reste indisponible localement, documenter le blocage dans `PLAN.md` et faire au minimum une revue statique des ports.
  - Findings: Traefik reste routé vers le Service frontend sur le port 80, le Service cible toujours le pod sur 8080, et la CiliumNetworkPolicy autorise désormais Traefik vers les pods frontend sur le port 8080. Vérifié par revue statique de `deploy/k8s/ingress.yaml`, `deploy/k8s/frontend-service.yaml`, `deploy/k8s/frontend-deployment.yaml`, `deploy/k8s/networkpolicy.yaml`, et `git diff --check`; `kubectl` n’est pas disponible ici pour la validation server-side.

- [x] Résoudre le blocage `npm audit --audit-level=high` frontend.
  - Tenter d’abord une correction non breaking via `npm audit fix` dans `frontend/` afin de mettre à jour la dépendance transitive `brace-expansion`.
  - Ne pas utiliser `npm audit fix --force` sans décision explicite, car cela peut migrer `react-router-dom` vers v7 et introduire un changement majeur.
  - Si `react-router` moderate reste signalé mais que `npm audit --audit-level=high` passe, documenter la vulnérabilité moderate résiduelle et créer une décision/follow-up séparé pour une migration React Router v7.
  - Validation: `cd frontend && npm audit --audit-level=high`, `npm run lint`, `npm run build`, et `npm run test` si disponible/stable.
  - Findings: `npm audit fix` a remonté `brace-expansion` en `5.0.9`, ce qui fait passer `npm audit --audit-level=high`; il reste deux advisories moderate sur `react-router` / `react-router-dom` liés à une migration v7 explicite. Vérifié avec `npm audit --audit-level=high`, `npm run lint`, `npm run build`, `npm run test`, et `git diff --check`; seul `frontend/package-lock.json` a changé.

- [x] Réduire les permissions GitHub Actions du workflow Docker.
  - Remplacer les permissions workflow-level trop larges par des permissions job-level ou une structure équivalente: PR builds avec `contents: read` seulement; publication GHCR avec `packages: write` uniquement sur `push` main/tags.
  - Conserver les guards existants: pas de `pull_request_target`, pas de secrets sur PR externe, `docker/login-action` et `push: true` uniquement hors `pull_request`.
  - Préserver le check no-shell des images PR.
  - Validation: `git diff --check`; `actionlint` si disponible; revue manuelle des triggers/permissions; GitHub Actions PR run après push.
  - Findings: `.github/workflows/docker.yml` now runs PR image builds under `contents: read` only, while `packages: write` is scoped to dedicated publish jobs that run only on trusted push/tag events. The PR no-shell checks are preserved. Verified by manual workflow review and `git diff --check`; `actionlint` is not installed here.

- [x] Revalider l’ensemble après corrections.
  - `make lint`.
  - `make build`.
  - `make test` si l’environnement dispose du compilateur C requis pour `-race`; sinon documenter le blocage et exécuter `cd backend && go test ./...` plus les tests frontend.
  - `docker compose config`.
  - `docker compose up -d --build postgres backend frontend` puis smoke tests frontend `/`, `/healthz`, `/readyz`, `/api/healthz`, backend `/api/healthz`, backend `/api/readyz`.
  - Vérifier que les images backend/frontend restent sans shell: `docker run --rm --entrypoint /bin/sh <image>` doit échouer.
  - Vérifier `npm audit --audit-level=high`.
  - Findings: `make lint`, `make build`, `cd backend && go test ./...`, `cd frontend && npm run test`, `docker compose config`, and a full `docker compose up -d --build postgres backend frontend` smoke test all passed. HTTP smoke checks covered frontend `/`, `/healthz`, `/readyz`, `/api/healthz` and backend `/api/healthz`, `/api/readyz`. Both runtime images still fail `docker run --rm --entrypoint /bin/sh <image>` as expected. `npm audit --audit-level=high` passes; the Docker build path also remained green after the workflow permission split. `make test` is still blocked in this environment because the `-race` backend target needs a C compiler (`aarch64-linux-gnu-gcc`) that is not installed here.

- [x] Préparer PR et rollout production après validation.
  - Avant PR: inspecter `git status`, `git diff`, et vérifier que seuls les changements attendus sont inclus.
  - PR: inclure le contexte de remediation, les validations, les validations bloquées (`kubectl`, `make test -race` si toujours bloqués), et le rollback.
  - Après merge/release: déployer avec `make deploy VERSION=<new-version>`.
  - Vérifier `make deploy-status`, `make deploy-logs`, frontend public, API proxy, probes frontend/backend, login/registration, et SMTP/TLS si concerné.
  - Findings: `git status -sb` et `git diff --stat` confirment uniquement les changements attendus du hardening (`.github/workflows/docker.yml`, `backend/Dockerfile`, `deploy/k8s/frontend-deployment.yaml`, `deploy/k8s/frontend-service.yaml`, `deploy/k8s/networkpolicy.yaml`, `docker-compose.yml`, `frontend/Dockerfile`, `frontend/package-lock.json`, plus `frontend/.dockerignore` et `frontend/server.go`). `git diff --check` est propre. La PR est prête à être ouverte/mergée; le déploiement `make deploy VERSION=<new-version>` reste une opération post-merge.

## Validation

- Git/workflow:
  - `git status -sb`.
  - `git diff --check`.
  - `actionlint` si installé.
  - Revue manuelle: aucun `pull_request_target`, pas de secrets dans PR workflow, permissions PR minimales, push GHCR uniquement sur événements de confiance.
- Frontend dependencies:
  - `cd frontend && npm audit --audit-level=high`.
  - `cd frontend && npm run lint`.
  - `cd frontend && npm run build`.
  - `cd frontend && npm run test`.
- Kubernetes/Cilium:
  - Revue statique des ports frontend: IngressRoute service port 80, Service `targetPort: 8080`, Deployment `containerPort: 8080`, NetworkPolicy ingress pod port `8080`.
  - `kubectl apply --dry-run=server -f deploy/k8s/`.
  - `kubectl diff -f deploy/k8s/`.
- Docker/Compose hardening regression:
  - `docker compose config`.
  - `docker compose up -d --build postgres backend frontend`.
  - Smoke tests HTTP locaux.
  - Runtime shell absence checks for backend/frontend.
- Broader build/test:
  - `make lint`.
  - `make build`.
  - `make test` where environment supports Go race detector C compiler; fallback: `cd backend && go test ./...` plus frontend test command, with blocker documented.

## Risks

- NetworkPolicy port correction is small but production-critical; an incorrect port can cause frontend outage behind Traefik.
- Rebase/branch refresh can conflict with the existing dirty root worktree or newer `origin/main`; use a separate worktree and inspect diffs before continuing.
- `npm audit fix` can modify `package-lock.json` broadly; inspect dependency diff and avoid force updates unless explicitly accepted.
- React Router v7 migration may require code changes and routing behavior validation; do not hide it inside an infra-only remediation if avoidable.
- Tightening GitHub Actions permissions can break GHCR publishing if `packages: write` is not granted on trusted push/tag jobs.
- `kubectl` validation remains environment-dependent; static YAML review is not a substitute for server-side validation before production rollout.

## Rollback

- NetworkPolicy rollback: revert the single `deploy/k8s/networkpolicy.yaml` port change to the previous value if the frontend rollout fails, then reapply previous manifests from the last known-good release.
- Dependency rollback: revert `frontend/package.json` / `frontend/package-lock.json` changes and redeploy previous frontend image if runtime behavior regresses.
- GitHub Actions rollback: revert `.github/workflows/docker.yml` permission changes if image publishing fails, while preserving no-shell checks if possible.
- Branch/plan rollback: abandon the refreshed worktree/branch and recreate from `origin/main` if rebase/reapply becomes confusing.
- Production rollback: redeploy previous known-good image tag with `make deploy VERSION=<previous-version>` and monitor with `make deploy-status` / `make deploy-logs`.
- No database rollback expected; no schema changes planned.

## Notes / Decisions

- Decision: fix the CiliumNetworkPolicy port mismatch before any rollout; this is the highest production-impact finding.
- Decision: preserve frontend Service port 80 and Traefik IngressRoute service port 80; only the pod-facing target/network policy should move to 8080.
- Decision: first try non-breaking dependency remediation for the high-severity audit blocker; React Router v7 is treated as separate unless required to pass `npm audit --audit-level=high`.
- Decision: reduce Docker workflow permissions without changing publication event model.
- Assumption: the hardening branch should ultimately contain the remediation plan in `PLAN.md` so reviewers can compare diff against plan in one PR.
- Assumption: `kubectl` server-side validation will be run from a cluster-capable environment before production deployment.
- Unresolved: whether the moderate React Router advisories must be fixed in this same PR or accepted temporarily with a follow-up migration plan.
- First `/next` task for `platform-build`: refresh or recreate the clean hardening worktree/branch from latest `origin/main`, preserve/reapply the existing hardening diff, and copy this remediation `PLAN.md` into that branch; verify with `git status -sb` and `git diff --check`; do not change application, Docker, Kubernetes, dependency, or workflow files yet.
