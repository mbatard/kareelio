---
description: Implement the next unchecked task from PLAN.md and update the plan.
agent: platform-build
model: openai/gpt-5-mini
---

Read `PLAN.md`, identify the next unchecked task, implement only that task, verify it, then update `PLAN.md`.

Additional user context:

$ARGUMENTS

Rules:
- Keep changes minimal.
- Do not skip validation unless blocked; if blocked, document why in `PLAN.md`.
- Do not start a second task unless the first task requires it to compile or pass tests.
