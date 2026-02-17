---
name: ezbookkeeping-project-spec
description: "Global implementation rules for ezbookkeeping. Use for cross-layer features that touch Go backend, Vue mobile/desktop UI, storage, and migration safety. Enforces Software Design Philosophy style simplification and data-safe delivery."
---

# ezbookkeeping Project Spec Skill

## Quick Start

- Read `references/global-prompt.md` and apply it as the default design posture.
- If file location is unclear, use `ezbookkeeping-project-structure` and read `references/project-structure.md` there.
- Run `rg --files cmd pkg src conf docs skills` before broad edits to confirm current layout.

## Core Rules

- Keep modules deep: simplify call sites first, then hide complexity behind focused interfaces.
- Prefer one clear source of truth for config and constants. Do not duplicate provider URLs, model IDs, or feature flags.
- Keep changes incremental and reversible. For data/storage changes, preserve backward compatibility and migration paths.
- Protect user data first: avoid destructive behavior and document migration implications whenever storage shape or config semantics change.
- Keep mobile and desktop product behavior aligned when adding end-user features.

## Execution Steps

1. Identify the main layer: `pkg/settings`, `pkg/api`, `pkg/services`, `pkg/llm`, or `src/views`.
2. Define the smallest stable interface change that removes the immediate complexity.
3. Implement backend contract first, then wire frontend adapters, then update UI entry points.
4. Update docs/config samples (`conf/ezbookkeeping.ini`, `docs/*`) in the same change.
5. Run quality gates and fix all regressions before finalizing.

## Quality Gates

- `go test ./...`
- `go vet ./...`
- `npm run lint`
- `npm run test` (when frontend logic changed)
- `gofmt -w <changed_go_files>`

## When To Load References

- `references/global-prompt.md`: whenever you need architecture-level tradeoff guidance.
