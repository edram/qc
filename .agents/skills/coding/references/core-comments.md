---
name: core-comments
description: Require comments for important intent, contracts, sources, and opaque logic; remove narration
---

## Usage

Inspect the project's instructions, adjacent code, and tooling first. Match its comment language, documentation syntax, placement, and tags.

Use precise names and types first, then comment on meaning they cannot preserve. A comment is required when removing it would likely cause misuse, a mistaken edit, lost attribution, or force a maintainer to reconstruct important intent. Importance is relative to the declaration's own scope: a local variable can be critical, while a global type can be obvious.

Keep two purposes distinct:

- **Inline comments** preserve intent the code cannot express: *why*, constraints, risks, invariants, and non-obvious consequences.
- **Documentation comments** describe an important declaration's meaning, contract, valid uses, and behavioral variants. This is not line-by-line narration.

### Required comments

Add a comment for:

- **External sources.** When code is copied, ported, adapted, or materially based on a website or repository, state the relationship and include the exact page or a stable commit permalink. Use the project's native documentation form (`@see`, Javadoc, KDoc, JSDoc, or equivalent); otherwise use `Adapted from:`, `Ported from:`, `Based on:`, or `See:`. Preserve any required copyright or license notice—a URL is not a substitute for license compliance.
- **Important variables at any scope.** Document variables that carry key state, business rules, invariants, units or encodings, lifecycle boundaries, snapshots, or concurrency semantics. Record the context that naming and types cannot preserve; do not restate either one.
- **Important types and fields.** Document types that define central domain concepts, ownership, state models, protocols, or usage constraints. Comment an individual field only when that field has independently important semantics; do not comment every field merely because its type is important.
- **Critical or multi-meaning methods.** Document key methods, overloads, distinct operating modes, non-obvious side effects or errors, and required call ordering. Distinguish the contract of each meaning without repeating the implementation.
- **Intrinsically opaque logic.** Document difficult algorithms, regular expressions, compatibility paths, workarounds, and performance, security, or concurrency boundaries. Explain the model, reason, or invariant needed to verify the logic—not each step.

Before finishing, audit the changed code only:

1. Add any required source, declaration, or reasoning comment that is missing.
2. Remove comments that merely translate the next statement or repeat a name or type.
3. Confirm retained comments still match the code, project conventions, and linked source.

## Key Points

- A declaration's visibility or kind does not make it important. Obvious values, plain fields, simple accessors, and straightforward control flow need no comment.
- Good naming does not replace contracts, rationale, or attribution that code cannot express.

<!--
Source references:
- https://github.com/getsentry/skills/blob/main/skills/code-simplifier/SKILL.md
- https://google.github.io/styleguide/docguide/best_practices.html
- https://go.dev/doc/comment
- https://docs.oracle.com/en/java/javase/20/docs/specs/javadoc/doc-comment-spec.html
- https://kotlinlang.org/docs/kotlin-doc.html
-->
