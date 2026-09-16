---
name: features-evaluation
description: Lightweight, script-free way to verify a skill triggers and adds value
---

## Usage

No harness or scripts in this repo — evaluation is a quick manual loop you run mentally
or in a scratch session.

**1. Find the gap first.** Before writing, run the target task *without* the skill and note
where the agent fails or lacks context. The skill should close exactly those gaps — not
document imagined ones.

**2. Sanity-check triggering.** List a handful of prompts and confirm the `description`
would (or wouldn't) match:

```
SHOULD trigger:
- "create a skill for parsing CSV files"
- "fix the SKILL.md in skills/foo"
- "make this skill easier to discover"

SHOULD NOT trigger:
- "write a commit message for this diff"   (→ git-commit-convention)
- "rename this variable"
```

If a should-trigger prompt doesn't obviously match the wording, add the missing terms to
the `description`. If a should-not prompt matches, tighten it.

**3. Verify it changes behavior.** Re-run the gap task *with* the skill loaded. The output
should now differ in the way the skill intends. If it doesn't, the relevant guidance is
buried or missing — promote it or make it more prominent.

**4. Iterate the description.** Triggering problems are almost always description problems.
Adjust wording to cut false positives (fires when it shouldn't) and false negatives (fails
to fire), then re-check the prompt list.

## Key Points

- Improvements must generalize across prompts — don't overfit the wording to one test example.
- The `name`/`description` matter most for triggering; the body matters most for behavior. Diagnose which failed before editing.
- Two or three honest trigger prompts catch more real problems than a long imagined list.

<!--
Source references:
- https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices
- https://github.com/anthropics/skills/blob/main/skills/skill-creator/SKILL.md
-->
