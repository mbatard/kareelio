---
description: Safely commit intended changes with Conventional Commits after inspecting status/diff/log.
agent: platform-commit
model: openai/gpt-5-mini
---

Prepare a safe commit for the current work.

User intent / preferred commit message:

$ARGUMENTS

Rules:
- Inspect `git status`, `git diff`, and recent `git log` first.
- Stage only intended files.
- Run relevant validation if not already done.
- Use Conventional Commits.
- Do not push unless the user explicitly requested push/PR.
