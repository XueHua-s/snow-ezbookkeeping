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

- Backend HTTP behavior starts from `pkg/api/*`, `cmd/webserver_api_routes.go` and `cmd/webserver.go`.
- Domain/data behavior is usually in `pkg/services/*` and `pkg/models/*`.
- Config authority is in `pkg/settings/` (schema, loaders, defaults and validation); `conf/ezbookkeeping.ini` documents the public options.
- AI provider and model protocol code is in `pkg/llm/provider/*`.
- Shared page behavior belongs in `src/views/base/*`, shared state in `src/stores/*`, and shared rendering in `src/components/common/*`; platform routes/views/components adapt desktop and mobile navigation and UI.
- Client-server wiring lives in `src/lib/services/*`; `src/lib/services.ts` is the existing facade. General server flags use `src/lib/server_settings.ts`; persistent assistant capabilities and effective limits use the authenticated assistant capabilities endpoint and shared assistant store.

## Execution Steps

1. Confirm the feature boundary and data ownership.
2. Trace from API route to service/model before editing.
3. Apply changes at the lowest correct layer; avoid patching symptoms in UI only.
4. Keep platform-specific UI thin by reusing base view logic where possible.
5. Re-run quality gates for both Go and frontend sides when touchpoints cross layers.

## References

- `references/project-structure.md`
- `references/dependencies.md`
