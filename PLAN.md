# PLAN.md

## Objective

Améliorer la sécurité de la PR #54 suite aux retours GitHub Advanced Security et CI : supprimer les alertes CodeQL “Uncontrolled data used in path expression” dans `frontend/server.go`, corriger la vulnérabilité high `nanoid <3.3.17` remontée par `npm audit --audit-level=high`, puis revalider et mettre à jour la PR sans élargir le périmètre du hardening.

## Constraints

- Protected `main`: feature branch + PR only; no direct pushes to `main`.
- Branche de travail: `chore/harden-runtime-containers`, PR #54 ouverte contre `main`; ne pas pousser sur `main`.
- Pendant cette étape de planification, ne modifier que `PLAN.md`.
- Garder les corrections minimales et liées aux deux retours sécurité confirmés.
- Ne pas réintroduire Nginx ni changer l’architecture du hardening sans décision explicite; l’objectif reste une image frontend runtime sans shell.
- Ne pas introduire de secrets, tokens, logs sensibles, ou accès à `secrets.*` dans les workflows PR.
- Éviter toute migration majeure React Router dans ce correctif; les advisories moderate React Router restent un follow-up séparé sauf si elles deviennent high/critical ou disposent d’un patch non breaking.
- Ne pas modifier les manifestes Kubernetes/Compose/Docker sauf si requis pour faire passer les tests; si infra change, inclure validation et rollback.
- Pour GitHub Actions, ne pas utiliser `pull_request_target`; conserver les permissions minimales.

## Current State

- Worktree analysé: `/tmp/opencode/kareelio-runtime-harden-baseline`.
- Branche: `chore/harden-runtime-containers`, tracking `origin/chore/harden-runtime-containers`, propre au moment de l’analyse locale après le commit `fb8badc fix(hardening): align runtime hardening and ci`.
- PR: #54 (`fix(hardening): align runtime hardening and ci`) ouverte contre `main`.
- Checks GitHub observés:
  - `Frontend`: failed.
  - `CodeQL`: failed via review/annotations GitHub Advanced Security.
  - `Analyze (go)`, `Analyze (javascript-typescript)`, `Backend`, `Build Backend`, `Build Frontend`: passed.
- Alertes CodeQL confirmées dans `frontend/server.go`:
  - `frontend/server.go:94`: `os.Stat(fsPath)` dépend d’un chemin dérivé de `r.URL.Path`.
  - `frontend/server.go:115`: `http.ServeFile(..., filePath)` reçoit ce chemin dérivé.
  - Type: `Uncontrolled data used in path expression`.
- Échec CI frontend confirmé dans les logs PR:
  - `npm audit --audit-level=high` signale `nanoid <3.3.17` high severity (`GHSA-2v37-7h3g-55p8`).
  - Le lockfile courant contient `nanoid@3.3.16` sous `frontend/package-lock.json`.
  - `nanoid` est transitive via `postcss@8.5.23` (`postcss` déclare `nanoid: ^3.3.16`).
- État local après relecture:
  - `npm audit --json` confirme `nanoid` high, plus les advisories moderate React Router déjà documentées.
  - `npm ls nanoid` montre `postcss@8.5.23 -> nanoid@3.3.16`.
- Les changements de hardening précédents restent attendus et validés: backend/frontend runtime `scratch`, frontend Go static/proxy server, Compose hardening, K8s frontend port 8080 + CiliumNetworkPolicy, workflow Docker permissions réduites.

## Tasks

