---
name: features-writing-style
description: Conciseness, degrees of freedom, terminology, and reusable content patterns
---

## Usage

**Be concise — assume the model is smart.** Challenge every sentence: "Does this justify
its token cost?" Cut anything an LLM already knows.

```markdown
# Good (~50 tokens) — assumes the agent knows what a PDF is
Use pdfplumber for text extraction:
import pdfplumber
with pdfplumber.open("file.pdf") as pdf:
    text = pdf.pages[0].extract_text()

# Bad (~150 tokens) — explains PDFs, libraries, and pip from scratch
```

**Match degrees of freedom to the task.** Think of the agent on a path:

| Freedom | Use when | Form |
|---------|----------|------|
| High | many valid approaches; context decides | prose steps, heuristics |
| Medium | a preferred pattern with variation | templates, pseudocode with params |
| Low | fragile, must be exact | one exact command, "do not modify" |

```markdown
# High — open field, trust the agent
1. Analyze the code structure
2. Check for edge cases
3. Suggest readability improvements

# Low — narrow bridge, cliffs both sides
Run exactly: python scripts/migrate.py --verify --backup
Do not add flags.
```

**Use consistent terminology.** Pick one term and keep it: always "reference file", not a
mix of "ref / doc / page". Inconsistency makes instructions ambiguous.

**Lean on patterns over prose** — templates for output shape, input→output example pairs
when quality depends on seeing examples, and checklists for multi-step workflows.

## Key Points

- Trust model capability (theory of mind). Prefer plain guidance over walls of MUST/NEVER; reserve emphatic language for genuinely fragile, must-not-skip steps.
- Give one default with an escape hatch, not a menu of options (see [features-anti-patterns](features-anti-patterns.md)).
- Concrete beats abstract: a real example teaches the desired style faster than a description of it.

<!--
Source references:
- https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices
- https://github.com/anthropics/skills/blob/main/skills/skill-creator/SKILL.md
-->
