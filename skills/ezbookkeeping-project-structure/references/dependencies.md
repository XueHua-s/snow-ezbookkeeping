# ezbookkeeping Dependency and Tooling Notes

## Backend

- Language/runtime: Go.
- DB/ORM stack includes `xorm` via internal packages.
- Config parser: `gopkg.in/ini.v1`.
- HTTP routing uses Gin through `cmd/` and `pkg/core/`; outbound provider requests reuse `pkg/httpclient/`.

## Frontend

- Framework: Vue 3 + TypeScript.
- Desktop UI: Vuetify.
- Mobile UI: Framework7.
- Build tool: Vite; use npm and the checked-in lockfile. `vite.config.mts`, `vitest.config.mts` and `vitest.browser.config.mts` are explicit ESM configs; keep the package’s CommonJS default.
- State: Pinia.
- i18n: vue-i18n.

## Quality Commands

- Frontend type/lint: `npm run lint` (source-line check, vue-tsc, then ESLint with `--fix`; this command can modify files).
- Frontend unit tests: `npm run test`; browser component tests: `npm run test:browser`.
- Harness `.test.mjs` files import Vitest and run in its Node environment, including environment/cleanup and bounded-command/build-resources regressions. Use `npm run test` or a focused `npx vitest run` with explicit paths; do not use `node --test`. Browser review scripts are separate integration entry points.
- Migration helper checks use Python unittest: `python3 -m unittest discover -s scripts -p assistant_migration_runtime_test.py -v`. They are not part of Vitest. POSIX process-group helpers target macOS/Linux; no Windows process-tree support is implied.
- Backend tests: `go test ./...`.
- Backend static checks: `go vet ./...`.

## Dependency Policy

- Reuse existing provider abstractions in `pkg/llm/provider/*` before adding new transport layers.
- Put provider transport options in `LLMConfig`, runner options in `AIAssistantRuntimeConfig`, and persistence/lifecycle limits in `AIAssistantExperienceConfig`; use settings loaders instead of ad-hoc feature env reads.
- Keep general frontend server flags in `src/lib/server_settings.ts`; persistent assistant capabilities/limits come through `src/lib/services/assistant_conversations.ts` into the shared store. Reuse the existing Axios transport, Pinia, Vue composables and Markdown renderer; consult package.json/go.mod for current versions rather than introducing parallel stacks.
