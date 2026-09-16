---
name: core-structure
description: Repository layout, SKILL.md and reference file templates, and naming conventions
---

## Usage

Every skill is a directory under the `skills/` subdirectory:

```
<repo root>/
  skills/                        # subdirectory — create skills HERE, not at repo root
    {skill-name}/
      SKILL.md                   # index: metadata + overview + reference tables
      references/
        {category}-{topic}.md    # one concept per file
```

**SKILL.md template:**

```markdown
---
name: {skill-name}
description: {what it does + when to use it — third person, with trigger terms}
metadata:
  author: edram
  version: YYYY.MM.DD
  source: {optional URL if derived from external docs}
---

Brief overview paragraph.

Key principles as a short bullet list.

> Optional version/scope note.

## Core

| Topic | Description | Reference |
|-------|-------------|-----------|
| Topic Name | Short description | [core-topic](references/core-topic.md) |

## Features

| Topic | Description | Reference |
|-------|-------------|-----------|
| Topic Name | Short description | [features-topic](references/features-topic.md) |
```

**`references/*.md` template:**

```markdown
---
name: {concept-name}
description: {one-line summary}
---

## Usage

{Code-first practical examples — the primary content}

## Key Points

- Non-obvious behavior, edge cases, defaults

<!--
Source references:
- {source-url}
-->
```

## Key Points

- File names use `{category}-{topic}.md`. Categories match the SKILL.md table they appear in: `core-` for must-know fundamentals, `features-` for techniques that improve quality. The reference's `name:` frontmatter equals its filename without extension.
- Skill `name` (and directory name) must match: lowercase letters, numbers, hyphens only. Prefer gerund (`writing-documentation`) or noun-phrase (`pdf-processing`) forms; avoid vague names (`helper`, `utils`, `tools`).
- `## Core` and `## Features` are the conventional split; a small skill may use only `## Core`. Keep tables to the columns shown — `Topic | Description | Reference`.
- No manifest, config, or tooling files are required — a conforming SKILL.md + references is a complete skill.

<!--
Source references:
- AGENTS.md (repository house spec)
- https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices
-->