- [x] Corriger les alertes CodeQL de path traversal dans `frontend/server.go`.
  - Modifier uniquement `frontend/server.go` sauf nécessité de tests/formatage.
  - Supprimer le flux direct `r.URL.Path -> filepath.Join(distDir, ...) -> os.Stat/http.ServeFile`.
  - Implémenter une résolution sûre des assets statiques, par exemple:
    - traiter `/` et les routes SPA sans extension via un chemin constant prévalidé `index.html`;
    - rejeter explicitement les chemins contenant un segment caché (`/.env`, `/.git`, etc.) afin de conserver l’intention de l’ancien bloc Nginx `location ~ /\.`;
    - servir uniquement les fichiers asset connus par extension via un mécanisme qui contraint la lecture à `distDir` (`http.FileServer(http.Dir(distDir))`, `fs.ValidPath`/`http.FS(os.DirFS(distDir))`, ou équivalent sûr);
    - éviter `http.ServeFile` avec un chemin assemblé depuis une entrée utilisateur.
  - Préserver le comportement existant: SPA fallback, `/healthz`, `/readyz`, `/api/` reverse proxy, cache immutable pour assets, `no-store` pour `index.html`, headers sécurité.
  - Validation ciblée:
    - `docker build -t kareelio-frontend-hardened ./frontend`.
    - Démarrer le frontend avec backend/compose ou image locale et vérifier `/`, `/healthz`, `/readyz`, `/api/healthz`, un asset `/assets/...`, et un chemin caché type `/.env` qui doit répondre 404.
    - `git diff --check`.
  - Validation GitHub attendue après push: l’annotation CodeQL PR #54 doit disparaître.
  - Findings: `frontend/server.go` ne transmet plus un chemin dérivé de `r.URL.Path` à `os.Stat`/`http.ServeFile`; les assets sont servis via `http.FileServer(http.Dir(distDir))`, les chemins cachés sont rejetés explicitement, et `index.html` reste servi depuis un chemin constant. Vérifié avec `docker build -t kareelio-frontend-hardened ./frontend`, `docker compose up -d --build postgres backend frontend`, smoke tests sur `/`, `/healthz`, `/readyz`, `/api/healthz`, `/assets/index-DCDKC_dC.js`, `/.env`, `/.git/config`, puis `docker compose down`; `git diff --check` est propre.

- [x] Corriger la vulnérabilité high `nanoid <3.3.17` sans migration majeure.
  - Dans `frontend/`, exécuter une correction non breaking (`npm audit fix` ou mise à jour lockfile équivalente) pour remonter `nanoid` à `3.3.17+`.
  - Ne pas utiliser `npm audit fix --force` sans décision explicite; ne pas migrer React Router v7 dans cette tâche.
  - Inspecter le diff de `frontend/package-lock.json` pour vérifier que le changement reste limité à `nanoid`/lockfile et ne modifie pas les dépendances runtime majeures.
  - Validation:
    - `cd frontend && npm audit --audit-level=high`.
    - `cd frontend && npm run lint`.
    - `cd frontend && npm run build`.
    - `cd frontend && npm run test`.
  - Documenter les advisories moderate React Router restantes comme follow-up si elles persistent.
  - Findings: `npm audit fix` a remonté `nanoid` en `3.3.18` dans `frontend/package-lock.json` sans migration React Router. Vérifié avec `cd frontend && npm audit --audit-level=high` (qui passe toujours), `npm run lint`, `npm run build`, `npm run test`, et `git diff --check`; le diff lockfile reste limité à `nanoid`.

- [x] Revalider la PR #54 après les corrections sécurité.
  - `git diff --check`.
  - `make lint`.
  - `make build`.
  - `make test` si l’environnement dispose du compilateur C requis pour `go test -race`; sinon documenter le blocage et exécuter `cd backend && go test ./...` plus `cd frontend && npm run test`.
  - `docker compose config`.
  - `docker compose up -d --build postgres backend frontend`, puis smoke tests:
    - frontend `/`, `/healthz`, `/readyz`;
    - frontend proxy `/api/healthz`;
    - backend `/api/healthz`, `/api/readyz`;
    - asset statique sous `/assets/...`;
    - chemin caché `/.env` ou `/.git/config` attendu en 404.
  - Vérifier que les images backend/frontend restent sans shell: `docker run --rm --entrypoint /bin/sh <image>` doit échouer.
  - Nettoyer les containers de validation avec `docker compose down`.
  - Findings: `git diff --check`, `make lint`, `make build`, `cd backend && go test ./...`, `cd frontend && npm run test`, `docker compose config`, `docker compose up -d --build postgres backend frontend`, smoke tests frontend `/`, `/healthz`, `/readyz`, `/api/healthz`, `/assets/index-DCDKC_dC.js`, `/.env`, `/.git/config`, backend `/api/healthz`, `/api/readyz`, and both shell-absence checks all passed. `make test` remains blocked in this environment because `go test -race` needs `CGO_ENABLED=1` and a C compiler (`aarch64-linux-gnu-gcc`) that are not installed here. `npm audit --audit-level=high` still passes; the remaining frontend advisories are the documented moderate React Router follow-up.

