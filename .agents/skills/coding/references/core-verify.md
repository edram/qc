---
name: core-verify
description: Verify changed behavior with proportionate checks, attribute failures, and stop when the relevant criteria pass
---

## Usage

Define observable success criteria before implementation, then check the resulting behavior against them.

1. **Derive checks from the contract.** Cover the intended behavior, affected failure paths, and compatibility requirements. If expected behavior is unclear, return to [core-think-first](core-think-first.md).
2. **Select proportionate verification.** Reuse project tooling and relevant tests. Choose checks that can detect an incorrect implementation; add regression coverage when it protects meaningful behavior, not merely to mirror the code. Complete required project checks.
3. **Attribute failures before fixing them.** Distinguish regressions introduced by the change from pre-existing failures, environment problems, and unstable checks. Use a baseline or focused reproduction when needed; do not assume a failure is unrelated without evidence.
4. **Correct and recheck the affected behavior.** Fix failures caused by the change and rerun relevant checks. Keep unrelated repairs outside the patch.
5. **Stop when the criteria pass.** Broaden or repeat verification only when new edits, failures, unresolved risks, or required checks justify it.

## Key Points

- Prefer the cheapest check that provides real evidence of correctness; inspection alone does not establish runtime behavior.
- Keep verification evidence tied to the code state checked, including the check performed, its result, and any unverified behavior.
- A blocked, skipped, or unstable check does not count as a pass. Missing tools or dependencies are verification limitations, not proof of a code defect.

<!--
Source references:
- https://github.com/multica-ai/andrej-karpathy-skills/blob/main/skills/karpathy-guidelines/SKILL.md
-->
