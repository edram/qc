---
name: git-commit-convention
description: Guides planning git commits and writing or reviewing conventional commit messages. Use when asked to commit changes, decide whether or how to split changes into commits, generate a commit message, or review commit history and style.
metadata:
  author: edram
  version: 2026.09.20
  source: https://www.conventionalcommits.org/en/v1.0.0/
---

Conventional Commits 1.0.0 defines a human-readable and machine-parseable commit message format. Before writing a message, plan the commit boundary from the complete change set; a multi-round conversation may produce several coherent commits rather than one catch-all commit. In the spec, `feat` implies MINOR, `fix` implies PATCH, and breaking changes imply MAJOR; other type-to-version mappings are project conventions.

- Only `type` and `description` are required; everything else is optional
- Scope narrows the affected area: `feat(auth):`, `fix(api):`
- Breaking changes use `!` suffix or `BREAKING CHANGE:` footer (or both)
- Footer tokens follow Git trailer convention
- Default to English, but match the project's existing convention — check `git log` and follow whatever language recent commits use
- Decide commit boundaries from purpose and dependency, not from the number of conversation rounds
- Validate each message against the staged diff: every staged change must belong to the stated purpose, and every material claim must be supported by the diff
- When asked to commit, state the proposed commit plan before creating any commit; then validate each staged commit separately

```
<type>[optional scope][!]: <description>

[optional body]

[optional footer(s)]
```

## Core

| Topic | Description | Reference |
|-------|-------------|-----------|
| Commit Planning | Inspect, group, split, and validate changes before committing | [core-planning](references/core-planning.md) |
| Format | Full grammar, header/body/footer rules, length limits | [core-format](references/core-format.md) |
| Types | Common types, project-level SemVer impact, and when to use each | [core-types](references/core-types.md) |

## Features

| Topic | Description | Reference |
|-------|-------------|-----------|
| Breaking Changes | `!` syntax and `BREAKING CHANGE` footer — when and how | [features-breaking](references/features-breaking.md) |
| Footers | Token format rules, issue references, common tokens | [features-footers](references/features-footers.md) |
| AI-Generated Changes | Required `Co-Authored-By` attribution whenever an agent contributes to the commit | [features-ai-attribution](references/features-ai-attribution.md) |
