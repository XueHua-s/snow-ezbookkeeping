---
name: ezbookkeeping-project-structure
description: "Repository map for ezbookkeeping. Use when locating feature entry points, selecting the right layer for changes, or planning cross backend/frontend edits."
---

# ezbookkeeping Project Structure Skill

## Quick Start

- Read `references/project-structure.md` for module boundaries and major entry points.
- Read `references/dependencies.md` before adding dependencies or choosing toolchains.
- Run `rg --files cmd pkg src conf templates docs` when paths are uncertain.

## Navigation Rules

- Backend HTTP behavior starts from `pkg/api/*` and `cmd/webserver.go`.
- Domain/data behavior is usually in `pkg/services/*` and `pkg/models/*`.
- Config authority is in `pkg/settings/setting.go` and `conf/ezbookkeeping.ini`.
- AI provider and model protocol code is in `pkg/llm/provider/*`.
- Shared frontend behavior belongs in `src/views/base/*`; platform rendering lives in `src/views/desktop/*` and `src/views/mobile/*`.
- Client-server wiring is usually in `src/lib/services.ts` and `src/lib/server_settings.ts`.

## Execution Steps

1. Confirm the feature boundary and data ownership.
2. Trace from API route to service/model before editing.
3. Apply changes at the lowest correct layer; avoid patching symptoms in UI only.
4. Keep platform-specific UI thin by reusing base view logic where possible.
5. Re-run quality gates for both Go and frontend sides when touchpoints cross layers.

## References

- `references/project-structure.md`
- `references/dependencies.md`
