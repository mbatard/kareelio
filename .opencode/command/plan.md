---
description: Create or update PLAN.md for an incremental Platform/SRE task.
agent: platform-plan
model: openai/gpt-5.5
---

Create or update `PLAN.md` for this objective:

$ARGUMENTS

Requirements:
- Analyze the current repository state first.
- Keep the plan incremental and production-safe.
- Include validation and rollback for infra changes.
- Include exact next task for `/next`.
- Do not modify source code outside `PLAN.md`.
