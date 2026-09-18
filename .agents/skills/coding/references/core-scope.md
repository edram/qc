---
name: core-scope
description: Limit edits to the requested behavior and its necessary dependencies while preserving unaffected contracts
---

## Usage

Every changed line should support the requested outcome or a necessary dependency of that outcome.

Within the requested change, clear responsibilities take precedence over minimizing the diff. Choose the smallest complete change that satisfies both behavior and responsibility boundaries, not the fewest changed lines or files.

When editing existing code:

- Include affected callers, contracts, and tests when they must change for the implementation to remain correct. File boundaries alone do not determine scope.
- Include focused extraction when needed to give the changed behavior a cohesive owner. Do not use that extraction as a reason to restructure unrelated legacy code.
- Exclude unrelated refactoring, reformatting, and cleanup, even in adjacent code.
- Match existing naming, error handling, and structure where they remain compatible with the requested change.
- Preserve unaffected behavior. For a pure refactor, preserve observable outputs, side effects, error behavior, and relevant ordering guarantees.

Remove imports, variables, and helpers made unused by the change. Leave pre-existing dead code and unrelated defects outside the patch unless their removal or correction is required by the task.

## Key Points

- When a patch expands, trace each additional edit to a concrete dependency, correctness requirement, or responsibility boundary needed by the changed behavior.
- Review the diff for unrelated changes before verification.

<!--
Source references:
- https://github.com/multica-ai/andrej-karpathy-skills/blob/main/skills/karpathy-guidelines/SKILL.md
- https://github.com/getsentry/skills/blob/main/skills/code-simplifier/SKILL.md
-->
