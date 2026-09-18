---
name: core-simplicity
description: Avoid speculative complexity while preserving useful abstractions and readability
---

## Usage

Choose the simplest implementation that satisfies the stated behavior without sacrificing clear responsibilities.

- Do not add speculative features, configuration, or flexibility.
- Introduce abstractions for concrete responsibility boundaries or meaningful shared behavior, not hypothetical reuse. Caller count alone is not a deciding factor.
- Do not add handling for states or errors that cannot occur.

## Key Points

- Prefer readable control flow over code-golfed one-liners and nested ternaries.
- Preserve useful abstractions; fewer lines or files are not a reason to collapse responsibility boundaries.

<!--
Source references:
- https://github.com/multica-ai/andrej-karpathy-skills/blob/main/skills/karpathy-guidelines/SKILL.md
- https://github.com/getsentry/skills/blob/main/skills/code-simplifier/SKILL.md
-->
