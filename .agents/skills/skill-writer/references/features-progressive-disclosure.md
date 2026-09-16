---
name: features-progressive-disclosure
description: Organizing a skill so the agent loads only what each task needs
---

## Usage

Skills load in three tiers. Author each tier for its job:

| Tier | Loaded | Contains |
|------|--------|----------|
| Metadata | always | `name` + `description` |
| SKILL.md body | on trigger | overview + routing tables |
| `references/*.md` | on demand | depth, examples, edge cases |

**Keep references one level deep.** Every reference links directly from SKILL.md. Do not
chain references through each other — the agent may only partially read a file reached by a
nested link, getting incomplete information.

```markdown
# Good — flat fan-out from SKILL.md
SKILL.md → references/core-format.md
SKILL.md → references/features-breaking.md

# Bad — nested; the deep file may be read partially
SKILL.md → references/advanced.md → references/details.md → (the real content)
```

**Organize by domain** so unrelated context never loads. A skill covering several areas
splits references by area, and the SKILL.md table tells the agent which one to open:

```
bigquery-skill/
  SKILL.md
  references/
    core-finance.md      # revenue, billing
    core-sales.md        # pipeline
    features-marketing.md
```

**Table of contents for long references.** Any reference over ~100 lines starts with a
contents list, so a partial read still reveals the full scope:

```markdown
## Contents
- Authentication and setup
- Core methods
- Error handling
```

## Key Points

- Keep the SKILL.md body lean (well under ~500 lines); when it grows, move depth into a reference rather than expanding the router.
- Splitting reduces cost: an unopened reference consumes zero context tokens. Bundle comprehensive material freely — it's free until read.
- A reference exists to be loaded *independently*. If two files are always needed together, merge them.

<!--
Source references:
- https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices
-->
