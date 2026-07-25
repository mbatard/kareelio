---
description: Push the current feature branch and create/update a GitHub PR using the repository PR template.
agent: platform-commit
model: openai/gpt-5-mini
---

Create or update a PR for the current branch.

User context:

$ARGUMENTS

Rules:
- Confirm the branch is not `main`.
- Inspect status, diff, recent commits, and remote tracking.
- Push the current branch.
- Create a PR against `main` using the existing template style.
- If `gh` is not authenticated but GitHub API works through sbx, use the GitHub API without exposing tokens.
