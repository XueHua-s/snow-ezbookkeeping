# ezbookkeeping Dependency and Tooling Notes

## Backend

- Language/runtime: Go.
- DB/ORM stack includes `xorm` via internal packages.
- Config parser: `gopkg.in/ini.v1`.
- HTTP stack and routing are under project internal modules.

## Frontend

- Framework: Vue 3 + TypeScript.
- Desktop UI: Vuetify.
- Mobile UI: Framework7.
- Build tool: Vite.
- State: Pinia.
- i18n: vue-i18n.

## Quality Commands

- Frontend type/lint: `npm run lint`.
- Frontend unit tests: `npm run test`.
- Backend tests: `go test ./...`.
- Backend static checks: `go vet ./...`.

## Dependency Policy

- Reuse existing provider abstractions in `pkg/llm/provider/*` before adding new transport layers.
- Prefer adding configuration fields to `LLMConfig` over introducing ad-hoc env reads in feature modules.
- Keep frontend settings access centralized in `src/lib/server_settings.ts`.
