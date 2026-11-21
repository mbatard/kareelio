# Plan — audit npm et vérifier Nginx

## Objectif

Auditer les dépendances npm du dépôt, appliquer les mises à jour nécessaires sans changement majeur forcé, et vérifier les références Nginx dans les Dockerfile.

## Tâches

- [x] Exécuter `npm audit` sur les manifestes npm suivis et examiner les vulnérabilités.
- [x] Appliquer les mises à jour npm compatibles et mettre à jour les lockfiles.
- [x] Vérifier les Dockerfile : si une image Nginx y est utilisée, la passer à la dernière version disponible.
- [x] Relancer les audits et validations pertinentes; examiner README, CHANGELOG et AGENTS pour des mises à jour requises.

## Risques / suivi

- L’audit racine signale encore 36 vulnérabilités modérées dans le `npm` embarqué par `@semantic-release/npm` (dont `glob` et `minimatch`). L’audit production (`npm audit --omit=dev`) est propre. `npm audit fix` ne sait pas remplacer ces dépendances embarquées; `--force` propose de rétrograder `semantic-release` 25 vers 21.1.2, donc n’a pas été appliqué.
- Aucun Dockerfile ne référence une image Nginx; aucune modification Docker n’est nécessaire.
- Aucun changement README/CHANGELOG/AGENTS n’est requis pour ces mises à jour internes de dépendances.
