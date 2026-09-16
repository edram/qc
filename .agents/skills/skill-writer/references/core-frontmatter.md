---
name: core-frontmatter
description: Rules for the name and description fields that drive skill discovery
---

## Usage

The frontmatter is the only part of a skill pre-loaded into every session. `name` and `description` decide whether the skill is found at all, so they get the most care.

```yaml
---
name: skill-writer
description: Author and revise agent skills in this repository's format. Use when asked to create a new skill, scaffold a SKILL.md, add or edit reference files, or improve a skill's description.
metadata:
  author: edram
  version: 2026.09.16
  source: https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices
---
```

**`description` — write what + when.** State the capability, then the triggering contexts and key terms a user would actually say:

```yaml
# Good — what it does AND when, with concrete triggers
description: Extract text and tables from PDF files, fill forms, merge documents. Use when working with PDFs or when the user mentions forms or document extraction.

# Good — third person, names the file types that trigger it
description: Analyze Excel spreadsheets, create pivot tables, generate charts. Use when analyzing spreadsheets, tabular data, or .xlsx files.

# Bad — vague, no triggers, gives nothing to match on
description: Helps with documents
```

**Always third person.** The description is injected into the system prompt; first/second person degrades matching:

```yaml
description: Processes Excel files and generates reports          # good
description: I can help you process Excel files                   # avoid
description: You can use this to process Excel files              # avoid
```

## Key Points

- `name`: ≤64 chars; lowercase letters, numbers, hyphens only; no XML tags; must not contain the reserved words `anthropic` or `claude`. Match the directory name.
- `description`: non-empty, ≤1024 chars, no XML tags. Pack in trigger terms (file types, verbs, synonyms) — this is what disambiguates one skill from 100 others.
- `metadata` (this repo): `author: edram`, `version: YYYY.MM.DD`, optional `source:` URL. **Bump `version` on every edit** to a skill.
- A weak `description` is the most common reason a good skill never fires. Spend effort here, not on a longer body.

<!--
Source references:
- https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices
-->
