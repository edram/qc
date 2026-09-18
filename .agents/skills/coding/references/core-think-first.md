---
name: core-think-first
description: Inspect existing code, establish behavioral contracts, and assess implementation choices before editing
---

## Usage

Before editing, build enough understanding to choose a focused implementation. Scale the analysis to the change; a local correction does not require a system-wide design exercise.

1. **Inspect the existing path.** Read applicable project instructions, the affected implementation, its callers, and relevant tests. Use repository evidence to resolve assumptions before designing new behavior.
2. **Establish the contract.** Identify intended inputs, outputs, side effects, failure behavior, and compatibility constraints. Distinguish requested behavior changes from behavior that must remain stable.
3. **Trace the impact.** Identify responsibility owners, dependencies, and affected consumers. Check for an existing capability before introducing another implementation.
4. **Choose the smallest coherent approach.** Compare alternatives when they materially affect correctness, coupling, or complexity. Follow established boundaries and avoid designing for hypothetical requirements.
5. **Identify verification.** Translate the contract into observable checks using [core-verify](core-verify.md).

## Key Points

- Existing code and tests are evidence of current behavior, not proof that the behavior is correct; compare them with the requested contract.
- Treat unresolved assumptions about behavior as uncertainties, not implementation facts.
- Analyze only the affected area and the dependencies needed to understand it.

<!--
Source references:
- https://github.com/multica-ai/andrej-karpathy-skills/blob/main/skills/karpathy-guidelines/SKILL.md
-->
