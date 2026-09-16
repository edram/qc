---
name: skill-writer
description: Author and revise agent skills in this repository's format. Use when asked to create a new skill, scaffold a SKILL.md, add or edit reference files, fix a skill's structure, or improve a skill's discovery/description.
metadata:
  author: edram
  version: 2026.09.16
  source: https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices
---

A skill is a progressive-disclosure document set the agent loads on demand. Only the `name` + `description` are pre-loaded into context; the SKILL.md body is read when the skill triggers, and `references/*.md` files are read one at a time only when needed. Authoring a skill means writing for that loading model — not dumping a manual into context.

- **SKILL.md is a router, not an encyclopedia.** It gives a short overview and points to reference files; depth lives in `references/`.
- **Only add what the model doesn't already know.** Every token competes with the conversation. Skip what an LLM is already confident about.
- **`description` drives discovery.** Write it in third person, stating *what the skill does* and *when to use it*, with concrete trigger terms.
- **Match this repo's house format**, not Anthropic's generic `# heading` layout. Model new skills on the existing `git-commit-convention` skill.
- **Use `doc-writer` for the prose.** This skill owns skill structure, metadata, and progressive disclosure; `doc-writer` owns audience, voice, evidence, and Markdown presentation. This skill's house format wins if their guidance overlaps.
- Skills live in `skills/{skill-name}/` — the `skills/` **subdirectory**, never the repo root (the repo itself is also named `skills`).

> This skill is pure markdown — no scripts or tooling, matching repo convention.

## Core

| Topic | Description | Reference |
|-------|-------------|-----------|
| Structure | Directory layout, SKILL.md + reference templates, file/skill naming | [core-structure](references/core-structure.md) |
| Frontmatter | `name` and `description` rules — the discovery-critical metadata | [core-frontmatter](references/core-frontmatter.md) |
| Workflow | End-to-end process for creating and modifying a skill | [core-workflow](references/core-workflow.md) |

## Features

| Topic | Description | Reference |
|-------|-------------|-----------|
| Progressive Disclosure | Router pattern, one-level references, when to split, TOCs | [features-progressive-disclosure](references/features-progressive-disclosure.md) |
| Writing Style | Conciseness, degrees of freedom, terminology, examples/templates | [features-writing-style](references/features-writing-style.md) |
| Anti-Patterns | Common mistakes + condensed pre-ship review checklist | [features-anti-patterns](references/features-anti-patterns.md) |
| Evaluation | Lightweight, script-free way to sanity-check triggering and value | [features-evaluation](references/features-evaluation.md) |
