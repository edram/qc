---
name: features-anti-patterns
description: Common skill-authoring mistakes and a condensed pre-ship review checklist
---

## Usage

Avoid these recurring mistakes:

- **Windows-style paths.** Always forward slashes, cross-platform: `references/guide.md`, not `references\guide.md`.
- **Too many options.** Give one default with an escape hatch, not a menu.

  ```markdown
  # Bad
  Use pypdf, or pdfplumber, or PyMuPDF, or pdf2image...
  # Good
  Use pdfplumber. For scanned PDFs needing OCR, use pdf2image with pytesseract.
  ```

- **Time-sensitive text.** Don't write "before August 2025, use the old API." Put deprecated material in a collapsed "old patterns" section instead.

  ```markdown
  <details><summary>Legacy v1 API (deprecated)</summary>
  The v1 endpoint was api.example.com/v1 — no longer supported.
  </details>
  ```

- **Deeply nested references.** Every reference links directly from SKILL.md (see [features-progressive-disclosure](features-progressive-disclosure.md)).
- **Unqualified MCP tool names.** Always `ServerName:tool_name`, e.g. `GitHub:create_issue` — bare names fail to resolve.
- **Assuming packages are installed.** State the dependency before using it (`pip install pypdf`, then the code).

**Pre-ship checklist** (condensed for this repo):

```
- [ ] description is third person, specific, names triggers
- [ ] description states both what it does AND when to use it
- [ ] SKILL.md body lean; depth pushed to references
- [ ] references are one level deep from SKILL.md
- [ ] consistent terminology throughout
- [ ] examples are concrete, with code
- [ ] forward slashes everywhere; no time-sensitive text
- [ ] file names {category}-{topic}.md; name fields match filenames
- [ ] metadata.version bumped (today's date)
```

## Key Points

- Most of these are also things *this* skill must itself obey — a skill that breaks its own rules is the loudest anti-pattern.
- When unsure whether content belongs in SKILL.md or a reference: if it's needed to *route*, keep it in SKILL.md; if it's *depth*, move it to a reference.

<!--
Source references:
- https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices
-->
