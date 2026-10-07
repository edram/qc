---
name: core-types
description: Common commit types, their project-level SemVer conventions, and usage guidance
---

## Usage

| Type | Typical SemVer impact | Use for |
|------|--------|---------|
| `feat` | MINOR | New feature visible to users or API consumers |
| `fix` | PATCH | Bug fix |
| `perf` | Project convention | Performance improvement with no behavior change |
| `refactor` | — | Code restructuring — no new feature, no bug fix |
| `style` | — | Formatting, whitespace, missing semicolons — zero logic change |
| `test` | — | Adding or correcting tests |
| `docs` | — | Documentation only |
| `build` | — | Build system, packaging, tooling config, or external dependency changes |
| `ci` | — | CI/CD configuration and workflow changes |
| `chore` | — | Maintenance that does not fit a more specific type and has no direct user-facing behavior |
| `revert` | — | Reverts a previous commit |

```
feat(api): add pagination to GET /users

fix(auth): handle expired token on refresh

perf(db): replace N+1 query with single JOIN

refactor(cart): extract price calculation to domain service

style: apply prettier formatting across src/

test(user): add coverage for null email edge case

docs: document rate-limiting behavior in README

build: upgrade webpack to v5

ci: add caching step for node_modules in GitHub Actions

chore: remove unused env variable from .env.example

revert: feat(api): add pagination to GET /users
```

### `revert` specifics

Use the project's established revert format. When following Git's standard
format, the header uses the original commit's full header and the body includes:
```
revert: feat(api): add pagination to GET /users

This reverts commit a1b2c3d. The pagination implementation caused
a regression in the search endpoint response time.
```

Include the reason when it helps explain why the revert is necessary.

## Key Points

- Only `feat`, `fix`, and breaking-change markers have formal SemVer implications in the spec; other type mappings are project conventions
- `refactor` vs `chore`: use `refactor` for structural source-code changes; use `chore` for maintenance without a more specific type or direct user-facing behavior
- `style` vs `refactor`: `style` is purely cosmetic (a formatter ran); `refactor` changes structure
- Adding a breaking change to any type bumps MAJOR regardless of the type's normal SemVer level
- The spec does not define a universal list of types; use the project's existing type vocabulary when it differs from this common set

<!--
Source references:
- https://www.conventionalcommits.org/en/v1.0.0/
- https://github.com/angular/angular/blob/main/contributing-docs/commit-message-guidelines.md
-->
