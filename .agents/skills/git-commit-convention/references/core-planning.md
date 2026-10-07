---
name: core-planning
description: Plan commit boundaries and verify that each commit message matches its staged changes
---

## Usage

Before committing, inspect the whole change set, not only the last turn's edits.
Include staged and unstaged tracked changes, and account for untracked files:

```bash
git status --short
git diff HEAD --stat
git diff HEAD --name-status
git diff HEAD
git ls-files --others --exclude-standard
git log -5 --oneline
```

Use recent history to infer the project's local convention: language,
capitalization, type vocabulary, scope names, header length, and footer style.
When the local convention differs from the generic examples in this skill,
follow the project unless the user asks for a migration.

Group changes by a single coherent purpose. Keep changes together when they are
required for the same behavior and should be reviewed, reverted, or released as
one unit. Split them when they have different purposes, can be understood or
reverted independently, or mix implementation with unrelated cleanup, docs, or
formatting. Inspect the contents of relevant untracked files before including
them in a plan; `git diff HEAD` does not show them.

Use these questions to decide whether a boundary is real:

- Does each proposed commit have one purpose that can be summarized in one
  subject?
- Could a reviewer understand, revert, or release one commit independently?
- Does the ordering reflect dependencies between commits?
- Is the split useful, or does it merely separate implementation from the test
  that verifies the same behavior?

Use the smallest number of commits that preserves those boundaries. A useful
plan states the commit order and the intent of each commit, for example:

```text
1. fix(parser): handle empty input
   Includes the implementation and its regression test.
2. docs(parser): document the accepted input
```

If the complete change set has one coherent purpose, plan one commit. Do not
split tests from their implementation merely because their files differ; keep
them together when they verify the same behavior and the project prefers
atomic commits.

Stage one planned commit at a time, using path selection or interactive staging
when a file contains changes for multiple commits. Before each commit, inspect
the staged view:

```bash
git diff --cached --check
git diff --cached --stat
git diff --cached --name-status
git diff --cached
```

Derive the message from that staged diff and check both directions:

- All staged files and hunks belong to the message's single purpose; if not,
  split them out or rewrite the commit plan. The message does not need to list
  every file or hunk individually.
- The type, scope, subject, body, and footer make no material claim that the
  staged diff cannot support. If a claim is unsupported, stage the missing
  change, narrow the message, or move the change to another commit.

After staging, confirm that no intended change was left behind and no unrelated
change was included. Use `git status --short` because remaining untracked files
are not shown by `git diff`:

```bash
git diff                 # remaining unstaged changes
git diff --cached        # exactly what this commit will contain
git status --short
```

Run the project's focused checks when the staged boundary supports a meaningful
check. If an intermediate commit intentionally depends on an earlier commit,
validate the ordered series or the final result instead of claiming that each
partial commit was independently tested. Commit the validated unit, then repeat
the staged-diff review for the next planned commit.

## Key Points

- Conversation turns are not commit boundaries. Several rounds may refine one
  coherent change, or one round may contain several independent changes.
- Do not use a broad message such as `chore: update project` to conceal mixed
  feature, fix, test, and documentation changes.
- A clean working tree is not required before planning; preserve unrelated user
  changes and exclude them from the staged commit.
- Prefer an explicit plan over automatically squashing all work into one commit,
  but do not create extra commits without a review, revert, release, or
  dependency benefit.

<!--
Source references:
- https://git-scm.com/docs/git-diff
- https://git-scm.com/docs/git-add
-->
