---
name: coding
description: Constraints for writing and editing code with discipline — favor minimal code, architectural naming, decoupled responsibilities, separation of capability implementation and business orchestration, and purposeful comments. Use when implementing a feature, writing or editing code, refactoring, renaming variables or files, simplifying, adding a function, or reviewing coupling, responsibilities, orchestration, and scope creep.
metadata:
  author: edram
  version: 2026.09.18
---

Constraints for analyzing, implementing, and verifying code, with focused changes and explicit responsibility boundaries. Adapted from Andrej Karpathy's coding guidelines and Sentry's code-simplifier, with additional structure and comment discipline.

- **Think before coding** — inspect the existing implementation, establish behavioral contracts, and assess the affected dependencies.
- **Prioritize clear responsibilities** — satisfy the requested behavior and responsibility boundaries, then choose the smallest complete change. Fewer lines or files do not justify mixing responsibilities.
- **Keep changes focused** — include necessary responsibility separation, but exclude unrelated historical refactoring and speculative features.
- **Name from the architecture** — variable and file names express domain, ownership, and role within the system.
- **Decouple responsibilities** — each file and method owns one responsibility, with explicit inputs, outputs, and dependencies.
- **Separate implementation from orchestration** — capability code implements focused operations; business orchestration combines those operations into a workflow without absorbing their implementation details.
- **Comment deliberately** — important declarations, external sources, and hard-to-understand logic must be documented; explain *why*, constraints, and contracts, never narrate *how*.
- **Verify changed behavior** — define observable criteria, use proportionate checks, and distinguish regressions from unrelated failures.

## Core

Apply Core across coding tasks. Keep it language-independent; add language- or stack-specific guidance and code examples in Features references, loaded only when relevant.

| Topic | Description | Reference |
|-------|-------------|-----------|
| Think First | Existing behavior, contracts, and implementation choices | [core-think-first](references/core-think-first.md) |
| Scope | Change boundaries and responsibility-first tradeoffs | [core-scope](references/core-scope.md) |
| Structure | Naming, responsibilities, dependencies, and orchestration | [core-structure](references/core-structure.md) |
| Simplicity | Necessary abstractions and readable implementation | [core-simplicity](references/core-simplicity.md) |
| Comments | Required context, contracts, and attribution | [core-comments](references/core-comments.md) |
| Verify | Checks, failure attribution, and stopping conditions | [core-verify](references/core-verify.md) |

<!--
Source references:
- https://github.com/multica-ai/andrej-karpathy-skills/blob/main/skills/karpathy-guidelines/SKILL.md
- https://github.com/getsentry/skills/blob/main/skills/code-simplifier/SKILL.md
-->