- [ ] Mettre à jour la PR #54 après validation.
  - Committer les corrections en Conventional Commit, par exemple `fix(frontend): constrain static file serving` ou `fix(hardening): address frontend security checks` selon le diff final.
  - Pousser uniquement la branche `chore/harden-runtime-containers`.
  - Vérifier les checks PR #54:
    - `Frontend` doit passer;
    - `CodeQL` / GitHub Advanced Security ne doit plus afficher les deux alertes path expression;
    - les jobs Docker PR doivent rester verts.
  - Mettre à jour la description/commentaire PR avec les validations et les éventuels blocages locaux (`kubectl`, `make test -race`) si nécessaire.

## Validation

- CodeQL/path traversal:
  - Revue du diff `frontend/server.go` pour confirmer qu’aucun chemin fichier n’est construit naïvement depuis `r.URL.Path`.
  - Validation PR GitHub Advanced Security après push.
- Frontend dependency/security:
  - `cd frontend && npm audit --audit-level=high`.
  - `cd frontend && npm run lint`.
  - `cd frontend && npm run build`.
  - `cd frontend && npm run test`.
- Runtime/Compose regression:
  - `docker build -t kareelio-frontend-hardened ./frontend`.
  - `docker compose config`.
  - `docker compose up -d --build postgres backend frontend` + smoke tests HTTP.
  - No-shell checks backend/frontend.
- Broader validation:
  - `make lint`.
  - `make build`.
  - `make test` where supported; fallback documented if `go test -race` lacks a C compiler.
- GitHub Actions:
  - PR #54 checks after push (`gh pr checks 54`).
  - Revue: pas de `pull_request_target`, pas de secrets sur PR, publish GHCR seulement sur événements de confiance.

## Risks

- Une correction trop permissive du serveur statique peut laisser une traversal (`..`, chemins absolus, symlinks, fichiers cachés) ou casser le fallback SPA.
- Remplacer trop largement la logique static/proxy peut introduire une régression de cache headers, CSP/security headers, health probes, ou proxy `/api/`.
- `npm audit fix` peut modifier plus que `nanoid`; le diff lockfile doit être inspecté avant commit.
- Les advisories React Router moderate restent un risque résiduel si elles sont acceptées temporairement; ne pas les masquer dans ce correctif.
- CodeQL peut encore alerter si le code conserve un flux utilisateur vers `os.Stat`, `os.Open` ou `http.ServeFile` même après sanitation apparente.
- Les validations `kubectl` restent non disponibles localement; ne pas déployer en production sans validation cluster-side.

## Rollback

- Serveur frontend: revert du commit de correction `frontend/server.go` pour revenir au comportement précédent de la PR #54 si le frontend/proxy régresse.
- Dépendances frontend: revert de `frontend/package-lock.json` si le lockfile cause une régression inattendue; cela réouvrira le blocage audit high et devra être remplacé par une autre correction.
- Runtime hardening global: revert de la PR #54 complète et redeploy du dernier tag connu bon via `make deploy VERSION=<previous-version>` si un problème production est découvert après merge.
- Docker/Compose validation: `docker compose down` pour nettoyer l’environnement local après tests.
- Aucun rollback DB prévu; aucune migration ou changement de données n’est dans le périmètre.

## Notes / Decisions

- Decision: corriger d’abord CodeQL dans `frontend/server.go` car il s’agit d’un commentaire bloquant GitHub Advanced Security sur du code nouvellement ajouté.
- Decision: corriger `nanoid` via lockfile/audit non breaking; ne pas utiliser `npm audit fix --force`.
- Decision: maintenir React Router v7 comme follow-up séparé tant que les advisories restent moderate et que le fix est breaking.
- Assumption: le serveur frontend doit continuer à tourner en image `scratch` sans shell; la solution ne doit pas revenir à Nginx dans ce correctif.
- Assumption: les assets Vite restent servis sous `/assets/` avec extensions connues; les routes sans extension doivent fallback vers `index.html`.
- Unresolved: CodeQL peut nécessiter une validation GitHub après push pour confirmer que les alertes path expression sont closes.
- First `/next` task for `platform-build`: update only `frontend/server.go` to eliminate the CodeQL path-expression flow by constraining static file serving to `distDir` without passing user-derived filesystem paths to `os.Stat`/`http.ServeFile`; preserve SPA fallback, `/api/` proxy, health/readiness endpoints, security/cache headers, and hidden-file denial; verify with `docker build -t kareelio-frontend-hardened ./frontend`, local frontend/compose smoke tests including `/.env` returning 404, and `git diff --check`; do not change dependencies or workflows in this step.
