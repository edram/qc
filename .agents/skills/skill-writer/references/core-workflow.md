---
name: core-workflow
description: Step-by-step process for creating a new skill or modifying an existing one
---

## Usage

Copy this checklist and track progress:

```
Skill Progress:
- [ ] Step 1: Resolve target (new vs edit; inspect prior art)
- [ ] Step 2: Capture intent (what / when / expected output)
- [ ] Step 3: Research & synthesize sources
- [ ] Step 4: Author SKILL.md router, then references
- [ ] Step 5: Review against anti-patterns + bump version
```

**Step 1 — Resolve target.** Decide create vs modify. Read a sibling skill first
(`skills/git-commit-convention/SKILL.md` and one of its references) to lock onto the
house style. For an edit, locate the *specific* reference file to change rather than
rewriting the SKILL.md.

**Step 2 — Capture intent.** Pin down three things before writing:
- *What* the skill does (the capability).
- *When* it should trigger (the phrases/contexts/file types) → this becomes the `description`.
- *Expected output* (what a successful invocation produces).

**Step 3 — Research & synthesize.** Pull source material (docs, examples), then keep only
what an agent must *do* and wouldn't already know. Discard intros, install guides, and
background an LLM is already confident about. Rewrite for action, don't copy prose.

**Step 4 — Author.** Use `doc-writer` to set the reader, voice, and evidence standard, then
apply this skill's repository format. Write the SKILL.md router first: frontmatter → overview →
principle bullets → Core/Features tables. Then write each reference, **one concept per file**,
code-first under `## Usage`, gotchas under `## Key Points`, source comment at the bottom.

**Step 5 — Review.** Run the [features-anti-patterns](features-anti-patterns.md) checklist,
sanity-check triggering per [features-evaluation](features-evaluation.md), and bump
`metadata.version` (use today's date).

## Key Points

- Router-first ordering prevents the common failure of stuffing everything into SKILL.md — you decide what's a reference *before* writing prose.
- `doc-writer` improves the content; it does not override this skill's frontmatter, layout, naming, or progressive-disclosure rules.
- For edits, change the narrowest file that fixes the problem and bump the version; avoid touching unrelated references.
- Don't over-split. Prefer a few dense references over many thin ones (see [core-structure](core-structure.md)).

<!--
Source references:
- https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices
- https://github.com/anthropics/skills/blob/main/skills/skill-creator/SKILL.md
-->
